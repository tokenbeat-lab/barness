# 工单 06：两轴 code-review

固定点 `26d955b260c2ac3f08bdce3bba810772d5948d49`；
审查命令 `git diff 26d955b260c2ac3f08bdce3bba810772d5948d49...HEAD`。
实施 `c9124ef`，补强断言 `fb30d25`；两个独立只读 agent 并行审查并复审补强内容。
未纳入用户起点已有的其他工作区变更。最终执行/证据封装由主 agent 完成。

## Standards

0 documented-standard breaches；0 actionable baseline smells。

Decisions 诊断使用 bounded body、duplicate-aware envelope/error 转换和共享 key redaction，
符合 AGENTS.md 4/9 与 ADR-0009 的 tenant-only diagnostic 边界。公共 E2E 复用既有
runtime 和 Probe，mutation 在实际发送后的 barrier 才执行；同账户/key 没有合并租户授权，
符合 ADR-0003/0008/0020。独立 fixture 来自真实的宽／窄能力控制，不是推测性抽象。

补强复审：四路线费用按独立 fixture 的 31/47 Input、chat 1 Output、image 20 Image Output
算出的字面量正确；Luna 按捕获 data URL 核对租户自己的真 PNG/JPEG，能检测跨租户串图。
504 披露准确标为 route-specific exclusion 和 DECLARED_NOT_PARITY。无新增发现。

证据封装补充审查发现 2 项：P2 聚合 observer-records.json 名称未触发现有字段白名单，
以及 P3 README 把适用 artifact 写成所有历史 case 都完整。已保留原聚合文件并添加
字节相同的 observations-* 审计视图，内部 manifest 登记映射及同一 SHA；同时改正文，
明确前置失败与装配失败的实际保留内容。只读复审确认两项均解决，最终无遗留。

## Spec

0 actionable findings；无遗漏／错误实现／范围扩张。

OD40–46/58 的有限重试、2xx 不重放、取消/超时/资源、并发租户与操作隔离、不可变快照、
有界 metadata-only Observer，以及新调用撤销与在途完成/费用保留均通过公共入口表达。
生产诊断修复落实 R13“非 2xx 错误保留有界脱敏诊断和 x-request-id”，未增加第二个 runtime
或 fallback。已有 DecisionsUsageAfter2xx 与 DecisionsImagesCancellation 是保留执行的相关覆盖。

补强复审：每操作独立费用和图片精确字节加强 R14；修正 Pi 504 覆盖例外的错误措辞符合 R13。
全部在工单范围内。完整证据、race/full/vet 与 SHA-256/整包审计见 validation 和 README，
这些证据不授予新 Decisions parity、压力或 live 的发布结论。

证据封装补充审查发现 1 项文档准确性问题：两项历史装配失败只有 assertions.json，
前置取消没有 HTTP 脚本。已修正完整性措辞并说明历史 HEAD/working-tree 来源，
复审确认解决；归档始终保留原 verdict 与实际文件，不补造缺失的历史记录。

最终：Standards 0；Spec 0。
