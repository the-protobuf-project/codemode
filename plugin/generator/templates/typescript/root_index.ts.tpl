{{range .Packages}}export * as {{.Alias}} from {{str .Path}};
{{end}}export * from {{str .Codemode}};
