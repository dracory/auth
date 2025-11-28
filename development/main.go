package main

import (
	"context"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/darkoatanasovski/htmltags"
	"github.com/dracory/auth"
	"github.com/dracory/auth/development/scribble"
	"github.com/dracory/env"
	"github.com/jordan-wright/email"
)

// ctx-aware wrapper callbacks adapting existing dev helpers to the new ctx-first signatures

func emailSendWithContext(ctx context.Context, userID string, subject string, body string) error {
	return emailSend(userID, subject, body)
}

func userFindByAuthTokenWithContext(ctx context.Context, sessionID string, options auth.UserAuthOptions) (string, error) {
	return userFindByAuthToken(sessionID, options)
}

func userFindByUsernameWithContext(ctx context.Context, username string, firstName string, lastName string, options auth.UserAuthOptions) (string, error) {
	return userFindByUsername(username, firstName, lastName, options)
}

func userLoginWithContext(ctx context.Context, username string, password string, options auth.UserAuthOptions) (string, error) {
	return userLogin(username, password, options)
}

func userLogoutWithContext(ctx context.Context, userID string, options auth.UserAuthOptions) error {
	return userLogout(userID, options)
}

func userPasswordChangeWithContext(ctx context.Context, username string, newPassword string, options auth.UserAuthOptions) error {
	return userPasswordChange(username, newPassword, options)
}

func userRegisterWithContext(ctx context.Context, username string, password string, firstName string, lastName string, options auth.UserAuthOptions) error {
	return userRegister(username, password, firstName, lastName, options)
}

func userStoreAuthTokenWithContext(ctx context.Context, sessionID string, userID string, options auth.UserAuthOptions) error {
	return userStoreAuthToken(sessionID, userID, options)
}

func passwordlessUserFindByEmailWithContext(ctx context.Context, email string, options auth.UserAuthOptions) (string, error) {
	return passwordlessUserFindByEmail(email, options)
}

func passwordlessUserRegisterWithContext(ctx context.Context, email string, firstName string, lastName string, options auth.UserAuthOptions) error {
	return passwordlessUserRegister(email, firstName, lastName, options)
}

