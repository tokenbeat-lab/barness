# Decisions 运行时与授权撤销

Decisions 复用 Client 的调用生命周期、Binding/Credential/Catalog 快照、有限重试、
准入、时限与 Observer。公共验证入口是 Classify / WithHooks(...).Classify；Jev、
聊天及图像生成的对照使用同一个 Client。

Binding.Retry 零值只发送一次。显式配置时每次 Attempt 重新取得准入许可，所有
Attempt 固定请求 bytes、headers、key、BindingVersion、CredentialVersion 和
CatalogVersion/Hash。header/payload 回调每 Logical Call 一次；response 回调仅在
最终 2xx 一次。任何 2xx 后的读取、回调、解码或校验失败都不重放；合法 refusal
保持 stop/nil，已唯一验证的 usage 保留。退避不持有 body 或许可。

非 2xx 在 MaxErrorBodyBytes 内读取。只有无重复 envelope/error 字段的 JSON
error.message 字符串进入调用者诊断，并走共享 key 脱敏；HTML、损坏或歧义正文
使用固定可读诊断。HTTP status、x-request-id、Retry-After 单独保留。诊断可能
包含厂商回显的请求文字，只交给本次 Tenant；Observer 不接收错误正文。

Unary 成功响应使用 MaxOutputBytes，不套 SSE Frame 或事件队列限额。完整 JSON
读到 EOF、解码和原子结果校验期间持有许可；limit+1 后停止读取，错误体超限也
不重试。取消/时限的 Code/Phase 按发生边界返回，取消准入等待者只回收其自身资源。

云端 Host 可以针对新 Luna 路线：

- 设 Binding.Enabled=false，新调用 tenant_denied/binding。
- 从 Binding.AllowedModels 移除 classifier 型号，新调用 tenant_denied/capability。
- 设 Binding.AllowClassifierImages=false，直接图片调用 tenant_denied/capability；
  回调加图为 tenant_denied/request。纯文本仍可授权调用。

这些变更只影响新解析的调用。已取得一致快照的在途调用及其重试按旧授权完成，
已报告 usage/cost 不撤回。没有自动 Jev/chat fallback；同 Account Scope/key 下的
其他 Tenant 和其他 Binding 不合并授权。恢复配置后，以新 RequestID 调用获得新版本。
本地 Client 的目录/政策不可变，替换这两项须重新装配 Client。

公共证据、失败目录与重放见[工单 06](../../.scratch/barness-ai-openai-decisions/issues/06-runtime-isolation-and-rollback.md)
和[运行时证据](../../.scratch/barness-ai-openai-decisions/runtime-evidence/README.md)。
共享运行时默认零重试及 504 行为在自己的披露表登记，不称为新 Decisions Pi parity；
容量压力、固定 oracle、新路线 live 和发布门禁分别由后续工单完成。
