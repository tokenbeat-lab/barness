# 31: SDK 自带的宿主 OS/架构请求头

**What to build:** 去除 openai-go（及 anthropic-sdk-go）自动附加的 `X-Stainless-Os`、`X-Stainless-Arch`、`X-Stainless-Runtime-Version` 等描述宿主的请求头：`ai/transport.go` 的 `userAgent` 注释说明不向厂商账户暴露共享宿主的 OS、版本与架构（维护者决定 2026-10-01），但 SDK 仍发送这些头（维护者决定 2026-10-02，ADR-0016 维护者决定第 7 条）。

**Blocked by:** —

**Status:** resolved

- [x] 确认两个 SDK 发送的全部 `X-Stainless-*` 头及其取值来源
- [x] 与 pi 的差异登记为差分账本扩展（pi 经 openai-node/anthropic SDK 同样发送这类头）
- [x] 在 adapter 的请求构建处统一删除，离线 E2E 断言服务端收不到这些头

## Comments

**2026-10-02 — from issue 23.** Observed in the live smoke's wire captures (redacted request headers, run against a local TLS fake): openai-go v3.66.0 sends `x-stainless-arch: arm64`, `x-stainless-os: MacOS`, `x-stainless-runtime-version: go1.26.2` and others alongside `User-Agent: barness-ai`.

**2026-10-02 — maintainer decision** (ADR-0016 维护者决定 7): remove the host-describing `X-Stainless-*` headers from both SDKs' requests, consistent with the `userAgent` decision of 2026-10-01; register the difference from pi in the differential ledger.

**2026-10-02 — implemented.**

- **Headers (both SDKs, same source):** `internal/requestconfig` of openai-go v3.66.0 and anthropic-sdk-go v1.75.0 sets, on every request, `X-Stainless-Lang: go`, `X-Stainless-Package-Version`, `X-Stainless-OS` (`runtime.GOOS`, normalized), `X-Stainless-Arch` (`runtime.GOARCH`, normalized), `X-Stainless-Runtime: go` and `X-Stainless-Runtime-Version` (`runtime.Version()`), plus `X-Stainless-Retry-Count` and `X-Stainless-Timeout`. Only OS, Arch and Runtime-Version describe the machine; the rest describe the SDK or the request and stay (their pi differences were approved 2026-10-01). Gemini is direct HTTP and sends none.
- **Removal:** `withoutHostDescription` (`ai/sdk_middleware.go`) is the last SDK middleware on both paths (`openai_sdk.go` for Responses and Chat, `anthropic.go`); it deletes the three headers on every attempt, whoever set them — a host's header transform included, as the decision removes them without exception.
- **E2E:** E06 `checkWire` asserts no request of any protocol, scenario or entry carries them (280 requests failed before the change); the P01 and P02 header-transform cases assert that even an `X-Stainless-Arch` the transform sets is not sent.
- **Ledger:** per SDK protocol (openai-responses, anthropic-messages, openai-completions) `X-Stainless-Os`, `-Arch` and `-Runtime-Version` are `only_pi` extensions; the earlier `changed` Runtime-Version entries are replaced. `BARNESS_AI_PIDIFF=1` full run: 0 pending.
