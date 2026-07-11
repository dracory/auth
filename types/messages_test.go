package types

import (
	"strings"
	"testing"
)

func TestMessageConstants(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		wantNonEmpty bool
	}{
		{"MsgEmailRequired", MsgEmailRequired, true},
		{"MsgPasswordRequired", MsgPasswordRequired, true},
		{"MsgFirstNameRequired", MsgFirstNameRequired, true},
		{"MsgLastNameRequired", MsgLastNameRequired, true},
		{"MsgTokenRequired", MsgTokenRequired, true},
		{"MsgVerificationCodeRequired", MsgVerificationCodeRequired, true},
		{"MsgPasswordsDoNotMatch", MsgPasswordsDoNotMatch, true},
		{"MsgEmailInvalid", MsgEmailInvalid, true},
		{"MsgVerificationCodeInvalidLength", MsgVerificationCodeInvalidLength, true},
		{"MsgVerificationCodeInvalidCharacters", MsgVerificationCodeInvalidCharacters, true},
		{"MsgVerificationCodeExpired", MsgVerificationCodeExpired, true},
		{"MsgSerializedFormatMalformed", MsgSerializedFormatMalformed, true},
		{"MsgInvalidCredentials", MsgInvalidCredentials, true},
		{"MsgUserNotFound", MsgUserNotFound, true},
		{"MsgRegistrationFailed", MsgRegistrationFailed, true},
		{"MsgRegistrationFailedFn", MsgRegistrationFailedFn, true},
		{"MsgRegistrationFailedEmailTpl", MsgRegistrationFailedEmailTpl, true},
		{"MsgRegistrationFailedEmailSend", MsgRegistrationFailedEmailSend, true},
		{"MsgRegistrationFailedGeneric", MsgRegistrationFailedGeneric, true},
		{"MsgPasswordResetFailed", MsgPasswordResetFailed, true},
		{"MsgLogoutFailed", MsgLogoutFailed, true},
		{"MsgPasswordValidationFailed", MsgPasswordValidationFailed, true},
		{"MsgInternalServer", MsgInternalServer, true},
		{"MsgFailedToProcess", MsgFailedToProcess, true},
		{"MsgFailedToGenerateCode", MsgFailedToGenerateCode, true},
		{"MsgFailedToSendEmail", MsgFailedToSendEmail, true},
		{"MsgLinkNotValidOrExpired", MsgLinkNotValidOrExpired, true},
		{"MsgTooManyRequests", MsgTooManyRequests, true},
		{"EmailSubjectRegistrationCode", EmailSubjectRegistrationCode, true},
		{"MsgLoginSuccess", MsgLoginSuccess, true},
		{"MsgRegistrationSuccess", MsgRegistrationSuccess, true},
		{"MsgRegistrationCodeSent", MsgRegistrationCodeSent, true},
		{"MsgLoginCodeSent", MsgLoginCodeSent, true},
		{"MsgPasswordResetLinkSent", MsgPasswordResetLinkSent, true},
		{"MsgPasswordResetSuccess", MsgPasswordResetSuccess, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantNonEmpty && tt.value == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}

func TestMessageConstantsNoDuplicates(t *testing.T) {
	seen := make(map[string]string)
	all := []struct {
		name, value string
	}{
		{"MsgEmailRequired", MsgEmailRequired},
		{"MsgPasswordRequired", MsgPasswordRequired},
		{"MsgFirstNameRequired", MsgFirstNameRequired},
		{"MsgLastNameRequired", MsgLastNameRequired},
		{"MsgTokenRequired", MsgTokenRequired},
		{"MsgVerificationCodeRequired", MsgVerificationCodeRequired},
		{"MsgPasswordsDoNotMatch", MsgPasswordsDoNotMatch},
		{"MsgEmailInvalid", MsgEmailInvalid},
		{"MsgVerificationCodeInvalidLength", MsgVerificationCodeInvalidLength},
		{"MsgVerificationCodeInvalidCharacters", MsgVerificationCodeInvalidCharacters},
		{"MsgVerificationCodeExpired", MsgVerificationCodeExpired},
		{"MsgSerializedFormatMalformed", MsgSerializedFormatMalformed},
		{"MsgInvalidCredentials", MsgInvalidCredentials},
		{"MsgUserNotFound", MsgUserNotFound},
		{"MsgRegistrationFailed", MsgRegistrationFailed},
		{"MsgRegistrationFailedFn", MsgRegistrationFailedFn},
		{"MsgRegistrationFailedEmailTpl", MsgRegistrationFailedEmailTpl},
		{"MsgRegistrationFailedEmailSend", MsgRegistrationFailedEmailSend},
		{"MsgRegistrationFailedGeneric", MsgRegistrationFailedGeneric},
		{"MsgPasswordResetFailed", MsgPasswordResetFailed},
		{"MsgLogoutFailed", MsgLogoutFailed},
		{"MsgPasswordValidationFailed", MsgPasswordValidationFailed},
		{"MsgInternalServer", MsgInternalServer},
		{"MsgFailedToProcess", MsgFailedToProcess},
		{"MsgFailedToGenerateCode", MsgFailedToGenerateCode},
		{"MsgFailedToSendEmail", MsgFailedToSendEmail},
		{"MsgLinkNotValidOrExpired", MsgLinkNotValidOrExpired},
		{"MsgTooManyRequests", MsgTooManyRequests},
		{"EmailSubjectRegistrationCode", EmailSubjectRegistrationCode},
		{"MsgLoginSuccess", MsgLoginSuccess},
		{"MsgRegistrationSuccess", MsgRegistrationSuccess},
		{"MsgRegistrationCodeSent", MsgRegistrationCodeSent},
		{"MsgLoginCodeSent", MsgLoginCodeSent},
		{"MsgPasswordResetLinkSent", MsgPasswordResetLinkSent},
		{"MsgPasswordResetSuccess", MsgPasswordResetSuccess},
	}

	for _, m := range all {
		if existing, ok := seen[m.value]; ok {
			t.Errorf("duplicate message value %q used by both %s and %s", m.value, existing, m.name)
		}
		seen[m.value] = m.name
	}
}

func TestMessageConstantsAreTrimmed(t *testing.T) {
	all := []struct {
		name, value string
	}{
		{"MsgEmailRequired", MsgEmailRequired},
		{"MsgPasswordRequired", MsgPasswordRequired},
		{"MsgInternalServer", MsgInternalServer},
		{"MsgLoginSuccess", MsgLoginSuccess},
		{"MsgRegistrationCodeSent", MsgRegistrationCodeSent},
	}

	for _, m := range all {
		if strings.TrimSpace(m.value) != m.value {
			t.Errorf("%s has leading or trailing whitespace: %q", m.name, m.value)
		}
	}
}
