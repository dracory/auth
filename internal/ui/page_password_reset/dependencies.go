package page_password_reset

// Dependencies contains the dependencies required to render the password reset page.
type Dependencies struct {
	Endpoint           string
	EnableRegistration bool
	Token              string
	ErrorMessage       string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}
