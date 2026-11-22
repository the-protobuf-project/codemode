{{- /* A resource name as a template literal type. Shared by resources/<name>.ts and sandbox.d.ts. */ -}}
{{define "resource"}}{{.Doc}}{{.Export}}type {{.Name}} = `{{.Template}}`;
{{end}}
