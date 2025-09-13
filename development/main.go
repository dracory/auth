package main

import (
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/darkoatanasovski/htmltags"
	"github.com/dracory/auth/development/scribble"
	"github.com/jordan-wright/email"

	"github.com/dracory/auth"

	"github.com/gouniverse/utils"
)

func main() {
	os.Remove(utils.Env("DB_DATABASE")) // remove database
	log.Println("1. Initializing environment variables...")
	utils.EnvInitialize(".env")

	log.Println("2. Initializing database...")
	var err error
	jsonStore, err = scribble.New("temp", nil)
	if err != nil {
		log.Panic("Database is NIL: " + err.Error())
		return
	}

	authUsernameAndPassword, errUsernameAndPassword := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                utils.Env("APP_URL") + "/auth-username-and-password",
		UrlRedirectOnSuccess:    "/user/dashboard-after-username-and-password",
		FuncEmailSend:           emailSend,
		FuncUserFindByAuthToken: userFindByAuthToken,
		FuncUserFindByUsername:  userFindByUsername,
		FuncUserLogin:           userLogin,
		FuncUserLogout:          userLogout,
		FuncUserPasswordChange:  userPasswordChange,
		FuncUserRegister:        userRegister,
		FuncUserStoreAuthToken:  userStoreAuthToken,
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
		Endpoint:             utils.Env("APP_URL") + "/auth-passwordless",
		UrlRedirectOnSuccess: "/user/dashboard-after-passwordless",

		EnableRegistration: true,

		FuncEmailSend:           emailSend,
		FuncTemporaryKeyGet:     temporaryKeyGet,
		FuncTemporaryKeySet:     temporaryKeySet,
		FuncUserFindByEmail:     passwordlessUserFindByEmail,
		FuncUserFindByAuthToken: userFindByAuthToken,
		FuncUserLogout:          userLogout,
		FuncUserRegister:        passwordlessUserRegister,
		FuncUserStoreAuthToken:  userStoreAuthToken,

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
	mux.Handle("/user/dashboard-after-username-and-password", authUsernameAndPassword.AuthMiddleware(messageHandler("<html>User page. Logout at: <a href='"+authUsernameAndPassword.LinkLogout()+"'>"+authUsernameAndPassword.LinkLogout()+"</a>")))

	mux.HandleFunc("/auth-passwordless/", authPasswordless.AuthHandler)
	mux.Handle("/user/dashboard-after-passwordless", authPasswordless.AuthMiddleware(messageHandler("<html>User page. Logout at: <a href='"+authPasswordless.LinkLogout()+"'>"+authPasswordless.LinkLogout()+"</a>")))

	log.Println("4. Starting server on http://" + utils.Env("SERVER_HOST") + ":" + utils.Env("SERVER_PORT") + " ...")
	if strings.HasPrefix(utils.Env("APP_URL"), "https://") {
		log.Println(utils.Env("APP_URL") + " ...")
	} else {
		log.Println("URL: http://" + utils.Env("APP_URL") + " ...")
	}

	srv := &http.Server{
		Handler: mux,
		Addr:    utils.Env("SERVER_HOST") + ":" + utils.Env("SERVER_PORT"),
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
	host := utils.Env("MAIL_HOST")
	port := utils.Env("MAIL_PORT")
	user := utils.Env("MAIL_USERNAME")
	pass := utils.Env("MAIL_PASSWORD")
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
