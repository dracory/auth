package page_register

// Dependencies contains the dependencies required to render the register page.
type Dependencies struct {
	Passwordless       bool
	EnableVerification bool

	Endpoint string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}
