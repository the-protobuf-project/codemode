export const catalog: Catalog = {
  prelude: {{tpl .Prelude}},
  types: {
{{range .Types}}    {{.Key}}: {{tpl .Expr}},
{{end}}  },
  namespaces: [
{{range .Namespaces}}    {
      name: {{str .Name}},
      service: {{str .Service}},
      instructions: {{str .Instructions}},
      declaration: {{tpl .Decl}},
      methods: [{{.Methods}}],
    },
{{end}}  ],
  methods: [
{{range .Methods}}    {
      path: {{str .Path}},
      namespace: {{str .Namespace}},
      name: {{str .Name}},
      rpc: {{str .RPC}},
      description: {{str .Description}},
      requiresApproval: {{.Approval}},
      replay: {{str .Replay}},
      readOnly: {{.ReadOnly}},
{{if .RevertedBy}}      revertedBy: {{str .RevertedBy}},
{{end}}{{if .Streaming}}      streaming: {{str .Streaming}},
{{end}}      declaration: {{tpl .Decl}},
      types: [{{.Types}}],
      examples: [{{.Examples}}],
    },
{{end}}  ],
  snippets: [
{{range .Snippets}}    {
      name: {{str .Name}},
      description: {{str .Description}},
      code: {{tpl .Code}},
{{if .InputType}}      inputType: {{str .InputType}},
{{end}}    },
{{end}}  ],
};
