# codemode.v1 annotations

`codemode/v1/annotations.proto` defines the Code Mode options: the sandbox
namespace, instructions and snippets of a service, and the approval, replay,
revert and examples of a method. See [docs/annotations.md](../docs/annotations.md).

`codemodepb/` holds the Go types the plugin reads options through. Regenerate
them after changing the proto:

```sh
just proto
```
