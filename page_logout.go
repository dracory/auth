package auth

import (
	"net/http"

	"github.com/dracory/hb"
)

func (a Auth) pageLogout(w http.ResponseWriter, r *http.Request) {
	webpage := webpage("Logout", a.funcLayout(a.pageLogoutContent()), a.pageLogoutScripts())
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write logout page response", "error", err)
	}
}

func (a Auth) pageLogoutContent() string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Sign out").Style("margin:0px;")
	buttonContinue := hb.NewButton().Class("btn btn-lg btn-success btn-block w-100").Text("Logout").OnClick("logoutFormValidate()")
	buttonContinueFormGroup := hb.NewDiv().Class("form-group mt-3").AddChild(buttonContinue)

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").AddChild(header)
	cardBody := hb.NewDiv().Class("card-body").AddChildren([]hb.TagInterface{
		alertGroup,
		buttonContinueFormGroup,
	})

	card := hb.NewDiv().Class("card card-default").
		Style("margin:0 auto;max-width: 360px;").
		AddChild(cardHeader).
		AddChild(cardBody)

	container := hb.NewDiv().Class("container").AddChild(card)

	return container.ToHTML()
}

func (a Auth) pageLogoutScripts() string {
	urlApiLogout := a.LinkApiLogout()
	urlSuccess := a.LinkLogin()
	logger := a.GetLogger()
	logger.Debug("logout page initialized",
		"api_logout_url", urlApiLogout,
	)

	return `
	var urlApiLogout = "` + urlApiLogout + `";
	var urlOnSuccess = "` + urlSuccess + `";
    /**
     * Raises an error message
     * @param  {String} error
     * @returns  {Boolean}
     */
    function logoutFormRaiseError(error) {
        $('div.alert-success').html('').hide();
        $('div.alert-danger').html(error).show();
        setTimeout(function () {
            $('div.alert-danger').html('').hide();
        }, 10000);
        return false;
    }

    function logoutFormRaiseSuccess(success) {
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
    function logoutFormValidate() {
        $('.buttonLogin .imgLoading').show();

        var data = {};

        $.post(urlApiLogout, data).then(function (response) {
            $('.buttonLogin .imgLoading').hide();

            if (response.status !== "success") {
                return logoutFormRaiseError(response.message);
            }

            $$.setAuthToken(response.data.token);
            $$.setAuthUser(response.data.user);
            logoutFormRaiseSuccess('Success');
            $('div.alert-danger').html('').hide();
            setTimeout(function () {
                $$.to(urlOnSuccess);
            }, 2000);
            return;
        }).fail(function (error) {
			console.log(error);
            $('.buttonLogin .imgLoading').hide();
            return logoutFormRaiseError('There was an error. Try again later!');
        });
    }
    $(function () {
        $("#email").focus();
    });
	`
}
