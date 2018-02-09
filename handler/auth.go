package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/leoleung0102/go-graphql-starter/config"
	"github.com/leoleung0102/go-graphql-starter/model"
	"github.com/leoleung0102/go-graphql-starter/service"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			isAuthorized = false
			userId       int64
		)
		ctx := r.Context()
		token, err := validateBearerAuthHeader(ctx, r)
		if err == nil {
			isAuthorized = true
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userIdByte, _ := base64.StdEncoding.DecodeString(claims["id"].(string))
				userId, _ = strconv.ParseInt(string(userIdByte[:]), 10, 64)
			} else {
				log.Println(err)
			}
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Println(w, "Requester ip: %q is not IP:port", r.RemoteAddr)
		}

		ctx = context.WithValue(ctx, "user_id", &userId)
		ctx = context.WithValue(ctx, "requester_ip", &ip)
		ctx = context.WithValue(ctx, "is_authorized", isAuthorized)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx := r.Context()
		loginResponse := &model.LoginResponse{}

		if r.Method != http.MethodPost {
			response := &model.Response{
				Code:  http.StatusMethodNotAllowed,
				Error: config.PostMethodSupported,
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}
		userCredentials, err := validateBasicAuthHeader(r)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: err.Error(),
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}
		user, err := ctx.Value("userService").(*service.UserService).ComparePassword(userCredentials)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusUnauthorized,
				Error: err.Error(),
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}

		tokenString, err := ctx.Value("authService").(*service.AuthService).SignJWT(user)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: config.TokenError,
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}

		response := &model.Response{
			Code: http.StatusOK,
		}
		loginResponse.Response = response
		loginResponse.AccessToken = *tokenString
		writeResponse(w, loginResponse, loginResponse.Code)
	})
}

func GenerateToken() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx := r.Context()
		tokenResponse := &model.TokenResponse{}
		userEmail := r.Header.Get("Email")

		encodedEmail := base64.StdEncoding.EncodeToString([]byte(userEmail))

		if r.Method != http.MethodPost {
			response := &model.Response{
				Code:  http.StatusMethodNotAllowed,
				Error: config.PostMethodSupported,
			}
			tokenResponse.Response = response
			writeResponse(w, tokenResponse, tokenResponse.Code)
			return
		}

		user,err := ctx.Value("userService").(*service.UserService).FindByEmail(userEmail)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: err.Error(),
			}
			tokenResponse.Response = response
			writeResponse(w, tokenResponse, tokenResponse.Code)
			return
		}

		tokenForURL, err := ctx.Value("authService").(*service.AuthService).GenerateToken(user)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: err.Error(),
			}
			tokenResponse.Response = response
			writeResponse(w, tokenResponse, tokenResponse.Code)
			return
		}

		go ctx.Value("emailService").(*service.EmailService).SendEmail(
			"leoleung@inno-lab.co",
			userEmail,
			"Reset Password",
			"",
			"reset",
			"http://localhost:3001/reset-password-validation?user=" + encodedEmail + "&token=" + tokenForURL,
		)

		response := &model.Response{
			Code: http.StatusOK,
		}

		tokenResponse.Response = response
		writeResponse(w, tokenResponse, tokenResponse.Code)
	})
}

func writeResponse(w http.ResponseWriter, response interface{}, code int) {
	jsonResponse, _ := json.Marshal(response)
	w.WriteHeader(code)
	w.Write(jsonResponse)
}

func validateBasicAuthHeader(r *http.Request) (*model.UserCredentials, error) {
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		return nil, errors.New(config.CredentialsError)
	}
	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return nil, errors.New(config.CredentialsError)
	}
	userCredentials := model.UserCredentials{
		Email:    pair[0],
		Password: pair[1],
	}
	return &userCredentials, nil
}

func CheckTokenValidation() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx := r.Context()

		encodedToken := r.URL.Query().Get("token")
		encodedEmail := r.URL.Query().Get("user")

		email,_ := base64.StdEncoding.DecodeString(encodedEmail)
		token,_ := base64.StdEncoding.DecodeString(encodedToken)

		if r.Method != http.MethodGet {
			response := &model.Response{
				Code:  http.StatusMethodNotAllowed,
				Error: config.PostMethodSupported,
			}
			writeResponse(w, response, response.Code)
			return
		}

		//Should set to cron job
		//go ctx.Value("authService").(*service.AuthService).CheckTokenExpire()

		validToken, err := ctx.Value("authService").(*service.AuthService).CheckTokenValidation(string(email),string(token))

		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: err.Error(),
			}
			writeResponse(w, response, response.Code)
			return
		}

		response := &model.Response{
			Code: http.StatusOK,
		}

		if validToken != "" {
			http.Redirect(w,r,"http://localhost:3001/reset-password?user=" + encodedEmail + "&token=" + validToken, 301)
		}else{
			http.Redirect(w,r,"http://localhost:3001/not-found", 404)
		}

		writeResponse(w, response, response.Code)
	})
}

func ResetPassword() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		//ctx := r.Context()

		/*token := r.URL.Query().Get("token")
		encodedEmail := r.URL.Query().Get("user")

		email,_ := base64.StdEncoding.DecodeString(encodedEmail)*/

		if r.Method != http.MethodGet {
			response := &model.Response{
				Code:  http.StatusMethodNotAllowed,
				Error: config.PostMethodSupported,
			}
			writeResponse(w, response, response.Code)
			return
		}

		//ctx.Value("userService").(*service.UserService).ResetPassword()

		/*validToken, err := ctx.Value("userService").(*service.UserService).ResetPassword()

		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: err.Error(),
			}
			writeResponse(w, response, response.Code)
			return
		}*/

		response := &model.Response{
			Code: http.StatusOK,
		}

		writeResponse(w, response, response.Code)
	})
}

func validateBearerAuthHeader(ctx context.Context, r *http.Request) (*jwt.Token, error) {
	var tokenString string
	keys, ok := r.URL.Query()["at"]
	if !ok || len(keys) < 1 {
		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Bearer" {
			return nil, errors.New(config.CredentialsError)
		}
		tokenString = auth[1]
	} else {
		tokenString = keys[0]
	}
	token, err := ctx.Value("authService").(*service.AuthService).ValidateJWT(&tokenString)
	return token, err
}
