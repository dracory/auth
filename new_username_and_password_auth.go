package auth

import "errors"

func NewUsernameAndPasswordAuth(config ConfigUsernameAndPassword) (*Auth, error) {
	if config.Endpoint == "" {
		return nil, errors.New("auth: endpoint is required")
	}

	if config.UrlRedirectOnSuccess == "" {
		return nil, errors.New("auth: url to redirect to on success is required")
	}

	if config.FuncTemporaryKeyGet == nil {
		return nil, errors.New("auth: FuncTemporaryKeyGet function is required")
	}

	if config.FuncTemporaryKeySet == nil {
		return nil, errors.New("auth: FuncTemporaryKeySet function is required")
	}

	if config.FuncUserFindByAuthToken == nil {
		return nil, errors.New("auth: FuncUserFindByAuthToken function is required")
	}

	if config.FuncUserFindByUsername == nil {
		return nil, errors.New("auth: FuncUserFindByUsername function is required")
	}

	if config.FuncUserLogin == nil {
		return nil, errors.New("auth: FuncUserLogin function is required")
	}

	if config.FuncUserLogout == nil {
		return nil, errors.New("auth: FuncUserLogout function is required")
	}

	if config.EnableRegistration && config.FuncUserRegister == nil {
		return nil, errors.New("auth: FuncUserRegister function is required")
	}

	if config.FuncUserStoreAuthToken == nil {
		return nil, errors.New("auth: FuncUserStoreToken function is required")
	}

	if config.FuncEmailSend == nil {
		return nil, errors.New("auth: FuncEmailSend function is required")
	}

	if config.UseCookies && config.UseLocalStorage {
		return nil, errors.New("auth: UseCookies and UseLocalStorage cannot be both true")
	}

	if !config.UseCookies && !config.UseLocalStorage {
		return nil, errors.New("auth: UseCookies and UseLocalStorage cannot be both false")
	}

	auth := &Auth{}
	auth.enableRegistration = config.EnableRegistration
	auth.enableVerification = config.EnableVerification
	auth.endpoint = config.Endpoint
	auth.passwordless = false
	auth.urlRedirectOnSuccess = config.UrlRedirectOnSuccess
	auth.useCookies = config.UseCookies
	auth.useLocalStorage = config.UseLocalStorage
	auth.funcEmailSend = config.FuncEmailSend
	auth.funcEmailTemplatePasswordRestore = config.FuncEmailTemplatePasswordRestore
	auth.funcLayout = config.FuncLayout
	auth.funcTemporaryKeyGet = config.FuncTemporaryKeyGet
	auth.funcTemporaryKeySet = config.FuncTemporaryKeySet
	auth.funcUserLogin = config.FuncUserLogin
	auth.funcUserLogout = config.FuncUserLogout
	auth.funcUserPasswordChange = config.FuncUserPasswordChange
	auth.funcUserRegister = config.FuncUserRegister
	auth.funcUserFindByAuthToken = config.FuncUserFindByAuthToken
	auth.funcUserFindByUsername = config.FuncUserFindByUsername
	auth.funcUserStoreAuthToken = config.FuncUserStoreAuthToken

	// If no user defined layout is set, use default
	if auth.funcLayout == nil {
		auth.funcLayout = auth.layout
	}

	// If no user defined email template is set, use default
	if auth.funcEmailTemplatePasswordRestore == nil {
		auth.funcEmailTemplatePasswordRestore = emailTemplatePasswordChange
	}

	// If no user defined email template is set, use default
	if auth.funcEmailTemplateRegisterCode == nil {
		auth.funcEmailTemplateRegisterCode = emailRegisterCodeTemplate
	}

	return auth, nil
}
