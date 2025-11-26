package auth

import "errors"

func NewPasswordlessAuth(config ConfigPasswordless) (*Auth, error) {
	auth := &Auth{}

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

	if config.FuncUserFindByEmail == nil {
		return nil, errors.New("auth: FuncUserFindByEmail function is required")
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

	if config.FuncLayout == nil {
		config.FuncLayout = auth.layout
	}

	auth.enableRegistration = config.EnableRegistration
	auth.endpoint = config.Endpoint
	auth.passwordless = true
	auth.urlRedirectOnSuccess = config.UrlRedirectOnSuccess
	auth.useCookies = config.UseCookies
	auth.useLocalStorage = config.UseLocalStorage
	auth.funcLayout = config.FuncLayout
	auth.funcTemporaryKeyGet = config.FuncTemporaryKeyGet
	auth.funcTemporaryKeySet = config.FuncTemporaryKeySet
	auth.funcUserLogout = config.FuncUserLogout
	auth.funcUserFindByAuthToken = config.FuncUserFindByAuthToken
	auth.funcUserStoreAuthToken = config.FuncUserStoreAuthToken
	auth.passwordlessFuncEmailTemplateLoginCode = config.FuncEmailTemplateLoginCode
	// auth.passwordlessFuncEmailTemplateRegisterCode = config.FuncEmailTemplateRegisterCode
	auth.passwordlessFuncEmailSend = config.FuncEmailSend
	auth.passwordlessFuncUserFindByEmail = config.FuncUserFindByEmail
	auth.passwordlessFuncUserRegister = config.FuncUserRegister

	// If no user defined email template is set, use default
	if auth.passwordlessFuncEmailTemplateLoginCode == nil {
		auth.passwordlessFuncEmailTemplateLoginCode = emailLoginCodeTemplate
	}

	// If no user defined email template is set, use default
	if auth.passwordlessFuncEmailTemplateRegisterCode == nil {
		auth.passwordlessFuncEmailTemplateRegisterCode = emailRegisterCodeTemplate
	}

	return auth, nil
}
