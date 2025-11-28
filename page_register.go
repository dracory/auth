package auth

import (
	"net/http"

	"github.com/dracory/hb"
)

func (a authImplementation) pageRegister(w http.ResponseWriter, r *http.Request) {
	content := ""
	scripts := ""
	if a.passwordless {
		content = a.pageRegisterPasswordlessContent()
		scripts = a.pageRegisterPasswordlessScripts()
	} else {
		content = a.pageRegisterUsernameAndPasswordContent()
		scripts = a.pageRegisterUsernameAndPasswordScripts()
	}

	webpage := webpage("Register", a.funcLayout(content), scripts)
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write register page response", "error", err)
	}
}

func (a authImplementation) pageRegisterPasswordlessContent() string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").Child(alertSuccess).Child(alertDanger)

	header := hb.NewHeading5().Text("Register").Style("margin:0px;")

	firstNameLabel := hb.NewLabel().Text("First Name")
	firstNameInput := hb.NewInput().Class("form-control").Name("first_name").Placeholder("Enter first name")
	firstNameFormGroup := hb.NewDiv().Class("form-group mt-3").Child(firstNameLabel).Child(firstNameInput)

	lastNameLabel := hb.NewLabel().Text("Last Name")
	lastNameInput := hb.NewInput().Class("form-control").Name("last_name").Placeholder("Enter last name")
	lastNameFormGroup := hb.NewDiv().Class("form-group mt-3").Child(lastNameLabel).Child(lastNameInput)

	emailLabel := hb.NewLabel().Text("E-mail Address")
	emailInput := hb.NewInput().Class("form-control").Name("email").Placeholder("Enter e-mail address")
	emailFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(emailLabel).AddChild(emailInput)

	buttonRegister := hb.NewButton().Class("btn btn-lg btn-success btn-block w-100").Children([]hb.TagInterface{
		hb.NewI().Class("bi bi-person-circle").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Register"),
	}).OnClick("registerFormValidate()")

	buttonRegisterFormGroup := hb.NewDiv().Class("form-group mt-3 mb-3").Child(buttonRegister)

	buttonLogin := hb.NewHyperlink().Class("btn btn-info text-white float-start").Children([]hb.TagInterface{
		hb.NewI().Class("bi bi-send").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Login"),
	}).Href(a.LinkLogin())

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").AddChild(header)
	cardBody := hb.NewDiv().Class("card-body").AddChildren([]hb.TagInterface{
		alertGroup,
		firstNameFormGroup,
		lastNameFormGroup,
		emailFormGroup,
		buttonRegisterFormGroup,
	})
	cardFooter := hb.NewDiv().Class("card-footer").Children([]hb.TagInterface{
		buttonLogin,
	})
	card := hb.NewDiv().Class("card card-default").Style("margin:0 auto;max-width: 360px;")
	card.AddChild(cardHeader).Child(cardBody).Child(cardFooter)

	container := hb.NewDiv().Class("container").Child(card)

	return container.ToHTML()
}

func (a authImplementation) pageRegisterPasswordlessScripts() string {
	urlApiRegister := a.LinkApiRegister()
	urlSuccess := a.LinkRegisterCodeVerify()

	return `
	var urlApiRegister = "` + urlApiRegister + `";
	console.log(urlApiRegister);
	var urlOnSuccess = "` + urlSuccess + `";
    /**
     * Raises an error message
     * @param  {String} error
     * @returns  {Boolean}
     */
    function registerFormRaiseError(error) {
        $('div.alert-success').html('').hide();
        $('div.alert-danger').html(error).show();
        setTimeout(function () {
            $('div.alert-danger').html('').hide();
        }, 10000);
        return false;
    }

    function registerFormRaiseSuccess(success) {
        $('div.alert-danger').html('').hide();
        $('div.alert-success').html(success).show();
        setTimeout(function () {
            $('div.alert-success').html('').hide();
        }, 10000);
        return false;
    }

    /**
     * Validate Login Form
     * @returns  {Boolean}
     */
    function registerFormValidate() {
		var first_name = $.trim($('input[name=first_name]').val());
		var last_name = $.trim($('input[name=last_name]').val());
        var email = $.trim($('input[name=email]').val());
        var password = $.trim($('input[name=password]').val());

		if (first_name === '') {
            return registerFormRaiseError('First name is required');
        }

		if (last_name === '') {
            return registerFormRaiseError('Last name is required');
        }

        if (email === '') {
            return registerFormRaiseError('Email is required');
        }

        $('.buttonLogin .imgLoading').show();

        var data = {"first_name": first_name, "last_name": last_name, "email": email};

        $.post(urlApiRegister, data).then(function (response) {
            $('.buttonLogin .imgLoading').hide();

            if (response.status !== "success") {
                return registerFormRaiseError(response.message);
            }

            registerFormRaiseSuccess('Success');
            $('div.alert-danger').html('').hide();
            setTimeout(function () {
                window.location.href=urlOnSuccess;
            }, 100);
            return;
        }).fail(function (error) {
			console.log(error);
            $('.buttonLogin .imgLoading').hide();
            return registerFormRaiseError('There was an error. Try again later!');
        });
    }
    $(function () {
        $("input[name=first_name").focus();
    });
	`
}

