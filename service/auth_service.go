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

	user := &model.User{}

	userSQL := `SELECT user.*
	FROM users user
	WHERE user.email = ? `
	row := a.db.QueryRowx(userSQL, userEmail)
	err = row.StructScan(user)
	if err != nil {
		a.log.Errorf("Error in retrieving user : %v", err)
	}

	resetPasswordTokenSQL := `INSERT INTO reset_password_token (id,user_id,token, is_expired,is_used)
	VALUES ($1, $2, $3, $4, $5)`

	_, err = a.db.Exec(resetPasswordTokenSQL, tokenID, user.ID, encodedToken, false, false)
	if err != nil {
		return nil, err
	}

	return &token, err
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
    `

	rows, err := a.db.Queryx(SQL, userEmail)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.StructScan(&resetPasswordToken)
		if err != nil {
			log.Fatal(err)
		}
		tokenByte, _ := base64.StdEncoding.DecodeString(resetPasswordToken.Token)

		tokenString := string(tokenByte)

		//go a.TokenUpdate(userEmail,resetPasswordToken.Token)

		if token == tokenString {
			return tokenString, err
		}
	}
	return "", err
}

/*func (a *AuthService) TokenUpdate(userEmail string, tokenString string){
	tx, err := a.db.Begin()

	tokenSQL := `UPDATE reset_password_token
				 SET is_expired = 1
				 WHERE user_id = (
					 SELECT id FROM users
					 WHERE email = $1
				 )
				 AND token = $2`

	if err != nil {
		a.log.Errorf("Error in updating token : %v", err)
	}

	_, err = tx.Exec(tokenSQL, userEmail, tokenString)

	if err != nil {
		tx.Rollback()
		a.log.Errorf("Error in updating token : %v", err)
	}
	err = tx.Commit()
	if err != nil {
		a.log.Errorf("Error in updating token : %v", err)
	}
}*/

func (a *AuthService) CheckTokenExpire(){

	tx := a.db.MustBegin()

	tokenSQL := `UPDATE reset_password_token
				 	 SET is_expired = 1
					 WHERE created_at < DATETIME('now', '-180 minutes')`

	tx.MustExec(tokenSQL)

	err := tx.Commit()
	if err != nil {
		a.log.Errorf("Error in updating token : %v", err)
	}
}