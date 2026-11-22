export { catalog } from {{str .Catalog}};
export { config } from {{str .Config}};

/** The implementation behind each sandbox namespace: a generated client or your own object. */
export interface Services {
{{range .Services}}  {{.Namespace}}: {{.Name}};
{{end}}}

/** Builds the connector for every namespace. */
export function connectors(services: Services, options: { [K in keyof Services]?: ConnectorOptions } = {}): Connector[] {
  return [
{{range .Services}}    {{.Connector}}(services.{{.Namespace}}, options.{{.Namespace}}),
{{end}}  ];
}
