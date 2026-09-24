# apervia-sdks

SDKs that read the identity Apervia already authenticated. Each language lives in its own directory and releases on its own tag. Go, Ruby, and Node share one header contract and one conformance fixture.

The licence is Apache-2.0. See [LICENSE](LICENSE).

## Versioning

Each language releases independently:

| Language | Tag |
| --- | --- |
| Go | `go/v<x.y.z>` |
| Ruby | `ruby/v<x.y.z>` |
| Node | `node/v<x.y.z>` |

## Header contract

The sidecar sets these headers on every request that reaches an app. An SDK reads only these names:

| Header | Meaning |
| --- | --- |
| `X-Apervia-User-ID` | The user |
| `X-Apervia-Tenant-ID` | The tenant |
| `X-Apervia-Token-ID` | The token id, for audit correlation |
| `X-Apervia-Env` | The environment the app runs as |
| `X-Apervia-App-Permissions` | Comma-separated app-plane permissions |
| `X-Apervia-App-Roles` | Comma-separated app-plane role names |

The platform story that sends `X-Apervia-App-Roles` is BIZ-141. The sidecar does not send that header yet.

The SDK ignores `X-Apervia-Roles` and `X-Apervia-Permissions`. Those names carry system-plane authority. The SDK also ignores every `X-Platform-*` header. The platform still emits `X-Platform-*` today. The rename to `X-Apervia-*` is BIZ-169, and it is not done yet.

When the `X-Apervia-*` identity headers are absent, the SDK reports an anonymous request. It never returns a partial identity. An empty `X-Apervia-App-Permissions` header is an empty permission list, not an anonymous identity. An absent `X-Apervia-App-Roles` header is an empty role list.

The shared cases are in [conformance/cases.json](conformance/cases.json). Persona user ids are in [conformance/personas.json](conformance/personas.json).

The platform description of this contract is **Identity and access > The platform identity token**:

<https://github.com/the-curve-consulting/apervia/blob/main/internal-docs/content/docs/identity/platform-identity-token.mdx>
