{{- /* A string union of an enum's value names. Shared by enums/<name>.ts and sandbox.d.ts. */ -}}
{{define "enum"}}{{.Doc}}{{.Export}}type {{.Name}} =
{{range .Values}}{{.Doc}}  | {{str .Name}}{{if .Last}};{{end}}
{{end}}{{end}}