func (a authImplementation) pageRegisterUsernameAndPasswordContent() string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Register").Style("margin:0px;")
	firstNameLabel := hb.NewLabel().Text("First Name")
	firstNameInput := hb.NewInput().Class("form-control").Name("first_name").Placeholder("Enter first name")
	firstNameFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(firstNameLabel).AddChild(firstNameInput)
	lastNameLabel := hb.NewLabel().Text("Last Name")
	lastNameInput := hb.NewInput().Class("form-control").Name("last_name").Placeholder("Enter last name")
	lastNameFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(lastNameLabel).AddChild(lastNameInput)
	emailLabel := hb.NewLabel().Text("E-mail Address")
	emailInput := hb.NewInput().Class("form-control").Name("email").Placeholder("Enter e-mail address")
	emailFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(emailLabel).AddChild(emailInput)
	passwordLabel := hb.NewLabel().AddChild(hb.NewHTML("Password"))
	passwordInput := hb.NewInput().Class("form-control").Name("password").Type(hb.TYPE_PASSWORD).Placeholder("Enter password")
	passwordFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(passwordLabel).AddChild(passwordInput)
	buttonRegister := hb.NewButton().Class("btn btn-lg btn-success btn-block w-100").Text("Register").OnClick("registerFormValidate()")
	buttonRegisterFormGroup := hb.NewDiv().Class("form-group mt-3 mb-3").AddChild(buttonRegister)
	buttonLogin := hb.NewHyperlink().Class("btn btn-info float-start").Text("Login").Href(a.LinkLogin())
	buttonForgotPassword := hb.NewHyperlink().Class("btn btn-warning float-end").Text("Forgot password?").Href(a.LinkPasswordRestore())

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").AddChild(header)
	cardBody := hb.NewDiv().Class("card-body").AddChildren([]hb.TagInterface{
		alertGroup,
		firstNameFormGroup,
		lastNameFormGroup,
		emailFormGroup,
		passwordFormGroup,
		buttonRegisterFormGroup,
	})
	cardFooter := hb.NewDiv().Class("card-footer").AddChildren([]hb.TagInterface{
		buttonLogin,
		buttonForgotPassword,
	})
	card := hb.NewDiv().Class("card card-default").Style("margin:0 auto;max-width: 360px;")
	card.AddChild(cardHeader).AddChild(cardBody).AddChild(cardFooter)

	container := hb.NewDiv().Class("container").Child(card)

	return container.ToHTML()
}

func (a authImplementation) pageRegisterUsernameAndPasswordScripts() string {
	urlApiRegister := a.LinkApiRegister()
	urlSuccess := a.LinkLogin()
	if a.enableVerification {
		urlSuccess = a.LinkRegisterCodeVerify()
	}

	return `
	var urlApiRegister = "` + urlApiRegister + `";
	console.log(urlApiRegister);
	var urlOnSuccess = "` + urlSuccess + `";
    /**
     * Raises an error message
     * @param  {String} error
     * @returns  {Boolean}
     */
    function registerFormRaiseError(error) {
        $('div.alert-success').html('').hide();
        $('div.alert-danger').html(error).show();
        setTimeout(function () {
            $('div.alert-danger').html('').hide();
        }, 10000);
        return false;
    }

    function registerFormRaiseSuccess(success) {
        $('div.alert-danger').html('').hide();
        $('div.alert-success').html(success).show();
        setTimeout(function () {
            $('div.alert-success').html('').hide();
        }, 10000);
        return false;
    }

    /**
     * Validate Login Form
     * @returns  {Boolean}
     */
    function registerFormValidate() {
		var first_name = $.trim($('input[name=first_name]').val());
		var last_name = $.trim($('input[name=last_name]').val());
        var email = $.trim($('input[name=email]').val());
        var password = $.trim($('input[name=password]').val());

		if (first_name === '') {
            return registerFormRaiseError('First name is required');
        }

		if (last_name === '') {
            return registerFormRaiseError('Last name is required');
        }

        if (email === '') {
            return registerFormRaiseError('Email is required');
        }

        if (password === '') {
            return registerFormRaiseError('Password is required');
        }

        $('.buttonLogin .imgLoading').show();

        var data = {"first_name": first_name, "last_name": last_name, "email": email, "password": password};

        $.post(urlApiRegister, data).then(function (response) {
            $('.buttonLogin .imgLoading').hide();

            if (response.status !== "success") {
                return registerFormRaiseError(response.message);
            }

            registerFormRaiseSuccess('Success');
            $('div.alert-danger').html('').hide();
            setTimeout(function () {
                window.location.href=urlOnSuccess;
            }, 2000);
            return;
        }).fail(function (error) {
			console.log(error);
            $('.buttonLogin .imgLoading').hide();
            return registerFormRaiseError('There was an error. Try again later!');
        });
    }
    $(function () {
        $("input[name=first_name").focus();
    });
	`
}
