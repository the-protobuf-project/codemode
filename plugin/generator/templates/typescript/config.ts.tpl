export const config: CodemodeConfig = {
  discovery: {{str .Discovery}},
  executor: {
{{range .Executor}}    {{.Key}}: {{.Expr}},
{{end}}    outbound: {{str .Outbound}},
    allow: [{{.Allow}}],
  },
  runtime: {
{{range .Runtime}}    {{.Key}}: {{.Expr}},
{{end}}  },
  mcp: { expose: {{str .Expose}} },
};
