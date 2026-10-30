package types

// UnwrapAuth resolves decorator wrappers (for example the multi-method
// alternatives decorator) back to the underlying auth implementation.
// Without it, capability assertions such as AuthPasswordInterface would
// fail on decorated impls even though the wrapped impl satisfies them.
func UnwrapAuth(a AuthSharedInterface) AuthSharedInterface {
	for {
		u, ok := a.(interface{ Unwrap() AuthSharedInterface })
		if !ok {
			return a
		}
		a = u.Unwrap()
	}
}
