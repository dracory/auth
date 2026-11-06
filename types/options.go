package types

// Per-page customization structs for the built-in auth pages. Every
// field falls back to the package's built-in copy when empty, so the
// zero value renders exactly today's pages. The structs are named
// "options" rather than "text" so non-copy knobs (icons, extra links,
// card CSS class) can join later without renaming. Brand stays global
// (ConfigShared.BrandName/BrandTagline/BrandLogoURL) — these types do
// not duplicate it.

// LoginPageOptions customizes the built-in password login page.
type LoginPageOptions struct {
	Title               string // default "Login to your account"
	Subtitle            string // default "Enter your email and password to sign in."
	EmailLabel          string // default "Email"
	EmailPlaceholder    string // default "you@example.com"
	PasswordLabel       string // default "Password"
	PasswordPlaceholder string // default "Enter your password"
	RememberMeLabel     string // default "Remember me on this device"
	SubmitLabel         string // default "Sign in"
	ForgotPasswordLabel string // default "Forgot password?"
	FooterPrompt        string // default "Don't have an account?"
	FooterLinkLabel     string // default "Register"
}

// OtpPageOptions customizes the built-in OTP login page (email -> code).
type OtpPageOptions struct {
	Title            string // default "Login to your account"
	Subtitle         string // default "No password needed. We'll email you a code."
	EmailLabel       string // default "Email"
	EmailPlaceholder string // default "you@example.com"
	RememberMeLabel  string // default "Remember me on this device"
	SubmitLabel      string // default "Send me a login code"
	FooterPrompt     string // default "Don't have an account?"
	FooterLinkLabel  string // default "Register"
}

// RegisterPageOptions customizes the built-in register pages. The four
// register variants (otp form, username/password form, verified-email
// form, authenticated completion) share this type; each variant keeps
// its own fallback strings, so e.g. the completion page still defaults
// to "Complete your registration" / "Save & continue" when Title and
// SubmitLabel are empty.
type RegisterPageOptions struct {
	Title                string // default "Create your account" / "Complete your registration" (completion)
	Subtitle             string // default per variant
	FirstNameLabel       string // default "First name"
	FirstNamePlaceholder string // default "Your first name"
	LastNameLabel        string // default "Last name"
	LastNamePlaceholder  string // default "Your last name"
	EmailLabel           string // default "Email" / "Email (verified)" (verified variants)
	EmailPlaceholder     string // default "you@example.com"
	PasswordLabel        string // default "Password" (username/password variant only)
	PasswordPlaceholder  string // default "Min 8 characters"
	SubmitLabel          string // default "Create account" / "Save & continue" (completion)
	FooterPrompt         string // default "Already have an account?"
	FooterLinkLabel      string // default "Login"
	ForgotPasswordLabel  string // default "Forgot password?" (username/password variant only)
}

// CodeVerifyPageOptions customizes the code-verification pages. Login
// and register variants occupy separate slots on PageOptions so their
// subtitles can differ. CodeLabel applies only to the register variant —
// the login variant uses segmented digit boxes with no visible label.
type CodeVerifyPageOptions struct {
	Title        string // default "Check your email"
	Subtitle     string // default per variant
	CodeLabel    string // default "Verification code" (register variant only)
	SubmitLabel  string // default "Verify code" / "Verify registration" (register)
	ResendPrompt string // default "Didn't get a code?"
	ResendLabel  string // default "Resend"
}

// MagicLinkPageOptions customizes the built-in magic-link login page.
type MagicLinkPageOptions struct {
	Title            string // default "Login to your account"
	Subtitle         string // default "We'll email you a secure link to log in instantly — no password needed."
	EmailLabel       string // default "Email"
	EmailPlaceholder string // default "you@example.com"
	SubmitLabel      string // default "Send me a quick login link"
	FooterPrompt     string // default "Don't have an account?"
	FooterLinkLabel  string // default "Register"
}

// PasswordRestorePageOptions customizes the password restore page.
type PasswordRestorePageOptions struct {
	Title                string // default "Forgot your password?"
	Subtitle             string // default "Enter your email and we'll send you a reset link."
	FirstNameLabel       string // default "First name"
	FirstNamePlaceholder string // default "Your first name"
	LastNameLabel        string // default "Last name"
	LastNamePlaceholder  string // default "Your last name"
	EmailLabel           string // default "Email"
	EmailPlaceholder     string // default "you@example.com"
	SubmitLabel          string // default "Send reset link"
	FooterPrompt         string // default "Remembered it?"
	FooterLinkLabel      string // default "Login"
	RegisterPrompt       string // default "Don't have an account?"
	RegisterLinkLabel    string // default "Register"
}

// PasswordResetPageOptions customizes the password reset page, including
// the error branch shown when the token link is invalid/expired.
type PasswordResetPageOptions struct {
	Title                      string // default "Set a new password"
	Subtitle                   string // default "Choose a strong password for your account."
	NewPasswordLabel           string // default "New password"
	NewPasswordPlaceholder     string // default "Min 8 characters"
	ConfirmPasswordLabel       string // default "Confirm new password"
	ConfirmPasswordPlaceholder string // default "Repeat password"
	SubmitLabel                string // default "Reset password"
	ErrorText                  string // default "Sorry, there was an error processing your request. Please select one of the following options:"
	ErrorRestoreLabel          string // default "request a reset of your password"
	ErrorLoginLabel            string // default "login to the system"
	ErrorRegisterLabel         string // default "create a new account"
	FooterPrompt               string // default "Remembered it?"
	FooterLinkLabel            string // default "Login"
	RegisterPrompt             string // default "Don't have an account?"
	RegisterLinkLabel          string // default "Register"
}

// LogoutPageOptions customizes the logout confirmation page.
type LogoutPageOptions struct {
	Title       string // default "Sign out"
	Subtitle    string // default "Are you sure you want to sign out of your account?"
	SubmitLabel string // default "Logout"
}

// PageOptions aggregates the per-page customization structs — one field
// per built-in page. Empty fields render the package's built-in copy.
type PageOptions struct {
	Login              LoginPageOptions
	LoginOtp           OtpPageOptions
	LoginCodeVerify    CodeVerifyPageOptions
	Register           RegisterPageOptions
	RegisterCodeVerify CodeVerifyPageOptions
	MagicLink          MagicLinkPageOptions
	PasswordRestore    PasswordRestorePageOptions
	PasswordReset      PasswordResetPageOptions
	Logout             LogoutPageOptions
}
