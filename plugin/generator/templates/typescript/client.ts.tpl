/** Calls {@link {{.Name}}} over a transport: HTTP, Connect, or a service binding. */
export class {{.Name}}Client implements {{.Name}} {
  readonly transport: Transport;

  constructor(transport: Transport) {
    this.transport = transport;
  }
{{range .Methods}}
{{- if .Streaming}}
{{.Doc}}  {{.Client}}(
{{- if eq .Streaming "server"}}request: {{.Input}}, options?: CallOptions): AsyncIterable<{{.Output}}>
{{- else if eq .Streaming "client"}}requests: AsyncIterable<{{.Input}}>, options?: CallOptions): Promise<{{.Output}}>
{{- else}}requests: AsyncIterable<{{.Input}}>, options?: CallOptions): AsyncIterable<{{.Output}}>{{end}} {
    if (!this.transport.{{.Streaming}}Stream) {
      throw new StatusError("UNIMPLEMENTED", "{{.Client}} streams, and this transport does not: use connectTransport");
    }
    return this.transport.{{.Streaming}}Stream({{$.Name}}Descriptor.methods.{{.Client}}, {{if eq .Streaming "server"}}request{{else}}requests{{end}}, options);
  }
{{else}}
{{.Doc}}  {{.Client}}(request: {{.Input}}, options?: CallOptions): Promise<{{.Output}}> {
    return this.transport.unary({{$.Name}}Descriptor.methods.{{.Client}}, request, options);
  }
{{end}}{{end}}}
