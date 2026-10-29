# xAI Imagine 集成审查

固定基线：`1e244ee26f024ce3534eed0600c43f6317d77adb`。受审版本：`dfe983ba6fe6ceb3adb592821d630eb44292cc14`。

## Standards

部分完成：静态审查发现 1 项硬违反和 1 项测试判断。

- **[P2] 硬违反：图片校验吞掉取消原因。** `ai/xai_images_response.go:61` 将 `validImageBase64` 的所有失败转为 `protocol`。该校验器在 `ai/openai_images_response.go:129` 检测取消后返回 `false`。取消发生在完整响应后的大图校验期间时，本路线丢失 `canceled/deadline_exceeded` 语义。`ai/xai_images.go:98` 也直接返回解码结果，缺少现有 OpenAI/Google 路线的最终取消检查。这违反 `AGENTS.md` 原则 5 的取消与故障要求及 ADR-0022 的 aborted 合同。请在图片循环、校验失败和发布前保留 `contextError`。

- **[P2] 判断：目录的凭据零读取断言无效。** `ai/e2e/xai_images_catalog_test.go:44` 只检查 `ReadsFor("catalog-discovery")`。该 ID 从未传入任何调用。若目录代码错误地使用另一个作用域读取凭据，这项断言仍通过。它不能证明断言名称中的“discovery never resolves credentials”。依据 `AGENTS.md` 原则 9 的行为可测试要求，请为目录构造与查询注入计数 resolver，并断言总读取数为零。

已验证：静态读取固定差异 `1e244ee26f024ce3534eed0600c43f6317d77adb...dfe983ba6fe6ceb3adb592821d630eb44292cc14`。检查 45 个 Go 文件，均不超过 500 行。原生 harness 保持候选与正式资格边界。工单保留先失败后实现的证据指针。

未验证：未执行新测试。上述取消结论来自代码路径分析。

Standards 共 2 项；最严重项为取消原因丢失。

## Spec

审查发现 3 项契约缺陷。范围：`1e244ee26f024ce3534eed0600c43f6317d77adb...dfe983ba6fe6ceb3adb592821d630eb44292cc14`，28 条提交。静态审查覆盖协议、价格、目录、测试与候选 harness。未执行测试。

1. **[P2] 重复字段删除独立的合法消费事实。** 规格第 165 行要求「合法 Reported Cost 与部分 token 报告不因 Content 清空而消失」。`ai/xai_images_usage.go:17–19` 遇任意重复字段就返回空用量和金额。例如唯一 `cost_in_usd_ticks:400000000` 加重复 `input_tokens` 会丢失 USD 0.04；重复 ticks 也会删除唯一合法 token。必须保留唯一字段，返回 protocol，并对重复字段保留未知。

2. **[P2] 图片验证中取消返回 protocol。** 规格第 146 行要求「取消沿现有 aborted 语义」。`ai/xai_images_response.go:61–62` 把 `validImageBase64` 的取消结果直接映射为 protocol。`ai/xai_images.go:98` 也没有验证结束后的取消检查。必须在验证失败和终态发布前检查 context。补充受控取消 E2E。

3. **[P2] 带类型资源缺少编码前总量检查。** 规格第 197 行要求「已知超量载荷在昂贵处理前拒绝」。`ai/xai_images_payload.go:61–73` 只检查各 URL。`ai/xai_images_request.go:83` 依赖的 `jsonPayloadSize` 把普通 struct 计为 1 字节。因此，两张各自合法、合计超过 MaxRequestBytes 的 `[]struct{Type, URL string}` 会先执行 Host Marshaler 并编码大请求，再拒绝。必须累计普通带类型资源的完整已知字节，补充总量 E2E。

未发现必须额外收费才能修正的事项。候选保持非正式支持。live、race、压力的 NOT_RUN 不是上述代码缺陷。

Spec：3 项，最高优先级 P2；主要影响是失败时丢失合法消费事实。

## 修复跟踪

两个审查轴分别保留。去重后4项，交单一实现者按公共E2E先失败后修复：取消、独立消费事实、带类型资源总量、目录resolver总计数。复核与最终普通回执尚未执行。
