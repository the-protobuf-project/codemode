{{.Doc}}export interface {{.Name}} {
{{range $i, $m := .Methods}}{{if $i}}
{{end}}{{$m.Doc}}  {{$m.Client}}(request: {{$m.Input}}, options?: CallOptions): Promise<{{$m.Output}}>;
{{end}}}

/** Describes each {@link {{.Name}}} RPC for transports. */
export const {{.Name}}Descriptor = {
  typeName: {{str .TypeName}},
  methods: {
{{range .Methods}}    {{.Client}}: {
      service: {{str $.TypeName}},
      name: {{str .RPC}},
      input: {{.InputCodec}},
      output: {{.OutputCodec}},
{{if .HTTP}}      http: {{.HTTP}},
{{end}}{{if .Idempotency}}      idempotency: {{str .Idempotency}},
{{end}}    } satisfies MethodDescriptor<{{.Input}}, {{.Output}}>,
{{end}}  },
} as const;
