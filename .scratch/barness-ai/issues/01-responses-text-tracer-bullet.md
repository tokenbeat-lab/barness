# 01: 首个 tracer bullet：OpenAI Responses 纯文本经 Client 走通

**What to build:** 库外 Go 调用程序通过公开 Client，以可信 CallScope（TenantID、RequestID）和 Target（BindingID、ModelID）发起一次 OpenAI × Responses 纯文本生成：Stream 与 Complete、full 与 simple 四个入口都能拿到统一事件和最终 assistant 消息。宿主替身提供绑定与凭据；本地受控 Provider 以脚本回放 SSE 并捕获实际请求；每次离线 E2E 生成可重放的脱敏证据包。这一票同时奠定后续所有票复用的主验收边界（spec Testing Decisions §1–§3）。

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] 建立 barness-ai Go module（包标识为合法 Go 标识符），公共领域类型与 SDK 类型在 adapter 边界独立转换，SDK 类型不出现在公开 API
- [ ] Client 构造必须传入资源策略；策略缺失、必需容量/时限为零或负数、字段关系非法时构造失败（D1 / ADR-0002 的构造校验部分；限额执行见 12、13）
- [ ] Client 构造后配置只读、并发安全；adapter 按 Binding.API 从显式注册、构造后只读的注册表选择
- [ ] 宿主替身实现 BindingResolver / CredentialResolver（按 `(TenantID, BindingID)` 定位、AllowedModels 与最小版本化模型目录取交集）；核心不读取环境变量、不缓存明文 key 或“已认证 Provider client”
- [ ] Responses adapter 使用 OpenAI Go SDK（研究锁定 v3.66.0 为起点，纳入时重新核实），关闭 SDK 默认重试、禁用 CookieJar、不跟随携带凭据的重定向
- [ ] Stream 提供 Next / Event / Result / Err / Close；Stream 系列先返回流对象，解析与 I/O 在后台进行；Complete 复用同一生产/归并过程
- [ ] 正常文本流产生 start → text start/delta/end → done，Result 与 done 携带相同最终消息，成功时 Result/Complete 的 error 与 Err() 均为 nil
- [ ] Result 携带不可变调用元数据：TenantID、RequestID、BindingID、实际 ProviderID/API/ModelID
- [ ] 以 response.completed 判定成功终态（其余终态见 05）
- [ ] 本地受控 Provider：服务端脚本控制分帧（含逐字节拆帧、CRLF、多行 SSE、心跳、未知事件），捕获 method/path/query、请求体与收到的测试 key 别名；测试 transport 只允许 loopback
- [ ] 每次 E2E 输出证据包：运行清单、case ID、版本与模型/fixture 哈希、脱敏请求与响应脚本、事件与最终结果、断言报告和单场景重放命令
- [ ] `go test ./...`、`go test -race ./...`、`go vet ./...` 全部通过且无外网依赖
