package shared

// OrDefault returns override when non-empty, otherwise the package's
// built-in copy — the rule that keeps every *PageOptions zero value
// rendering today's pages.
func OrDefault(override, fallback string) string {
	if override == "" {
		return fallback
	}
	return override
}
