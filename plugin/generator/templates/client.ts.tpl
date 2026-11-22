/** Calls {@link {{.Name}}} over a transport: HTTP, Connect, or a service binding. */
export class {{.Name}}Client implements {{.Name}} {
  readonly transport: Transport;

  constructor(transport: Transport) {
    this.transport = transport;
  }
{{range .Methods}}
{{.Doc}}  {{.Client}}(request: {{.Input}}, options?: CallOptions): Promise<{{.Output}}> {
    return this.transport.unary({{$.Name}}Descriptor.methods.{{.Client}}, request, options);
  }
{{end}}}
