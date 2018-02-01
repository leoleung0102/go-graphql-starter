package resolver

import (
	"github.com/leoleung0102/go-graphql-starter/config"
	"github.com/leoleung0102/go-graphql-starter/schema"
	"github.com/leoleung0102/go-graphql-starter/service"
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/gqltesting"
	"golang.org/x/net/context"
	"log"
	"testing"
)

var rootSchema = graphql.MustParseSchema(schema.GetRootSchema(), &Resolver{})
var ctx context.Context

func init() {
	db, err := config.OpenDB("../test.db")
	if err != nil {
		log.Fatal(err)
	}
	userService := service.NewUserService(db)
	ctx = context.WithValue(context.Background(), "userService", userService)
}

func TestBasic(t *testing.T) {
	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Context: ctx,
			Schema:  rootSchema,
			Query: `
				{
					user(email:"test@1.com") {
						id
						email
						password
					}
				}
			`,
			ExpectedResult: `
				{
					"user": {
					  "id": "1",
					  "email": "test@1.com",
					  "password": "********"
					}
				}
			`,
		},
	})
}
