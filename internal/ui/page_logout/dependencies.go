package page_logout

// Dependencies contains the dependencies required to render the logout page.
type Dependencies struct {
	Endpoint string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}
