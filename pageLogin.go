package auth

import (
	"net/http"

	"github.com/gouniverse/hb"
	"github.com/gouniverse/icons"
)

func (a Auth) pageLogin(w http.ResponseWriter, r *http.Request) {
	content := ""
	scripts := ""
	if a.passwordless {
		content = a.pageLoginPasswordlessContent()
		scripts = a.pageLoginPasswordlessScripts()
	} else {
		content = a.pageLoginContent()
		scripts = a.pageLoginScripts()
	}

	webpage := webpage("Login", a.funcLayout(content), scripts)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(webpage.ToHTML()))
}

func (a Auth) pageLoginPasswordlessContent() string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").Child(alertSuccess).Child(alertDanger)

	header := hb.NewHeading5().Text("Login").Style("margin:0px;")
	emailLabel := hb.NewLabel().Text("E-mail Address")
	emailInput := hb.NewInput().Class("form-control").Name("email").Placeholder("Enter e-mail address")
	emailFormGroup := hb.NewDiv().Class("form-group mt-3").Child(emailLabel).Child(emailInput)
	buttonLogin := hb.NewButton().Class("ButtonLogin btn btn-lg btn-success btn-block w-100").OnClick("loginFormValidate()").Children([]hb.TagInterface{
		icons.Icon("bi-send", 18, 18, "white").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Send me a login code"),
		hb.NewDiv().Class("ImgLoading spinner-border spinner-border-sm text-light").Style("display:none;margin-left:10px;"),
	})
	buttonLoginFormGroup := hb.NewDiv().Class("form-group mt-3 mb-3").Child(buttonLogin)
	buttonRegister := hb.NewHyperlink().Class("btn btn-info text-white float-start").Children([]hb.TagInterface{
		icons.Icon("bi-person-circle", 16, 16, "white").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Register"),
	}).Href(a.LinkRegister())

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").Child(header)
	cardBody := hb.NewDiv().Class("card-body").Children([]hb.TagInterface{
		alertGroup,
		emailFormGroup,
		buttonLoginFormGroup,
	})
	cardFooter := hb.NewDiv().Class("card-footer").AddChildren([]hb.TagInterface{})
	if a.enableRegistration {
		cardFooter.AddChild(buttonRegister)
	}

	card := hb.NewDiv().Class("card card-default").
		Style("margin:0 auto;max-width: 360px;").
		Child(cardHeader).
		Child(cardBody).
		Child(cardFooter)

	container := hb.NewDiv().Class("container").Child(card)

	return container.ToHTML()
}

func (a Auth) pageLoginPasswordlessScripts() string {
	urlApiLogin := a.LinkApiLogin()
	urlSuccess := a.LinkLoginCodeVerify()

	return `
	var urlApiLogin = "` + urlApiLogin + `";
	var urlOnSuccess = "` + urlSuccess + `";
    /**
     * Raises an error message
     * @param  {String} error
     * @returns  {Boolean}
     */
    function loginFormRaiseError(error) {
        $('div.alert-success').html('').hide();
        $('div.alert-danger').html(error).show();
        setTimeout(function () {
            $('div.alert-danger').html('').hide();
        }, 10000);
        return false;
    }

    function loginFormRaiseSuccess(success) {
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
    function loginFormValidate() {
        var email = $.trim($('input[name=email]').val());

        if (email === '') {
            return loginFormRaiseError('Email is required');
        }

        $('.ButtonLogin .ImgLoading').show();

        var data = {"email": email};

        $.post(urlApiLogin, data).then(function (response) {
            $('.ButtonLogin .ImgLoading').hide();

            if (response.status !== "success") {
                return loginFormRaiseError(response.message);
            }

            loginFormRaiseSuccess('Success');
            $('div.alert-danger').html('').hide();
            setTimeout(function () {
                $$.to(urlOnSuccess);
            }, 2000);
            return;
        }).fail(function (error) {
            console.log(error);
            $('.ButtonLogin .ImgLoading').hide();
            return loginFormRaiseError('There was an error. Try again later!');
        });
    }
    $(function () {
        $("#email").focus();
    });
	`
}

