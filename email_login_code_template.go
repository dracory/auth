package auth

import (
	"bytes"
	"html/template"
	"log/slog"
)

// emailLoginCodeTemplate returns the template for the login code verification email
func emailLoginCodeTemplate(email string, code string, options UserAuthOptions) string {
	msg := `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
<head></head>
<body>
	<p>
		Hello!
	<p>
	<p>
		Someone requested to login with your email {{.Email}}. Please use the code below to log in.
	</p>
	<p>
		{{.Code}}
	</p>
	<p>
		If you did not request to login no further action is required.
	</p>
	<p>
		Thanks,
		<br />
		The Admin Team
	</p>
</body>
<html>
`
	data := struct {
		Email string
		Code  string
	}{
		Email: email,
		Code:  code,
	}

	t, err := template.New("template").Parse(msg)
	if err != nil {
		slog.Error("login code email template parse failed",
			"error", err,
			"email", email,
			"code", code,
		)
		return ""
	}

	var doc bytes.Buffer
	errExecute := t.Execute(&doc, data)

	if errExecute != nil {
		slog.Error("login code email template execute failed",
			"error", errExecute,
			"email", email,
			"code", code,
		)
		return ""
	}

	s := doc.String()
	return s
}
