package auth

import (
	"net/http"

	"log/slog"

	"github.com/dracory/hb"
)

func (a Auth) pagePasswordRestore(w http.ResponseWriter, r *http.Request) {
	webpage := webpage("Restore Password", a.pagePasswordRestoreContent(), a.pagePasswordRestoreScripts())
	logger := a.logger
	if logger == nil {
		logger = slog.Default()
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write password restore page response", "error", err)
	}
}

func (a Auth) pagePasswordRestoreContent() string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().
		Style("margin:0px;").
		HTML("Restore password")
	firstNameLabel := hb.NewLabel().Text("First Name")
	firstNameInput := hb.NewInput().Class("form-control").Name("first_name").Placeholder("Enter first name")
	firstNameFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(firstNameLabel).AddChild(firstNameInput)
	lastNameLabel := hb.NewLabel().Text("Last Name")
	lastNameInput := hb.NewInput().Class("form-control").Name("last_name").Placeholder("Enter last name")
	lastNameFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(lastNameLabel).AddChild(lastNameInput)
	emailLabel := hb.NewLabel().Text("E-mail Address")
	emailInput := hb.NewInput().Class("form-control").Name("email").Placeholder("Enter e-mail address")
	emailFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(emailLabel).AddChild(emailInput)

	buttonContinue := hb.NewButton().
		Class("ButtonContinue btn btn-lg btn-success btn-block w-100").
		OnClick("passwordRestoreFormValidate()").
		HTML("Send Password Reset Link")

	buttonContinueFormGroup := hb.NewDiv().
		Class("form-group mt-3 mb-3").
		Child(buttonContinue)

	buttonLogin := hb.NewHyperlink().
		Class("btn btn-info float-start").
		Href(a.LinkLogin()).
		HTML("Login")

	buttonRegister := hb.NewHyperlink().
		Class("btn btn-warning float-end").
		Href(a.LinkRegister()).
		HTML("Register")

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").AddChild(header)
	cardBody := hb.NewDiv().Class("card-body").AddChildren([]hb.TagInterface{
		alertGroup,
		firstNameFormGroup,
		lastNameFormGroup,
		emailFormGroup,
		buttonContinueFormGroup,
	})
	cardFooter := hb.NewDiv().Class("card-footer").AddChildren([]hb.TagInterface{
		buttonLogin,
	})

	if a.enableRegistration {
		cardFooter.AddChild(buttonRegister)
	}

	card := hb.NewDiv().Class("card card-default").Style("margin:0 auto;max-width: 360px;")
	card.AddChild(cardHeader).AddChild(cardBody).AddChild(cardFooter)

	container := hb.NewDiv().Class("container").
		AddChild(card)

	return container.ToHTML()
}

func (a Auth) pagePasswordRestoreScripts() string {
	urlApiPasswordrestore := a.LinkApiPasswordRestore()
	urlSuccess := a.LinkLogin()

	return `
	var urlApiPasswordRestore = "` + urlApiPasswordrestore + `";
	var urlOnSuccess = "` + urlSuccess + `";

    /**
     * Raises an error message
     * @param  {String} error
     * @returns  {Boolean}
     */
    function passwordRestoreFormRaiseError(error) {
        $('div.alert-success').html('').hide();
        $('div.alert-danger').html(error).show();
        setTimeout(function () {
            $('div.alert-danger').html('').hide();
        }, 10000);
        return false;
    }

    function passwordRestoreFormRaiseSuccess(success) {
        $('div.alert-danger').html('').hide();
        $('div.alert-success').html(success).show();
        setTimeout(function () {
            $('div.alert-success').html('').hide();
        }, 10000);
        return false;
    }

    /**
     * Validates the Password Restore Form
	 * and sends the request to the backend
     * @returns  {Boolean}
     */
    function passwordRestoreFormValidate() {
		var first_name = $.trim($('input[name=first_name]').val());
		var last_name = $.trim($('input[name=last_name]').val());
        var email = $.trim($('input[name=email]').val());

        $('.ButtonContinue .imgLoading').show();

        var data = {"first_name": first_name, "last_name": last_name, "email": email};

        $.post(urlApiPasswordRestore, data).then(function (response) {
            $('.ButtonContinue .imgLoading').hide();

            if (response.status !== "success") {
                return passwordRestoreFormRaiseError(response.message);
            }

            passwordRestoreFormRaiseSuccess('Success');
            $('div.alert-danger').html('').hide();

            setTimeout(function () {
                window.location.href=urlOnSuccess;
            }, 2000);
            return;
        }).fail(function (error) {
			console.log(error);
            $('.ButtonContinue .imgLoading').hide();
            return passwordRestoreFormRaiseError('There was an error. Try again later!');
        });
    }
    $(function () {
        $("input[name=first_name").focus();
    });
	`
}
