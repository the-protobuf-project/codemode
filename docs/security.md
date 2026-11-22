# Security

Model-written code is untrusted. The design keeps it away from everything but
the methods you expose.

| Threat | Control |
| --- | --- |
| Code reaches the network | No `fetch` in the sandbox. With `outbound: allowlist`, `fetch` goes through the host, HTTPS only, to listed hosts; responses are capped and recorded |
| Code reads files, environment or secrets | QuickJS has no host objects. Deno runs with every permission denied by flag, which also holds when the host itself is Deno. Workers get no bindings but the RPC bridge |
| Credentials leak | Credentials live in the transport on the host. The sandbox sends and receives protojson only |
| Runaway code | A deadline per pass (default 30 s) and a memory cap (default 64 MB). QuickJS interrupts loops; Deno is killed; Workers and iframes are discarded |
| Forged or malformed calls | The bridge accepts only exposed `namespace.method` paths, and each input codec validates arguments before the service is called |
| Destructive calls without consent | Destructive and elicitation methods pause for approval by default; the approver sees the method and arguments from the log |
| Replay drift | A call that differs from the log at its sequence number fails the execution |
| Large values | Arguments, results and step values over `max_value_bytes` are refused; results returned to the model are truncated |

## Notes

- The `IframeExecutor` cannot interrupt a synchronous loop inside the frame;
  removing the frame at the deadline is the only stop. Prefer QuickJS in the
  browser for untrusted code.
- Approval is about intent, not authorisation. Services must still check that
  the caller may do what it asks.
- `rollback` is compensation, not a transaction: a revert can fail, and the
  report says which ones did.
