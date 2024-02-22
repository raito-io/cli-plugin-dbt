package array

func Map[E any, T ~[]E, O any](a T, fn func(e E) O)[]O {
	result := make([]O, len(a))
	for i, e := range a {
		result[i] = fn(e)
	}

	return result
}
