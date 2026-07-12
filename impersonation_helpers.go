package auth

import (
	"net/http"

	"github.com/dracory/auth/types"
)

func IsImpersonating(r *http.Request) bool {
	v := r.Context().Value(types.IsImpersonating{})
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func GetImpersonatorUserID(r *http.Request) string {
	v := r.Context().Value(types.ImpersonatorUserID{})
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
