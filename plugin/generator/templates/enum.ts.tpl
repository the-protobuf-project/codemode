{{template "enum" .Decl}}
/** Converts {@link {{.Decl.Name}}} between value names and numbers. */
export const {{.Codec}}: EnumCodec<{{.Decl.Name}}> = enumCodec({{str .TypeName}}, [
{{range .Decl.Values}}  [{{str .Name}}, {{.Number}}],
{{end}}]);
