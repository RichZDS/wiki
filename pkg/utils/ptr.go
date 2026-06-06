package utils

// Ptr returns a pointer to v. Useful for inlining pointer values in struct literals.
// Ptr 返回给定值的指针。
func Ptr[T any](v T) *T {
	return &v
}
