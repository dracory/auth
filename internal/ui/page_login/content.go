package page_login

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// LoginExtras carries optional branding, sibling-method links and the
// site-level legal links for the built-in login pages.
type LoginExtras struct {
	BrandName    string
	BrandTagline string
	BrandLogoURL string
	Alternatives []types.LoginAlternative
	// RememberMe renders the opt-in "remember me" checkbox; the login
	// scripts post remember=true when it is checked.
	RememberMe bool
	// Legal renders the "By continuing, you agree to …" line inside the
	// card when Terms/Privacy URLs are configured.
	Legal types.LegalLinks
}

// rememberBlock renders the opt-in remember-me checkbox inside the login
// card (empty when the feature is off).
func rememberBlock(x LoginExtras, label string) string {
	if !x.RememberMe {
		return ""
	}
	return hb.NewDiv().Class("form-check mb-3").Children([]hb.TagInterface{
		hb.NewInput().Class("form-check-input").Type("checkbox").
			Name("remember").ID("remember").Attr("value", "true"),
		hb.NewLabel().Class("form-check-label small text-muted").
			Attr("for", "remember").Text(shared.OrDefault(label, "Remember me on this device")),
	}).ToHTML()
}

// brandBlock renders the optional brand header above the login card.
func brandBlock(x LoginExtras) string {
	if x.BrandName == "" && x.BrandLogoURL == "" {
		return ""
	}
	logo := ""
	if x.BrandLogoURL != "" {
		logo = hb.NewImage().
			Src(x.BrandLogoURL).
			Style("width:56px;height:56px;object-fit:contain;margin-bottom:10px;border-radius:14px").
			ToHTML()
	}
	name := ""
	if x.BrandName != "" {
		name = hb.NewDiv().
			Text(x.BrandName).
			Style("font-size:1.45rem;font-weight:800;letter-spacing:.04em").ToHTML()
	}
	tagline := ""
	if x.BrandTagline != "" {
		tagline = hb.NewDiv().
			Text(x.BrandTagline).
			Style("font-size:.72rem;font-weight:700;letter-spacing:.22em;opacity:.65").ToHTML()
	}
	return hb.NewDiv().
		Class("auth-brand").
		Style("text-align:center;margin:0 auto 22px;max-width:400px").
		HTML(logo + name + tagline).ToHTML()
}

// alternativesBlock renders an "or" divider with one outlined button per
// sibling login method.
func alternativesBlock(x LoginExtras) string {
	if len(x.Alternatives) == 0 {
		return ""
	}
	btns := ""
	for _, alt := range x.Alternatives {
		btns += shared.AltLink(alt.URL, alt.Icon, alt.Label).ToHTML()
	}
	return shared.OrDivider().ToHTML() + btns
}

// LoginOtpContent builds the HTML content for the otp login page.
func LoginOtpContent(enableRegistration bool, urlRegister string) string {
	return LoginOtpContentX(enableRegistration, urlRegister, LoginExtras{}, types.OtpPageOptions{})
}

// LoginOtpContentX builds the otp login page with extras and per-page
// copy customization (zero-value options render the built-in copy).
func LoginOtpContentX(enableRegistration bool, urlRegister string, x LoginExtras, opts types.OtpPageOptions) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder(shared.OrDefault(opts.EmailPlaceholder, "you@example.com")).
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Login to your account"),
			shared.OrDefault(opts.Subtitle, "No password needed. We'll email you a code.")),
		shared.AlertGroup(),
		shared.FieldGroup(shared.OrDefault(opts.EmailLabel, "Email"), emailInput),
		hb.NewHTML(rememberBlock(x, opts.RememberMeLabel)),
		shared.PrimaryButton("ButtonLogin", "bi-send", shared.OrDefault(opts.SubmitLabel, "Send me a login code"), "loginFormValidate()"),
		hb.NewHTML(alternativesBlock(x)),
		hb.NewHTML(shared.LegalLine(x.Legal)),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		hb.NewHTML(brandBlock(x)),
		card,
	})

	if enableRegistration {
		container.Child(shared.FooterText(
			shared.OrDefault(opts.FooterPrompt, "Don't have an account?"),
			urlRegister,
			shared.OrDefault(opts.FooterLinkLabel, "Register")))
	}

	return container.ToHTML()
}

// LoginOtpScripts builds the JavaScript for the otp login page.
//
//go:embed login_otp.js
var loginOtpJS string

func LoginOtpScripts(urlApiLogin, urlSuccess string) string {
	s := loginOtpJS
	s = strings.ReplaceAll(s, "__URL_API_LOGIN__", urlApiLogin)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlSuccess)
	return s
}

// LoginContent builds the HTML content for the standard login page.
func LoginContent(enableRegistration bool, urlRegister, urlPasswordRestore string) string {
	return LoginContentX(enableRegistration, urlRegister, urlPasswordRestore, LoginExtras{}, types.LoginPageOptions{})
}

// LoginContentX builds the standard login page with extras and per-page
// copy customization (zero-value options render the built-in copy).
func LoginContentX(enableRegistration bool, urlRegister, urlPasswordRestore string, x LoginExtras, opts types.LoginPageOptions) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder(shared.OrDefault(opts.EmailPlaceholder, "you@example.com")).
		Attr("autocomplete", "username").
		Attr("required", "required")
	passwordInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("password").
		Type("password").
		Placeholder(shared.OrDefault(opts.PasswordPlaceholder, "Enter your password")).
		Attr("autocomplete", "current-password").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Login to your account"),
			shared.OrDefault(opts.Subtitle, "Enter your email and password to sign in.")),
		shared.AlertGroup(),
		shared.FieldGroup(shared.OrDefault(opts.EmailLabel, "Email"), emailInput),
		shared.FieldGroup(shared.OrDefault(opts.PasswordLabel, "Password"), passwordInput),
		hb.NewHTML(rememberBlock(x, opts.RememberMeLabel)),
		shared.PrimaryButton("ButtonLogin", "bi-box-arrow-in-right", shared.OrDefault(opts.SubmitLabel, "Sign in"), "loginFormValidate()"),
		hb.NewDiv().Class("text-center small mt-3").Child(
			hb.NewHyperlink().Href(urlPasswordRestore).Text(shared.OrDefault(opts.ForgotPasswordLabel, "Forgot password?"))),
		hb.NewHTML(alternativesBlock(x)),
		hb.NewHTML(shared.LegalLine(x.Legal)),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		hb.NewHTML(brandBlock(x)),
		card,
	})

	if enableRegistration {
		container.Child(shared.FooterText(
			shared.OrDefault(opts.FooterPrompt, "Don't have an account?"),
			urlRegister,
			shared.OrDefault(opts.FooterLinkLabel, "Register")))
	}

	return container.ToHTML()
}

// LoginScripts builds the JavaScript for the standard login page.
//
//go:embed login_password.js
var loginPasswordJS string

func LoginScripts(urlApiLogin, urlOnSuccess string) string {
	s := loginPasswordJS
	s = strings.ReplaceAll(s, "__URL_API_LOGIN__", urlApiLogin)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}
