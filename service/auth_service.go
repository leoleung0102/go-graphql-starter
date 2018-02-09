package service

import (
	"encoding/base64"
	"fmt"
	"github.com/leoleung0102/go-graphql-starter/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/op/go-logging"
	"strconv"
	"time"
	"github.com/satori/go.uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type AuthService struct {
	appName             *string
	signedSecret        *string
	expiredTimeInSecond *time.Duration
	log                 *logging.Logger
	db                  *sqlx.DB
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

func (a *AuthService) GenerateToken(user *model.User) (string, error) {

	token, err := uuid.NewV4()
	if err != nil {
		a.log.Errorf("Error in creating UUID: %v", err)
		return "", err
	}

	tokenID, err := uuid.NewV4()
	if err != nil {
		a.log.Errorf("Error in creating UUID: %v", err)
		return "", err
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token.String()), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return "", err
	}

	encodedToken := base64.StdEncoding.EncodeToString([]byte(token.String()))

	resetPasswordTokenSQL := `INSERT INTO reset_password_token (id,user_id,token, is_expired,is_used)
	VALUES ($1, $2, $3, $4, $5)`

	_, err = a.db.Exec(resetPasswordTokenSQL, tokenID, user.ID, string(hashedToken), false, false)
	if err != nil {
		a.log.Errorf("Error in inserting token: %v", err)
		return "", err
	}

	return encodedToken, err
}

func (a *AuthService) CheckTokenValidation(userEmail string, token string) (string, error) {

	resetPasswordToken := &model.Token{}

	SQL := `SELECT rpt.*
	FROM reset_password_token rpt
	INNER JOIN users u ON rpt.user_id = u.id
	WHERE u.email = ? 
	AND rpt.is_used = 0
	AND rpt.is_expired = 0
	AND DATETIME(rpt.created_at, '+60 minutes') > DATETIME('now')
	ORDER BY rpt.created_at DESC
	LIMIT 1
    `

	row := a.db.QueryRowx(SQL, userEmail)
	err := row.StructScan(resetPasswordToken)

	if err != nil {
		a.log.Errorf("Error in retrieving token : %v", err)
		log.Fatal(err)
	}

	tokenString := string(token)

	err = bcrypt.CompareHashAndPassword([]byte(resetPasswordToken.Token), []byte(tokenString))

	if err != nil {
		a.log.Errorf("Error in validating token - Compare Token : %v", err)
		return "", err
	}else{
		return token, err
	}
}

func (a *AuthService) CheckTokenExpire() {

	tx := a.db.MustBegin()

	tokenSQL := `UPDATE reset_password_token
				 	 SET is_expired = 1
					 WHERE created_at < DATETIME('now', '-60 minutes')`

	tx.MustExec(tokenSQL)

	err := tx.Commit()
	if err != nil {
		a.log.Errorf("Error in updating token status: %v", err)
	}
}