func main() {
	os.Remove(env.GetString("DB_DATABASE")) // remove database
	log.Println("1. Initializing environment variables...")
	env.Load(".env")

	log.Println("2. Initializing database...")
	var err error
	jsonStore, err = scribble.New("temp", nil)
	if err != nil {
		log.Panic("Database is NIL: " + err.Error())
		return
	}

	authUsernameAndPassword, errUsernameAndPassword := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                env.GetString("APP_URL") + "/auth-username-and-password",
		UrlRedirectOnSuccess:    "/user/dashboard-after-username-and-password",
		FuncEmailSend:           emailSendWithContext,
		FuncUserFindByAuthToken: userFindByAuthTokenWithContext,
		FuncUserFindByUsername:  userFindByUsernameWithContext,
		FuncUserLogin:           userLoginWithContext,
		FuncUserLogout:          userLogoutWithContext,
		FuncUserPasswordChange:  userPasswordChangeWithContext,
		FuncUserRegister:        userRegisterWithContext,
		FuncUserStoreAuthToken:  userStoreAuthTokenWithContext,
		FuncTemporaryKeyGet:     temporaryKeyGet,
		FuncTemporaryKeySet:     temporaryKeySet,
		UseCookies:              true,
		EnableRegistration:      true,
		EnableVerification:      true,
	})

	if errUsernameAndPassword != nil {
		log.Panicln(errUsernameAndPassword.Error())
	}

	authPasswordless, errPasswordless := auth.NewPasswordlessAuth(auth.ConfigPasswordless{
		Endpoint:             env.GetString("APP_URL") + "/auth-passwordless",
		UrlRedirectOnSuccess: "/user/dashboard-after-passwordless",

		EnableRegistration: true,

		FuncEmailSend:           emailSendWithContext,
		FuncTemporaryKeyGet:     temporaryKeyGet,
		FuncTemporaryKeySet:     temporaryKeySet,
		FuncUserFindByEmail:     passwordlessUserFindByEmailWithContext,
		FuncUserFindByAuthToken: userFindByAuthTokenWithContext,
		FuncUserLogout:          userLogoutWithContext,
		FuncUserRegister:        passwordlessUserRegisterWithContext,
		FuncUserStoreAuthToken:  userStoreAuthTokenWithContext,

		UseCookies: true,
	})

	if errPasswordless != nil {
		log.Panicln(errPasswordless.Error())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := "<h1>Index Page</h1>"
		html += "<p>Login with username and password at: <a href='" + authUsernameAndPassword.LinkLogin() + "'>" + authUsernameAndPassword.LinkLogin() + "</a></p>"
		html += "<p>Login without password at: <a href='" + authPasswordless.LinkLogin() + "'>" + authPasswordless.LinkLogin() + "</a></p>"
		w.Write([]byte("<html>" + html))
	})
	mux.HandleFunc("/auth-username-and-password/", authUsernameAndPassword.AuthHandler)
	mux.Handle("/user/dashboard-after-username-and-password", authUsernameAndPassword.WebAuthOrRedirectMiddleware(messageHandler("<html>User page. Logout at: <a href='"+authUsernameAndPassword.LinkLogout()+"'>"+authUsernameAndPassword.LinkLogout()+"</a>")))

	mux.HandleFunc("/auth-passwordless/", authPasswordless.AuthHandler)
	mux.Handle("/user/dashboard-after-passwordless", authPasswordless.WebAuthOrRedirectMiddleware(messageHandler("<html>User page. Logout at: <a href='"+authPasswordless.LinkLogout()+"'>"+authPasswordless.LinkLogout()+"</a>")))

	log.Println("4. Starting server on http://" + env.GetString("SERVER_HOST") + ":" + env.GetString("SERVER_PORT") + " ...")
	if strings.HasPrefix(env.GetString("APP_URL"), "https://") {
		log.Println(env.GetString("APP_URL") + " ...")
	} else {
		log.Println("URL: http://" + env.GetString("APP_URL") + " ...")
	}

	srv := &http.Server{
		Handler: mux,
		Addr:    env.GetString("SERVER_HOST") + ":" + env.GetString("SERVER_PORT"),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

// func Middleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("Here"))
// 		next.ServeHTTP(w, r)
// 	})
// }

func messageHandler(message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(message))
	})
}

// func mainDb(driverName string, dbHost string, dbPort string, dbName string, dbUser string, dbPass string) (*sql.DB, error) {
// 	var db *sql.DB
// 	var err error
// 	if driverName == "sqlite" {
// 		dsn := dbName + "?parseTime=true"
// 		db, err = sql.Open("sqlite3", dsn)
// 		// dsn := dbName
// 		// db, err = sql.Open("sqlite", dsn)
// 	}
// 	if driverName == "mysql" {
// 		dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=Local"
// 		db, err = sql.Open("mysql", dsn)
// 	}
// 	if driverName == "postgres" {
// 		dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort + " sslmode=disable TimeZone=Europe/London"
// 		db, err = sql.Open("postgres", dsn)
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	if db == nil {
// 		return nil, errors.New("database for driver " + driverName + " could not be intialized")
// 	}

// 	return db, nil
// }

// EmailSend sends an email
func emailSendTo(from string, to []string, subject string, htmlMessage string) (bool, error) {
	//drvr := os.Getenv("MAIL_DRIVER")
	host := env.GetString("MAIL_HOST")
	port := env.GetString("MAIL_PORT")
	user := env.GetString("MAIL_USERNAME")
	pass := env.GetString("MAIL_PASSWORD")
	addr := host + ":" + port

	nodes, errStripped := htmltags.Strip(htmlMessage, []string{}, true)

	textMessage := ""

	if errStripped == nil {
		//nodes.Elements   //HTML nodes structure of type *html.Node
		textMessage = nodes.ToString() //returns stripped HTML string
	}

	e := email.NewEmail()
	e.From = from
	e.To = to
	e.Subject = subject
	e.Text = []byte(textMessage)
	e.HTML = []byte(htmlMessage)
	err := e.Send(addr, smtp.PlainAuth("", user, pass, host))

	if err != nil {
		log.Fatal(err)
		return false, err
	}
	return true, nil
}
