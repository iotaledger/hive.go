// Code generated by go generate; DO NOT EDIT.
package lo

// NoVariadic turns a variadic function into a non-variadic one (variadic part empty).
func NoVariadic[V, R any](f func(...V) R) func() R {
	return func() R {
		return f()
	}
}

// NoVariadic1 turns a variadic function with 1 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic1[T1, V, R any](f func(T1, ...V) R) func(T1) R {
	return func(arg1 T1) R {
		return f(arg1)
	}
}

// NoVariadic2 turns a variadic function with 2 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic2[T1, T2, V, R any](f func(T1, T2, ...V) R) func(T1, T2) R {
	return func(arg1 T1, arg2 T2) R {
		return f(arg1, arg2)
	}
}

// NoVariadic3 turns a variadic function with 3 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic3[T1, T2, T3, V, R any](f func(T1, T2, T3, ...V) R) func(T1, T2, T3) R {
	return func(arg1 T1, arg2 T2, arg3 T3) R {
		return f(arg1, arg2, arg3)
	}
}

// NoVariadic4 turns a variadic function with 4 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic4[T1, T2, T3, T4, V, R any](f func(T1, T2, T3, T4, ...V) R) func(T1, T2, T3, T4) R {
	return func(arg1 T1, arg2 T2, arg3 T3, arg4 T4) R {
		return f(arg1, arg2, arg3, arg4)
	}
}

// NoVariadic5 turns a variadic function with 5 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic5[T1, T2, T3, T4, T5, V, R any](f func(T1, T2, T3, T4, T5, ...V) R) func(T1, T2, T3, T4, T5) R {
	return func(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5) R {
		return f(arg1, arg2, arg3, arg4, arg5)
	}
}

// NoVariadic6 turns a variadic function with 6 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic6[T1, T2, T3, T4, T5, T6, V, R any](f func(T1, T2, T3, T4, T5, T6, ...V) R) func(T1, T2, T3, T4, T5, T6) R {
	return func(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6) R {
		return f(arg1, arg2, arg3, arg4, arg5, arg6)
	}
}

// NoVariadic7 turns a variadic function with 7 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic7[T1, T2, T3, T4, T5, T6, T7, V, R any](f func(T1, T2, T3, T4, T5, T6, T7, ...V) R) func(T1, T2, T3, T4, T5, T6, T7) R {
	return func(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6, arg7 T7) R {
		return f(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	}
}

// NoVariadic8 turns a variadic function with 8 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic8[T1, T2, T3, T4, T5, T6, T7, T8, V, R any](f func(T1, T2, T3, T4, T5, T6, T7, T8, ...V) R) func(T1, T2, T3, T4, T5, T6, T7, T8) R {
	return func(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6, arg7 T7, arg8 T8) R {
		return f(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
	}
}

// NoVariadic9 turns a variadic function with 9 additional parameters into a non-variadic one (variadic part empty).
func NoVariadic9[T1, T2, T3, T4, T5, T6, T7, T8, T9, V, R any](f func(T1, T2, T3, T4, T5, T6, T7, T8, T9, ...V) R) func(T1, T2, T3, T4, T5, T6, T7, T8, T9) R {
	return func(arg1 T1, arg2 T2, arg3 T3, arg4 T4, arg5 T5, arg6 T6, arg7 T7, arg8 T8, arg9 T9) R {
		return f(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
	}
}
