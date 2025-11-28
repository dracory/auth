package auth

import (
	"net/http"

	"github.com/dracory/hb"
	"github.com/dracory/req"
)

func (a authImplementation) pagePasswordReset(w http.ResponseWriter, r *http.Request) {
	token := req.GetString(r, "t")
	errorMessage := ""

	if token == "" {
		errorMessage = "Link is invalid"
	} else {
		tokenValue, errToken := a.funcTemporaryKeyGet(token)
		if errToken != nil {
			errorMessage = "Link has expired"
		} else if tokenValue == "" {
			errorMessage = "Link is invalid or expired"
		}
	}

	h := a.pagePasswordResetContent(token, errorMessage)
	webpage := webpage("Reset Password", h, a.pagePasswordResetScripts())
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write password reset page response", "error", err)
	}
}

func (a authImplementation) pagePasswordResetContent(token string, errorMessage string) string {
	urlPasswordRestore := a.LinkPasswordRestore()
	urlLogin := a.LinkLogin()
	urlRegister := a.LinkRegister()
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger")
	if errorMessage != "" {
		alertDanger.Text(errorMessage)
	} else {
		alertDanger.Style("display:none")
	}
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Reset Password").Style("margin:0px;")
	tokenInput := hb.NewInput().Name("token").Value(token)
	passwordLabel := hb.NewLabel().Text("New Password")
	passwordInput := hb.NewInput().Class("form-control").Name("password").Placeholder("Enter new password")
	passwordFormGroup := hb.NewDiv().Class("form-group mt-3").Child(passwordLabel).Child(passwordInput)
	passwordConfirmLabel := hb.NewLabel().Text("Confirm New Password")
	passwordConfirmInput := hb.NewInput().Class("form-control").Name("password_confirm").Placeholder("Enter confirmation of new password")
	passwordConfirmFormGroup := hb.NewDiv().Class("form-group mt-3").Child(passwordConfirmLabel).Child(passwordConfirmInput)
	buttonContinue := hb.NewButton().Class("ButtonContinue btn btn-lg btn-success btn-block w-100").Text("Reset Password").OnClick("resetFormValidate()")
	buttonContinueFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(buttonContinue)
	buttonLogin := hb.NewHyperlink().Class("btn btn-info float-start").Text("Login").Href(a.LinkLogin())
	buttonRegister := hb.NewHyperlink().Class("btn btn-warning float-end").Text("Register").Href(a.LinkRegister())

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").AddChild(header)
	cardBody := hb.NewDiv().Class("card-body").AddChildren([]hb.TagInterface{
		alertGroup,
	})

	if errorMessage == "" {
		cardBody.AddChild(tokenInput)
		cardBody.AddChild(passwordFormGroup)
		cardBody.AddChild(passwordConfirmFormGroup)
		cardBody.AddChild(buttonContinueFormGroup)
	} else {
		cardBody.AddChild(hb.NewParagraph().Text("Sorry, there was an error processing your request. Please select one of the following options:"))
		cardBody.AddChild(hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlPasswordRestore).Text("request a reset of your password")))
		cardBody.AddChild(hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlLogin).Text("login to the system")))
		cardBody.AddChild(hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlRegister).Text("create a new account")))
	}

	cardFooter := hb.NewDiv().Class("card-footer").AddChildren([]hb.TagInterface{
		buttonLogin,
	})

	if a.enableRegistration {
		cardFooter.AddChild(buttonRegister)
	}

	card := hb.NewDiv().
		Class("card card-default").
		Style("margin:0 auto;max-width: 360px;")

	card.AddChild(cardHeader).AddChild(cardBody).AddChild(cardFooter)

	container := hb.NewDiv().Class("container").Child(card)

	return container.ToHTML()
}

func (a authImplementation) pagePasswordResetScripts() string {
	urlApiPasswordReset := a.LinkApiPasswordReset()
	urlSuccess := a.LinkLogin()

	return `
	var urlApiPasswordReset = "` + urlApiPasswordReset + `";
	var urlOnSuccess = "` + urlSuccess + `";
    /**
     * Raises an error message
     * @param  {String} error
     * @returns  {Boolean}
     */
    function resetFormRaiseError(error) {
        $('div.alert-success').html('').hide();
        $('div.alert-danger').html(error).show();
        setTimeout(function () {
            $('div.alert-danger').html('').hide();
        }, 10000);
        return false;
    }

    function resetFormRaiseSuccess(success) {
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
    function resetFormValidate() {
        var token = $.trim($('input[name=token]').val());
        var password = $.trim($('input[name=password]').val());
        var passwordConfirm = $.trim($('input[name=password_confirm]').val());

        $('.ButtonContinue .imgLoading').show();

        var data = {"password": password, "password_confirm": passwordConfirm, "token": token};

        $.post(urlApiPasswordReset, data).then(function (response) {
            $('.ButtonContinue .imgLoading').hide();

            if (response.status !== "success") {
                return resetFormRaiseError(response.message);
            }

            resetFormRaiseSuccess('Success');
            $('div.alert-danger').html('').hide();
            setTimeout(function () {
                window.location.href=urlOnSuccess;
            }, 2000);
            return;
        }).fail(function (error) {
			console.log(error);
            $('.ButtonContinue .imgLoading').hide();
            return resetFormRaiseError('There was an error. Try again later!');
        });
    }
    $(function () {
        $("input[name=first_name").focus();
    });
	`
}
