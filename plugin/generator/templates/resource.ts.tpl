{{template "resource" .Decl}}
/** Formats, parses and checks {@link {{.Decl.Name}}} values. */
export const {{.Decl.Name}}: ResourceName<{{.Vars}}, {{.Decl.Name}}> = resourceName({{str .Type}}, {{str .Pattern}});
