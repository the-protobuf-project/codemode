{{.Banner}}

{{template "sandbox-prelude"}}{{range .Types}}
{{.}}{{end}}{{range .Namespaces}}
{{template "namespace" .}}{{end}}
