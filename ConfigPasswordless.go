package auth

type ConfigPasswordless struct {

	// ===== START: shared by all implementations
	EnableRegistration      bool
	Endpoint                string
	FuncLayout              func(content string) string
	FuncTemporaryKeyGet     func(key string) (value string, err error)
	FuncTemporaryKeySet     func(key string, value string, expiresSeconds int) (err error)
	FuncUserFindByAuthToken func(sessionID string, userIP string, userAgent string) (userID string, err error)
	FuncUserLogout          func(userID string) (err error)
	FuncUserStoreAuthToken  func(sessionID string, userID string, userIP string, userAgent string) error
	UrlRedirectOnSuccess    string
	UseCookies              bool
	UseLocalStorage         bool
	// ===== END: shared by all implementations

	// ===== START: passwordless options
	FuncUserFindByEmail           func(email string) (userID string, err error)
	FuncEmailTemplateLoginCode    func(email string, logingLink string) string   // optional
	FuncEmailTemplateRegisterCode func(email string, registerLink string) string // optional
	FuncEmailSend                 func(email string, emailSubject string, emailBody string) (err error)
	FuncUserRegister              func(email string, firstName string, lastName string) (err error)
	// ===== END: passwordless options
}
