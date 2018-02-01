package resolver

import (
	"github.com/leoleung0102/go-graphql-starter/model"
	"github.com/leoleung0102/go-graphql-starter/service"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	//"strings"
)

func (r *Resolver) CreateUser(ctx context.Context, args *struct {
	Email    string
	Password string
}) (*userResolver, error) {
	user := &model.User{
		Email:     args.Email,
		Password:  args.Password,
		IPAddress: *ctx.Value("requester_ip").(*string),
	}

	user, err := ctx.Value("userService").(*service.UserService).CreateUser(user)
	if err != nil {
		ctx.Value("log").(*logging.Logger).Errorf("Graphql error : %v", err)
		return nil, err
	}
	ctx.Value("log").(*logging.Logger).Debugf("Created user : %v", *user)

	//i := strings.Index(user.Email, "@")
	//nickname := user.Email[0:i]

	recipient, emailErr := ctx.Value("emailService").(*service.EmailService).WelcomeEmail(
		ctx,
		"leoleung@inno-lab.co",
		user.Email,
		"Welcome to Good Malling",
		"This email was sent with Amazon SES using the AWS SDK for Go.",
	)

	if emailErr != nil {
		ctx.Value("logger").(*logging.Logger).Errorf("Welcome email error: %v", emailErr)
		return nil, emailErr
	}
	ctx.Value("logger").(*logging.Logger).Debugf("Welcome Email has sent to: %v", recipient)

	return &userResolver{user}, nil
}
