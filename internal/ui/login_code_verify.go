package ui

import "github.com/dracory/hb"

// LoginCodeVerifyContent builds the HTML for the login code verification page.
func LoginCodeVerifyContent(urlBack string) string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Login Code Verification").Style("margin:0px;")
	infoParagraph := hb.NewParagraph().Class("text-info").Text("We sent you a login code to your email. Please check your mailbox")
	verificationCodeLabel := hb.NewLabel().Text("Verification code")
	verificationCodeInput := hb.NewInput().Class("form-control").Name("verification_code").Placeholder("Enter verification code")
	verificationCodeFormGroup := hb.NewDiv().Class("form-group mt-3").Child(verificationCodeLabel).AddChild(verificationCodeInput)
	buttonLogin := hb.NewButton().Class("ButtonLogin btn btn-lg btn-success btn-block w-100 text-white").Children([]hb.TagInterface{
		hb.NewI().Class("bi bi-send").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Login"),
		hb.NewDiv().Class("ImgLoading spinner-border spinner-border-sm text-light").Style("display:none;margin-left:10px;"),
	}).OnClick("loginFormValidate()")
	buttonLoginFormGroup := hb.NewDiv().Class("form-group mt-3 mb-3").AddChild(buttonLogin)
	buttonBack := hb.NewHyperlink().Class("btn btn-info text-white float-start").Children([]hb.TagInterface{
		hb.NewI().Class("bi bi-chevron-left").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Resend code"),
	}).Href(urlBack)

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").Child(header)
	cardBody := hb.NewDiv().Class("card-body").Children([]hb.TagInterface{
		alertGroup,
		infoParagraph,
		verificationCodeFormGroup,
		buttonLoginFormGroup,
	})
	cardFooter := hb.NewDiv().Class("card-footer").Children([]hb.TagInterface{
		buttonBack,
	})

	card := hb.NewDiv().Class("card card-default").Style("margin:0 auto;max-width: 360px;").Children([]hb.TagInterface{
		cardHeader,
		cardBody,
		cardFooter,
	})

	container := hb.NewDiv().Class("container").Child(card)

	return container.ToHTML()
}

// LoginCodeVerifyScripts builds the JS for the login code verification page.
func LoginCodeVerifyScripts(urlApiLoginCodeVerify, urlOnSuccess string) string {
	return `
		var urlApiLoginCodeVerify = "` + urlApiLoginCodeVerify + `";
		var urlOnSuccess = "` + urlOnSuccess + `";
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
			var verificationCode = $.trim($('input[name=verification_code]').val());

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			$('.ButtonLogin .ImgLoading').show();

			var data = {"verification_code": verificationCode};

			$.post(urlApiLoginCodeVerify, data).then(function (response) {
				$('.ButtonLogin .ImgLoading').hide();

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				loginFormRaiseSuccess('Verification successful');
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
