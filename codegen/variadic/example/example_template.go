//go:build ignore

package example

//go:generate go run ../cmd 0 9 example.go

// PanicOnErr{{if hasParams}}{{paramCount}}{{end}} panics if err is not nil{{if hasParams}} and otherwise returns the remaining {{paramCount}} generic parameters.{{else}}.{{end}}
func PanicOnErr /*{{- if hasParams}}{{paramCount}}{{"["}}{{types}}{{" any]"}}{{end -}}*/ ( /*{{- if hasParams}}{{typedParams}}{{", "}}{{end -}}*/ err error) /*{{- if hasParams}}{{" ("}}{{types}}{{")"}}{{end}}*/ {
	if err != nil {
		panic(err)
	} /*{{- if hasParams}}*/

	return /*{{params}}{{end}}*/
}
