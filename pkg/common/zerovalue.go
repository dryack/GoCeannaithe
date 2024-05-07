package common

// ZeroValue returns the zero value of a given type T
func ZeroValue[T any]() T {
	var zero T
	return zero
}
