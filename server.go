package main

import (
	"github.com/leoleung0102/go-graphql-starter/config"
	h "github.com/leoleung0102/go-graphql-starter/handler"
	"github.com/leoleung0102/go-graphql-starter/resolver"
	"github.com/leoleung0102/go-graphql-starter/schema"
	"github.com/leoleung0102/go-graphql-starter/service"
	"log"
	"net/http"

	"github.com/leoleung0102/go-graphql-starter/loader"
	graphql "github.com/neelance/graphql-go"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

func main() {
	db, err := config.OpenDB("test.db")
	if err != nil {
		log.Fatalf("Unable to connect to db: %s \n", err)
	}
	viper.SetConfigName("Config")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	var (
		appName             = viper.Get("app-name").(string)
		signedSecret        = viper.Get("auth.jwt-secret").(string)
		expiredTimeInSecond = time.Duration(viper.Get("auth.jwt-expire-in").(int64))
		debugMode           = viper.Get("log.debug-mode").(bool)
		logFormat           = viper.Get("log.log-format").(string)
	)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
	})

	test, err := sess.Config.Credentials.Get()

	log.Println(test)

	// Create an SES session.
	svc := ses.New(sess)

	ctx := context.Background()
	log := h.NewLogger(&appName, debugMode, &logFormat)
	roleService := service.NewRoleService(db, log)
	userService := service.NewUserService(db, roleService, log)
	authService := service.NewAuthService(&appName, &signedSecret, &expiredTimeInSecond, log)
	emailService := service.NewEmailService(svc, log)

	ctx = context.WithValue(ctx, "log", log)
	ctx = context.WithValue(ctx, "roleService", roleService)
	ctx = context.WithValue(ctx, "userService", userService)
	ctx = context.WithValue(ctx, "authService", authService)
	ctx = context.WithValue(ctx,"emailService", emailService)

	graphqlSchema := graphql.MustParseSchema(schema.GetRootSchema(), &resolver.Resolver{})

	http.Handle("/login", h.AddContext(ctx, h.Login()))

	loggerHandler := &h.LoggerHandler{debugMode, log}
	http.Handle("/query", h.AddContext(ctx, loggerHandler.Logging(h.Authenticate(&h.GraphQL{Schema: graphqlSchema, Loaders: loader.NewLoaderCollection()}))))

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graphiql.html")
	}))

	log.Fatal(http.ListenAndServe(":3001", nil))
}
