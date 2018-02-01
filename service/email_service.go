package service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/op/go-logging"
	"bytes"
	"text/template"
	"github.com/leoleung0102/go-graphql-starter/model"
	"io/ioutil"
)

type EmailService struct {
	SES *ses.SES
	log *logging.Logger
}

var (
	Err error
	doc bytes.Buffer
)

func NewEmailService(ses *ses.SES, log *logging.Logger) *EmailService {
	return &EmailService{ses, log}
}

func (e *EmailService) SendEmail(Sender string, Recipient string, Subject string, TextBody string)(string, error) {

	t := template.New("emailContentTemplate")

	data, _ := ioutil.ReadFile("./service/template.html")

	emailTemplate := model.EmailTemplate{Recipient:Recipient}

	if t, Err =  t.Parse(string(data)); Err != nil {
		e.log.Errorf("error trying to parse mail template %v", Err)
	}

	// Apply the values we have initialized in our struct context to the template.
	if Err = t.Execute(&doc, emailTemplate); Err != nil {
		e.log.Errorf("error trying to execute mail template  %v", Err)
	}

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{
			},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(doc.String()),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
	}

	// Attempt to send the email.
	_, err := e.SES.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				e.log.Errorf(ses.ErrCodeMessageRejected , aerr.Error())
				return aerr.Error(), aerr;
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				e.log.Errorf(ses.ErrCodeMailFromDomainNotVerifiedException , aerr.Error())
				return aerr.Error(), aerr;
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				e.log.Errorf(ses.ErrCodeConfigurationSetDoesNotExistException , aerr.Error())
				return aerr.Error(), aerr;
			default:
				e.log.Errorf("" , aerr.Error())
				return aerr.Error(), aerr;
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			e.log.Errorf("" , err)
			return err.Error(), err
		}
	}else{
		return Recipient , err
	}
}
