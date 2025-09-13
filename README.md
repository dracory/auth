# Auth <a href="https://gitpod.io/#https://github.com/dracory/auth" style="float:right:"><img src="https://gitpod.io/button/open-in-gitpod.svg" alt="Open in Gitpod" loading="lazy"></a>

[![Tests Status](https://github.com/dracory/auth/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/dracory/auth/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/auth)](https://goreportcard.com/report/github.com/dracory/auth)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/auth)](https://pkg.go.dev/github.com/dracory/auth)

<img src="logo.jpg" width="100%" />

Authentication library for Golang with two separate flows: 

1. Username/email and password - The user logs in via a username and password. Using email instead of username is also supported. These days most of the applications will use email instad of username, as it is more convenient for the user not having to remember a username.

2. Passwordless - A verification code is emailed to the user on each login. The user does not have to remember a password. This flow is recommended because it is safer as it relieves the application from having to securely store passwords.

The aim of this library is to provide a quick preset for authentication, which includes:

1. User interface

2. HTTP handler

3. Authentication middleware

It will then leave the actual implementation to you - where to save the tokens, session, users, etc.

## Installation

```sh
go get github.com/dracory/auth
```

## Usage of the Username/email and Password Flow

- Implement your functions

```golang
// userFindByEmail find the user by the provided email, and returns the user ID
//
// retrieve the userID from your database
// note that the username can be an email (if you wish)
//
func userFindByUsername(username string, firstName string, lastName string) (userID string, err error) {
    // your code here
	return "yourUserId", nil
}

// userRegister registers the user
//
// save the user to your databases
// note that the username can be an email (if you wish)
//
func userRegister(username string, password string, first_name string, last_name string) error {
    // your code here
	return nil
}

// userLogin logs the user in
//
// find the user by the specified username and password, 
// note that the username can be an email (if you wish)
//
func userLogin(username string, password string, userIP string, userAgent string) (userID string, err error) {
    // your code here
	return "yourUserId", nil
}

// userLogout logs the user out
//
// remove the auth token from wherever you have stored it (i.e. session store or the cache store)
//
func userLogout(userID string) (err error) {
    // your code here (remove token from session or cache store)
	return nil
}

// userStoreAuthToken stores the auth token with the provided user ID
//
// save the auth token to your selected store it (i.e. session store or the cache store)
// make sure you set an expiration time (i.e. 2 hours)
//
func userStoreAuthToken(token string, userID string, userIP string, userAgent string) error {
    // your code here (store in session or cache store with desired timeout)
	return nil
}

// userFindByAuthToken find the user by the provided token, and returns the user ID
//
// retrieve the userID from your selected store  (i.e. session store or the cache store)
//
func userFindByAuthToken(token string, userIP string, userAgent string) (userID string, err error) {
    // your code here
	return "yourUserId", nil
}
```

- Setup the auth settings

```golang
auth, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
	EnableRegistration:              	true,
	EnableVerification:              	true, // optional, required only if you want the email to be verified before completing the registration
	Endpoint:                        	"/",
	UrlRedirectOnSuccess:            	"http://localhost/user/dashboard",
	FuncUserFindByAuthToken:         	userFindByAuthToken,
	FuncUserFindByUsername:          	userFindByUsername,
	FuncUserLogin:                   	userLogin,
	FuncUserLogout:                  	userLogout,
	FuncUserRegister:                	userRegister, // optional, required only if registration is enabled
	FuncUserStoreAuthToken:          	userStoreAuthToken,
	FuncEmailSend:				emailSend,
	FuncEmailTemplatePasswordRestore:	emailTemplatePasswordRestore, // optional, if you want to set custom email template
	FuncTemporaryKeyGet:             	tempKeyGet,
	FuncTemporaryKeySet:             	tempKeySet,
})
```

- Attach to router

```golang
mux := http.NewServeMux()

// Example index page with login link
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is the index page visible to anyone on the internet. Login at: " + auth.LinkLogin()))
})

// Attach the authentication URLs
mux.HandleFunc("/auth/", auth.Router)
```

- Used the AuthMiddleware to protect the authenticated routes

```golang
// Put your auth routes after the Auth middleware
mux.Handle("/user/dashboard", auth.AuthMiddleware(dashboardHandler("IN AUTHENTICATED DASHBOARD")))
```


## Usage of the Passwordless Flow

- Implement your functions

```golang
// emailSend sends an email to the provided email
func emailSend(email string, emailSubject string, emailBody string) error {
	// your code here
	return nil
}

// userFindByEmail find the user by the provided email, and returns the user ID
//
// retrieve the userID from your database
//
func userFindByEmail(email string) (userID string, err error) {
	// your code here
	return "yourUserId", nil
}

// userRegister registers the user
//
// save the user to your databases
// note that the username can be an email (if you wish)
//
func userRegister(email string, first_name string, last_name string) error {
    // your code here
	return nil
}

// userLogout logs the user out
//
// remove the auth token from wherever you have stored it (i.e. session store or the cache store)
//
func userLogout(userID string) (err error) {
    // your code here (remove token from session or cache store)
	return nil
}

// userStoreAuthToken stores the auth token with the provided user ID
//
// save the auth token to your selected store it (i.e. session store or the cache store)
// make sure you set an expiration time (i.e. 2 hours)
//
func userStoreAuthToken(token string, userID string) error {
    // your code here (store in session or cache store with desired timeout)
	return nil
}

// userFindByAuthToken find the user by the provided token, and returns the user ID
//
// retrieve the userID from your selected store  (i.e. session store or the cache store)
//
func userFindByAuthToken(token string) (userID string, err error) {
    // your code here
	return "yourUserId", nil
}
```

- Setup the auth settings

```golang
auth, err := auth.NewPasswordlessAuth(auth.ConfigPasswordless{
	EnableRegistration:		true,				// optional, required only if registration is required
	Endpoint:			"/",
	UrlRedirectOnSuccess:		"http://localhost/user/dashboard",
	FuncUserFindByAuthToken:	userFindByAuthToken,
	FuncUserFindByEmail:		userFindByEmail,
	FuncUserLogout:			userLogout,
	FuncUserRegister:		userRegister,			// optional, required only if registration is enabled
	FuncUserStoreAuthToken:		userStoreAuthToken,
	FuncEmailSend:			emailSend,
	FuncEmailTemplateLoginCode:	emailLoginCodeTemplate,		// optional, if you want to customize the template
	FuncEmailTemplateRegisterCode:	emailRegisterCodeTemplate,	// optional, if you want to customize the template
	FuncTemporaryKeyGet:        	tempKeyGet,
	FuncTemporaryKeySet:		tempKeySet,
})
```

- Attach to router

```golang
mux := http.NewServeMux()

// Example index page with login link
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is the index page visible to anyone on the internet. Login at: " + auth.LinkLogin()))
})

// Attach the authentication URLs
mux.HandleFunc("/auth/", auth.Router)
```

- Used the AuthMiddleware to protect the authenticated routes

```golang
// Put your auth routes after the Auth middleware
mux.Handle("/user/dashboard", auth.AuthMiddleware(dashboardHandler("IN AUTHENTICATED DASHBOARD")))
```



## Frequently Asked Questions

1. Can I use email and password instead of username and password to login users?

Yes you absolutely can.

2. Can I use username and password flow for regular users and passwordless flow for admin users?

Yes you can. You just instantiate two separate instances, and atatch each separate HTTP handler to listen on its own path. For instance you may use /auth for regular users and /auth-admin for administrators


## Other Noteable Auth Projects

- https://github.com/authorizerdev/authorizer
- https://github.com/markbates/goth
- https://github.com/teamhanko/hanko
- https://github.com/go-pkgz/auth
