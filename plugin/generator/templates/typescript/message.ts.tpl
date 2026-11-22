{{template "interface" .Decl}}
/** Converts {@link {{.Name}}} between its interface, protojson and the protobuf binary format. */
export const {{.Codec}}: MessageCodec<{{.Name}}> = {
  typeName: {{str .TypeName}},

  fromJson(json: unknown, path: string = {{str .TypeName}}): {{.Name}} {
{{- if .Empty}}
    protojson.object(json, path);
    return {};
{{- else}}
    const o = protojson.object(json, path);
    const m: {{.Name}} = {
{{range .FromJSON}}      {{.Key}}: {{.Expr}},
{{end}}    };
{{range .Oneofs}}    protojson.oneof(m, path, {{str .Name}}, [{{.Fields}}]);
{{end}}    return m;
{{- end}}
  },

{{if .Empty}}  toJson(): JsonObject {
    return {};
  },
{{else}}  toJson(m: {{.Name}}, options?: ToJsonOptions): JsonObject {
{{if .NeedsAll}}    const all = options?.emitDefaults === true;
{{end}}    const o: JsonObject = {};
{{range .ToJSON}}    {{.}}
{{end}}    return o;
  },
{{end}}
{{if .Empty}}  encode(_m: {{.Name}}, w: BinaryWriter = new BinaryWriter()): BinaryWriter {
    return w;
  },
{{else}}  encode(m: {{.Name}}, w: BinaryWriter = new BinaryWriter()): BinaryWriter {
{{range .Encode}}    {{.}}
{{end}}    return w;
  },
{{end}}
  decode(input: Uint8Array | BinaryReader, length?: number): {{.Name}} {
    const r = input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = r.end(length);
    const m: {{.Name}} = {
{{range .Init}}      {{.Key}}: {{.Expr}},
{{end}}    };
    while (r.pos < end) {
      const [no, wt] = r.tag();
      switch (no) {
{{range .Cases}}        case {{.No}}:
{{range .Lines}}          {{.}}
{{end}}          break;
{{end}}        default:
          r.skip(wt, no);
      }
    }
    return m;
  },
};
