package template

import (
 	"text/template"
	"bytes"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
)
var (
 	Err error
 	doc bytes.Buffer
)

type EmailContent struct {
	Recipient string
}

func NewTemplate (ctx context.Context, recipient string) (string, error) {
	// Welcome Email Template
	const emailContentTemplate =
		`
			<h1>Welcome {{.Recipient}}! Thank you for using Good Malling.<h1>
			<p>I hope you enjoy this</p>
			
			<h3>Good Malling</h3>
		`
	t := template.New("emailContentTemplate")
	template := EmailContent{Recipient: recipient}

	if t, Err = t.Parse(emailContentTemplate); Err != nil {
		ctx.Value("logger").(*logging.Logger).Errorf("error trying to parse mail template %v", Err)
	}

	// Apply the values we have initialized in our struct context to the template.
	if Err = t.Execute(&doc, template); Err != nil {
		ctx.Value("logger").(*logging.Logger).Errorf("error trying to execute mail template  %v", Err)
	}

	return doc.String(), Err;
}
