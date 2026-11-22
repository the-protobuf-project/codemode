{{range .Exports}}export * from {{str .}};
{{end}}{{if .Codecs}}
registerTypes([{{join .Codecs ", "}}]);
{{end}}
