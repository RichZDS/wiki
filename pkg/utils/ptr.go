package utils

// Ptr returns a pointer to v. Useful for inlining pointer values in struct literals.
func Ptr[T any](v T) *T {
	return &v
}
