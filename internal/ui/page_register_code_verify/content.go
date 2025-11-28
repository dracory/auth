package page_register_code_verify

import "github.com/dracory/hb"

// RegisterCodeVerifyContent builds the HTML for the registration code verification page.
func RegisterCodeVerifyContent(urlBack string) string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Registration Code Verification").Style("margin:0px;")

	verificationCodeLabel := hb.NewLabel().Text("Verification code")
	verificationCodeInput := hb.NewInput().
		Class("form-control").
		Name("verification_code").
		Placeholder("Enter verification code")
	verificationCodeFormGroup := hb.NewDiv().
		Class("form-group mt-3").
		Child(verificationCodeLabel).
		Child(verificationCodeInput)
	buttonVerify := hb.NewButton().
		Class("btn btn-lg btn-success btn-block w-100").
		Children([]hb.TagInterface{
			hb.NewI().Class("bi bi-person-circle").Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().Text("Verify Registration"),
		}).
		OnClick("registerCodeFormValidate()")
	buttonVerifyFormGroup := hb.NewDiv().
		Class("form-group mt-3 mb-3").
		AddChild(buttonVerify)
	buttonBack := hb.NewButton().Class("btn btn-info text-white float-start").Children([]hb.TagInterface{
		hb.NewI().Class("bi bi-chevron-left").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Resend code"),
	}).Href(urlBack)

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").Child(header)
	cardBody := hb.NewDiv().Class("card-body").Children([]hb.TagInterface{
		alertGroup,
		verificationCodeFormGroup,
		buttonVerifyFormGroup,
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

// RegisterCodeVerifyScripts builds the JS for the registration code verification page.
func RegisterCodeVerifyScripts(urlApiRegisterCodeVerify, urlOnSuccess string) string {
	return `
		var urlApiRegisterCodeVerify = "` + urlApiRegisterCodeVerify + `";
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
		function registerCodeFormValidate() {
			var verificationCode = $.trim($('input[name=verification_code]').val());

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			$('.buttonLogin .imgLoading').show();

			var data = {"verification_code": verificationCode};

			$.post(urlApiRegisterCodeVerify, data).then(function (response) {
				$('.buttonLogin .imgLoading').hide();

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
				$('.buttonLogin .imgLoading').hide();
				return loginFormRaiseError('There was an error. Try again later!');
			});
		}
		$(function () {
			$("#email").focus();
		});
	`
}
