/**
 * Exposes {@link {{.Name}}} in the Code Mode sandbox as `{{.Namespace}}`.
 * Pass the generated client, or any object implementing the interface.
 */
export function {{.Func}}(service: {{.Name}}, options: ConnectorOptions = {}): Connector {
  return connector(
    {
      namespace: {{str .Namespace}},
{{if .Instructions}}      instructions: {{str .Instructions}},
{{end}}      methods: [
{{range .Methods}}        unaryMethod({
          name: {{str .Name}},
          input: {{.InputCodec}},
          output: {{.OutputCodec}},
          requiresApproval: {{.Approval}},
          replay: {{str .Replay}},
{{if .ApprovalMsg}}          approvalMessage: {{str .ApprovalMsg}},
{{end}}          call: (request, context) => service.{{.Client}}(request, { signal: context.signal }),
{{with .Revert}}          revert: async ({{.Request}}, {{.Response}}, context) => {
            await service.{{.Client}}({{.Arg}}, { signal: context.signal });
          },
{{end}}        }),
{{end}}      ],
    },
    options,
  );
}
