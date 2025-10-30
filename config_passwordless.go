package auth

type ConfigPasswordless struct {

	// ===== START: shared by all implementations
	EnableRegistration      bool
	Endpoint                string
	FuncLayout              func(content string) string
	FuncTemporaryKeyGet     func(key string) (value string, err error)
	FuncTemporaryKeySet     func(key string, value string, expiresSeconds int) (err error)
	FuncUserFindByAuthToken func(sessionID string, options UserAuthOptions) (userID string, err error)
	FuncUserLogout          func(userID string, options UserAuthOptions) (err error)
	FuncUserStoreAuthToken  func(sessionID string, userID string, options UserAuthOptions) error
	UrlRedirectOnSuccess    string
	UseCookies              bool
	UseLocalStorage         bool
	// ===== END: shared by all implementations

	// ===== START: passwordless options
	FuncUserFindByEmail           func(email string, options UserAuthOptions) (userID string, err error)
	FuncEmailTemplateLoginCode    func(email string, logingLink string, options UserAuthOptions) string   // optional
	FuncEmailTemplateRegisterCode func(email string, registerLink string, options UserAuthOptions) string // optional
	FuncEmailSend                 func(email string, emailSubject string, emailBody string) (err error)
	FuncUserRegister              func(email string, firstName string, lastName string, options UserAuthOptions) (err error)
	// ===== END: passwordless options
}
