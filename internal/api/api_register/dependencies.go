package api_register

import "context"

// Dependencies defines the dependencies required for handling the registration
// API endpoint. It combines both passwordless and username+password flows
// behind a shared interface.
type Dependencies struct {
	// Passwordless indicates whether the registration flow is passwordless.
	Passwordless bool

	// RegisterPasswordlessInitDependencies are used when Passwordless is true.
	RegisterPasswordlessInitDependencies RegisterPasswordlessInitDependencies

	// RegisterWithUsernameAndPassword performs the username+password
	// registration when Passwordless is false. It is responsible for all
	// validation and business rules, and returns a user-facing success or
	// error message.
	RegisterWithUsernameAndPassword func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (successMessage, errorMessage string)
}
