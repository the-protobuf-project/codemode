{{- /* A message as an interface. Shared by messages/<name>.ts and sandbox.d.ts. */ -}}
{{define "interface"}}{{.Doc}}{{if not .Fields}}{{.Export}}type {{.Name}} = Record<string, never>;
{{else}}{{.Export}}interface {{.Name}} {
{{range $i, $f := .Fields}}{{if $i}}
{{end}}{{$f.Doc}}  {{$f.Key}}{{$f.Opt}}: {{$f.Type}};
{{end}}}
{{end}}{{end}}
