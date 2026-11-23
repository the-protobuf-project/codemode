{{.Doc}}export interface {{.Name}} {
{{range $i, $m := .Methods}}{{if $i}}
{{end}}{{$m.Doc}}  {{$m.Client}}(
{{- if eq $m.Streaming "server"}}request: {{$m.Input}}, options?: CallOptions): AsyncIterable<{{$m.Output}}>
{{- else if eq $m.Streaming "client"}}requests: AsyncIterable<{{$m.Input}}>, options?: CallOptions): Promise<{{$m.Output}}>
{{- else if eq $m.Streaming "bidi"}}requests: AsyncIterable<{{$m.Input}}>, options?: CallOptions): AsyncIterable<{{$m.Output}}>
{{- else}}request: {{$m.Input}}, options?: CallOptions): Promise<{{$m.Output}}>{{end}};
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
