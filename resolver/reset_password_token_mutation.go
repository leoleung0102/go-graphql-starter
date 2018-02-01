package resolver

import (
	"github.com/leoleung0102/go-graphql-starter/model"
	"github.com/leoleung0102/go-graphql-starter/service"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	//"strings"
)

func (r *Resolver) CreateToken(ctx context.Context, args *struct {
	Email    string
}) (*userResolver, error) {
	user := &model.User{
		Email:     args.Email,
	}

	user, err := ctx.Value("userService").(*service.UserService).ResetPassword(user)
	if err != nil {
		ctx.Value("log").(*logging.Logger).Errorf("Graphql error : %v", err)
		return nil, err
	}
	ctx.Value("log").(*logging.Logger).Debugf("Created user : %v", *user)

	return &userResolver{user}, nil
}