# 08: 形成分类图片设计负载与可核验容量证据

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 06: 证明故障、快照隔离与授权撤销行为
**Source tickets:** 06, 09
**Requirements:** R06, R13, R17, R19
**Acceptance cases:** OD47, OD57
**Updated:** 2026-10-12

**What to build:** Host 使用有限示例政策运行含 Luna 图片分类的本地和云端混合负载，获得实际请求/图片容量、并发、内存、吞吐和取消恢复证据；artifact 导入会拒绝缺失或伪造分类图片读数。

## Responsibility

负责设计负载声明、对应 Host 图片政策装配、公共 mixed-pressure 执行、资源 artifacts 与发布 artifact 的离线导入核验。全局发布合同由最终门禁票接线。协作时保留他人改动。

## Failure scenarios first

- 把 Jev 文本压力报告当 Luna 图片证明；增加图片能力但没有额外输入字节或并发读数。
- 仅检查一个 policy，漏每图/总图/count/完整 JSON、base64 扩张或取消后的资源。
- GC 单点读数掩盖暂存峰值；把剩余 headroom 和限额使用率混称。
- 缺分类图片读数、改用旧报告、缩负载或改预算仍获 gate PASS。

## Acceptance criteria

- [ ] 先写 public design-load 与实际 artifact 导入的删字段/错值/旧报告反例，再落地政策和负载扩展。
- [ ] 保留已有本地 4 / 云端 8 并发与 256/512 MiB 混合预算、旧负载合同；新增分类图片负载说明每租户调用组合、状态/问题/图片大小和支持路线。
- [ ] 同时覆盖 Jev 文本及 Luna 图片分类，不用替换旧调用来减少旧合同；必要时分有界批次，每批满足声明并发，预算依据清楚。
- [ ] 新 MaxImages、每图 MaxImageBytes、MaxTotalImageBytes、MaxRequestBytes 的真实 design/limit/ratio 全量保存，包含 base64/data URL/JSON 编码成本。
- [ ] 每项限额占用 ratio=used/limit <=75%，剩余余量至少 25%；不为 PASS 缩小负载或放宽原预算。
- [ ] 强制并发占满与 unary 完整结果保留期间采样，报告 GC reachable heap 下界及 TotalAlloc 保守上界、Provider 复制量、吞吐和资源归零。
- [ ] 取消后能恢复调用；资源、量测、配置与准确命令进入 manifest/audit；普通与 race 发布运行均要求适用证据。
- [ ] artifact 验证器实际要求分类图片字段/身份/负载/count/总 bytes，缺失或矛盾失败；存在旧 mixed-pressure 文件不足以通过新项。

## Completion

独立负载与导入核验可重复验证。原负载若已失败，保存原失败并定位责任，不改变合同使其消失；最终门禁票使用本票的真实新指标。
