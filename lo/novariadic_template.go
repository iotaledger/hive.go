//go:build ignore

package lo

//go:generate go run github.com/iotaledger/hive.go/codegen/variadic/cmd@latest 0 9 novariadic.go

// NoVariadic{{if hasParams}}{{paramCount}}{{end}} turns a variadic function {{- if hasParams}} with {{paramCount}} additional parameters{{end}} into a non-variadic one (variadic part empty).
func NoVariadic /*{{- if hasParams}}{{paramCount}}{{end -}}*/ [ /*{{- if hasParams}}{{types}}{{", "}}{{end -}}*/ V, R any](f func( /*{{- if hasParams}}{{types}}{{", "}}{{end -}}*/ ...V) R) func( /*{{- types -}}*/ ) R {
	return func( /*{{- typedParams -}}*/ ) R {
		return f( /*{{- params -}}*/ )
	}
}
