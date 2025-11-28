package page_login_code_verify

// Dependencies contains the dependencies required to render the login code verify
// page.
type Dependencies struct {
	Endpoint          string
	RedirectOnSuccess string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}
