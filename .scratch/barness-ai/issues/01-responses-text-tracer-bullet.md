# 01: 首个 tracer bullet：OpenAI Responses 纯文本经 Client 走通

**What to build:** 库外 Go 调用程序通过公开 Client，以可信 CallScope（TenantID、RequestID）和 Target（BindingID、ModelID）发起一次 OpenAI × Responses 纯文本生成：Stream 与 Complete、full 与 simple 四个入口都能拿到统一事件和最终 assistant 消息。宿主替身提供绑定与凭据；本地受控 Provider 以脚本回放 SSE 并捕获实际请求；每次离线 E2E 生成可重放的脱敏证据包。这一票同时奠定后续所有票复用的主验收边界（spec Testing Decisions §1–§3）。

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [x] 建立 barness-ai Go module（包标识为合法 Go 标识符），公共领域类型与 SDK 类型在 adapter 边界独立转换，SDK 类型不出现在公开 API
- [x] Client 构造必须传入资源策略；策略缺失、必需容量/时限为零或负数、字段关系非法时构造失败（D1 / ADR-0002 的构造校验部分；限额执行见 12、13）
- [x] Client 构造后配置只读、并发安全；adapter 按 Binding.API 从显式注册、构造后只读的注册表选择
- [x] 宿主替身实现 BindingResolver / CredentialResolver（按 `(TenantID, BindingID)` 定位、AllowedModels 与最小版本化模型目录取交集）；核心不读取环境变量、不缓存明文 key 或“已认证 Provider client”
- [x] Responses adapter 使用 OpenAI Go SDK（研究锁定 v3.66.0 为起点，纳入时重新核实），关闭 SDK 默认重试、禁用 CookieJar、不跟随携带凭据的重定向
- [x] Stream 提供 Next / Event / Result / Err / Close；Stream 系列先返回流对象，解析与 I/O 在后台进行；Complete 复用同一生产/归并过程
- [x] 正常文本流产生 start → text start/delta/end → done，Result 与 done 携带相同最终消息，成功时 Result/Complete 的 error 与 Err() 均为 nil
- [x] Result 携带不可变调用元数据：TenantID、RequestID、BindingID、实际 ProviderID/API/ModelID
- [x] 以 response.completed 判定成功终态（其余终态见 05）
- [x] 本地受控 Provider：服务端脚本控制分帧（含逐字节拆帧、CRLF、多行 SSE、心跳、未知事件），捕获 method/path/query、请求体与收到的测试 key 别名；测试 transport 只允许 loopback
- [x] 每次 E2E 输出证据包：运行清单、case ID、版本与模型/fixture 哈希、脱敏请求与响应脚本、事件与最终结果、断言报告和单场景重放命令
- [x] `go test ./...`、`go test -race ./...`、`go vet ./...` 全部通过且无外网依赖

## Comments

**2026-10-01 — implemented** (module `github.com/tokenbeat-lab/barness`, package `ai`; E2E in `ai/e2e`, test kit in `ai/internal/testkit`).

- The AllowedModels ∩ catalog intersection is enforced by the library core (`Client.authorizeModel`), per spec I4 ("实际可调用范围是目录与 binding 授权的交集"); the host double only resolves by `(TenantID, BindingID)`.
- OpenAI Go SDK pinned at v3.66.0 and re-verified by E2E: `openai.NewClient` reads `OPENAI_API_KEY/BASE_URL/ORG_ID/PROJECT_ID/CUSTOM_HEADERS`, so the adapter builds `responses.NewResponseService` per call from explicit options only and sends its own marshalled body. SDK retries are forced to 0; the shared `http.Client` has no cookie jar and does not follow redirects (v3.66.0 also guards credential redirects itself — ours is defense in depth).
- Plain-http loopback endpoints require the test-assembly opt-in `Config.AllowLoopbackHTTP`; otherwise endpoints must be https.
- Resource policy is validated at construction only; enforcement is tickets 12/13. The Stream event queue is unbounded until ticket 12.
- Built-in catalog lists only non-reasoning OpenAI Responses models (seeded from pi-ai 0.85.1 published data; frozen 0.87.1 checkout lacks generated data). Ticket 02 should re-derive it from the rebuilt 0.87.1 data.
- Evidence bundles go to `.evidence/barness-ai/<run-id>/` (git-ignored; override with `BARNESS_AI_EVIDENCE_DIR`); `manifest.json` lists each case's replay command.
