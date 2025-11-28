package page_login

// Dependencies contains the dependencies required to render the login page.
type Dependencies struct {
	Passwordless       bool
	EnableRegistration bool

	Endpoint          string
	RedirectOnSuccess string

	// Layout is the outer layout function supplied by the auth package.
	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}
