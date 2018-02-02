package service

import (
	"encoding/base64"
	"fmt"
	"github.com/leoleung0102/go-graphql-starter/model"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/op/go-logging"
	"strconv"
	"time"
	"github.com/satori/go.uuid"
	"github.com/jmoiron/sqlx"
)

type AuthService struct {
	appName             *string
	signedSecret        *string
	expiredTimeInSecond *time.Duration
	log                 *logging.Logger
	db  *sqlx.DB
}

func NewAuthService(appName *string, signedSecret *string, expiredTimeInSecond *time.Duration, log *logging.Logger, db *sqlx.DB) *AuthService {
	return &AuthService{appName, signedSecret, expiredTimeInSecond, log, db}
}

func (a *AuthService) SignJWT(user *model.User) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(user.ID, 10))),
		"created_at": user.CreatedAt,
		"exp":        time.Now().Add(time.Second * *a.expiredTimeInSecond).Unix(),
		"iss":        *a.appName,
	})

	tokenString, err := token.SignedString([]byte(*a.signedSecret))
	return &tokenString, err
}

func (a *AuthService) ValidateJWT(tokenString *string) (*jwt.Token, error) {
	token, err := jwt.Parse(*tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("	unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(*a.signedSecret), nil
	})
	return token, err
}

func (a *AuthService) GenerateToken(userEmail string) (*uuid.UUID, error) {
	token, err := uuid.NewV4()
	if err != nil {
		fmt.Errorf("Something went wrong: %s", err)
		return nil, err
	}

	tokenID, err := uuid.NewV4()
	if err != nil {
		fmt.Errorf("Something went wrong: %s", err)
		return nil, err
	}

	encodedToken := base64.StdEncoding.EncodeToString([]byte(token.String()))

	//decodedToken,_ := base64.StdEncoding.DecodeString(encodedToken)

	user := &model.User{}

	userSQL := `SELECT user.*
	FROM users user
	WHERE user.email = ? `
	row := a.db.QueryRowx(userSQL, userEmail)
	err = row.StructScan(user)
	if err != nil {
		a.log.Errorf("Error in retrieving user : %v", err)
	}

	resetPasswordTokenSQL := `INSERT INTO reset_password_token (id,user_id,token,is_used)
	VALUES ($1, $2, $3, $4)`

	// The token has not yet encoded
	_, err = a.db.Exec(resetPasswordTokenSQL, tokenID, user.ID, encodedToken, false)
	if err != nil {
		return nil, err
	}

	return &token, err
}

func (a *AuthService) ResetPassword(token uuid.UUID) ( error) {
	_, err := uuid.NewV4()
	if err != nil {
		fmt.Errorf("Something went wrong: %s", err)
	}

	return err
}