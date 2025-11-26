package auth

import (
	"bytes"
	"html/template"
	"log"
)

// emailRegisterCodeTemplate returns the template for the register code verification email
func emailRegisterCodeTemplate(email string, code string, options UserAuthOptions) string {
	msg := `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html>
<head></head>
<body>
	<p>
		Hello!
	<p>
	<p>
		Someone requested to register with your email {{.Email}}. Please use the code below to confirm your registration.
	</p>
	<p>
		{{.Code}}
	</p>
	<p>
		If you did not request to register no further action is required.
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
		log.Println(err)
		return ""
	}

	var doc bytes.Buffer
	errExecute := t.Execute(&doc, data)

	if errExecute != nil {
		log.Println(errExecute)
		return ""
	}

	s := doc.String()
	return s
}
