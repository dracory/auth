package utils

// LoginCodeLength returns the length of verification/login codes.
// extraHardened should be false in normal operation; it is only set to true
// when rate limiting is explicitly disabled (which should never happen in
// production). In that hardened mode it returns a longer length to
// significantly increase the search space; together with the hardened gamma
// this yields an astronomically large space.
func LoginCodeLength(extraHardened bool) int {
	if extraHardened {
		// Hardened length used together with the expanded gamma below.
		return 12
	}

	// Default length used by the library today.
	return 8
}

// LoginCodeGamma returns the character set (gamma) used for verification/
// login codes. extraHardened should be false in normal operation; it is only
// set to true when rate limiting is explicitly disabled (again, not
// recommended for production). In that hardened mode it returns a much
// larger alphabet to drastically increase entropy; combined with the
// hardened length, this makes brute-force attacks negligible even without
// rate limiting.
func LoginCodeGamma(extraHardened bool) string {
	if extraHardened {
		// Hardened gamma: 43 characters (21 consonant uppers + 20 consonant
		// lowers + digits). This matches the recommendation in the docs.
		return "bcdfghjklmpqrstvxyzBCDFGHJKLMNPQRSTVXYZ0123456789"
	}

	// Default gamma from consts.go.
	return "BCDFGHJKLMNPQRSTVXYZ"
}