func (a Auth) pageLoginContent() string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Login").Style("margin:0px;")

	emailLabel := hb.NewLabel().
		HTML("E-mail Address")
	emailInput := hb.NewInput().
		Class("form-control").
		Name("email").
		Placeholder("Enter e-mail address")
	emailFormGroup := hb.NewDiv().Class("form-group mt-3").
		Child(emailLabel).
		Child(emailInput)

	passwordLabel := hb.NewLabel().
		HTML("Password")
	passwordInput := hb.NewInput().Class("form-control").
		Name("password").
		Type("password").
		Placeholder("Enter password")
	passwordFormGroup := hb.NewDiv().Class("form-group mt-3").
		Child(passwordLabel).
		Child(passwordInput)

	buttonLogin := hb.NewButton().
		Class("ButtonLogin btn btn-lg btn-success text-white btn-block w-100").
		Children([]hb.TagInterface{
			icons.Icon("bi-door-open", 18, 18, "white").Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().
				HTML("Log in"),
			hb.NewDiv().
				Class("ImgLoading spinner-border spinner-border-sm text-light").
				Style("display:none;margin-left:10px;"),
		}).
		OnClick("loginFormValidate()")

	buttonLoginFormGroup := hb.NewDiv().Class("form-group mt-3 mb-3").Child(buttonLogin)

	buttonRegister := hb.NewHyperlink().
		Class("btn btn-info text-white float-start").
		Children([]hb.TagInterface{
			icons.Icon("bi-person-circle", 16, 16, "white").Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().Text("Register"),
		}).
		Href(a.LinkRegister())

	buttonForgotPassword := hb.NewHyperlink().
		Class("btn btn-warning text-white float-end").
		Children([]hb.TagInterface{
			icons.Icon("bi-pass", 16, 16, "white").Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().Text("Forgot password?"),
		}).Href(a.LinkPasswordRestore())

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").AddChild(header)
	cardBody := hb.NewDiv().Class("card-body").
		AddChildren([]hb.TagInterface{
			alertGroup,
			emailFormGroup,
			passwordFormGroup,
			buttonLoginFormGroup,
		})
	cardFooter := hb.NewDiv().Class("card-footer").AddChildren([]hb.TagInterface{
		buttonForgotPassword,
	})

	if a.enableRegistration {
		cardFooter.AddChild(buttonRegister)
	}

	card := hb.NewDiv().Class("card card-default").
		Style("margin:0 auto;max-width: 360px;").
		Child(cardHeader).
		Child(cardBody).
		Child(cardFooter)

	container := hb.NewDiv().Class("container").AddChild(card)

	return container.ToHTML()
}

func (a Auth) pageLoginScripts() string {
	urlApiLogin := a.LinkApiLogin()
	urlSuccess := a.LinkRedirectOnSuccess()

	return `
	var urlApiLogin = "` + urlApiLogin + `";
	var urlOnSuccess = "` + urlSuccess + `";
    /**
     * Raises an error message
     * @param  {String} error
     * @returns  {Boolean}
     */
    function loginFormRaiseError(error) {
        $('div.alert-success').html('').hide();
        $('div.alert-danger').html(error).show();
        setTimeout(function () {
            $('div.alert-danger').html('').hide();
        }, 10000);
        return false;
    }

    function loginFormRaiseSuccess(success) {
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
    function loginFormValidate() {
        var email = $.trim($('input[name=email]').val());
        var password = $.trim($('input[name=password]').val());

        if (email === '') {
            return loginFormRaiseError('Email is required');
        }

        if (password === '') {
            return loginFormRaiseError('Password is required');
        }

        $('.ButtonLogin .ImgLoading').show();

        var data = {"email": email, "password": password};

        $.post(urlApiLogin, data).then(function (response) {
            $('.ButtonLogin .ImgLoading').hide();

            if (response.status !== "success") {
                return loginFormRaiseError(response.message);
            }

            $$.setAuthToken(response.data.token);
            $$.setAuthUser(response.data.user);
            loginFormRaiseSuccess('Success');
            $('div.alert-danger').html('').hide();
            setTimeout(function () {
                $$.to(urlOnSuccess);
            }, 2000);
            return;
        }).fail(function (error) {
			console.log(error);
            $('.ButtonLogin .ImgLoading').hide();
            return loginFormRaiseError('There was an error. Try again later!');
        });
    }
    $(function () {
        $("#email").focus();
    });
	`
}
