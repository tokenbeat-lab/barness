# 08: 形成分类图片设计负载与可核验容量证据

**What to build:** Host 使用有限示例政策运行含 Luna 图片分类的本地和云端混合负载，获得实际请求/图片容量、并发、内存、吞吐和取消恢复证据；artifact 导入会拒绝缺失或伪造分类图片读数。

**Blocked by:** 06: 证明故障、快照隔离与授权撤销行为

**Status:** resolved
**Type:** task
**Source tickets:** 06, 09
**Requirements:** R06, R13, R17, R19
**Acceptance cases:** OD47, OD57
**Updated:** 2026-10-13

## Responsibility

负责设计负载声明、对应 Host 图片政策装配、公共 mixed-pressure 执行、资源 artifacts 与发布 artifact 的离线导入核验。全局发布合同由最终门禁票接线。协作时保留他人改动。

## Failure scenarios first

- 把 Jev 文本压力报告当 Luna 图片证明；增加图片能力但没有额外输入字节或并发读数。
- 仅检查一个 policy，漏每图/总图/count/完整 JSON、base64 扩张或取消后的资源。
- GC 单点读数掩盖暂存峰值；把剩余 headroom 和限额使用率混称。
- 缺分类图片读数、改用旧报告、缩负载或改预算仍获 gate PASS。

## Acceptance criteria

- [x] 先写 public design-load 与实际 artifact 导入的删字段/错值/旧报告反例，再落地政策和负载扩展。
- [x] 保留已有本地 4 / 云端 8 并发与 256/512 MiB 混合预算、旧负载合同；新增分类图片负载说明每租户调用组合、状态/问题/图片大小和支持路线。
- [x] 同时覆盖 Jev 文本及 Luna 图片分类，不用替换旧调用来减少旧合同；必要时分有界批次，每批满足声明并发，预算依据清楚。
- [x] 新 MaxImages、每图 MaxImageBytes、MaxTotalImageBytes、MaxRequestBytes 的真实 design/limit/ratio 全量保存，包含 base64/data URL/JSON 编码成本。
- [x] 每项限额占用 ratio=used/limit <=75%，剩余余量至少 25%；不为 PASS 缩小负载或放宽原预算。
- [x] 强制并发占满与 unary 完整结果保留期间采样，报告 GC reachable heap 下界及 TotalAlloc 保守上界、Provider 复制量、吞吐和资源归零。
- [x] 取消后能恢复调用；资源、量测、配置与准确命令进入 manifest/audit；普通与 race 发布运行均要求适用证据。
- [x] artifact 验证器实际要求分类图片字段/身份/负载/count/总 bytes，缺失或矛盾失败；存在旧 mixed-pressure 文件不足以通过新项。

## Completion

独立负载与导入核验可重复验证。原负载若已失败，保存原失败并定位责任，不改变合同使其消失；最终门禁票使用本票的真实新指标。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 06

2026-10-12：仅为新增路由扩展当前 oracle，不升级已经冻结的 Pi 版本。

### 原票 09

2026-10-12：计划 ready-for-agent 不等于已发布；本任务只写设计与实施工单。


## Answer

2026-10-13：实施提交 `fd1b05a`、复审修复 `9c3e0ff`、实际 native 问题字节补正
`0856b2b`。Host 本地/云端示例配置 Classifier.Images（4 张、总 1 MiB），旧四路线
混合负载、4/8 并发和 256/512 MiB 预算原样保留。另以每租户四次 Luna 图片分类，
成功、完整读取处取消、同 Client 同负载恢复三批各占满并发。目录为 Host 合成测试
目录，PNG 为有效 1×1 图片补齐字节的容量输入，未宣称正式目录或像素处理能力。

每调用两张 256 KiB PNG、128 KiB state、八个 4000 字节问题。实际 decoded/base64/
data URL/完整 JSON 为 262144/349528/349550/862760 字节，native predicate 4050 字节；
全部设计/限额/比例保存，政策最大占用率 50%。每批完整 unary 和结果保留采样、
GC 下界、TotalAlloc 上界、Provider 复制、吞吐、取消资源归零与恢复均保存。
race 最大累计分配约本地 222.05 MiB、云端 448.55 MiB，低于原预算。

实际 artifact 导入反例先 RED，再 GREEN；缺字段/null/错值/旧 Jev 报告/伪造小内存/
重复结果/错误 Attempt 或答案键均拒绝。全部变异重算 manifest 哈希，验证身份、
准确负载、实际 wire 与结果、三批资源的语义一致性。两轴独立复审剩余发现均 0。

最终源码普通与 race 各 182 个本票适用用例和 861 个 PIDiff 用例 PASS。全仓普通
7445 PASS、1 个旧 local chat TCP ENOBUFS FAIL；race 全仓 7446 PASS、0 FAIL，
两轮审计无发现。失败及原负载单独重放 PASS 均原样保留。更早旧 100ms CallTimeout/ReadIdleTimeout 竞争也保存原 FAIL 和独立重放，
不改变合同使失败消失。最终退出状态、审计与完整实测见
[证据及重放说明](../classifier-load-evidence/README.md)、
[容量表](../classifier-load-evidence/capacity.md)、
[执行收据](../classifier-load-evidence/validation.json)、
[两轴复审](../classifier-load-evidence/REVIEW.md)。固定包保留真实 applicable artifacts、
原失败、源码、准确命令、全仓 manifest/audit、整包成员 SHA-256。

本票独立负载与离线导入核验已交付；live 和 Release Gate 为 NOT_RUN，正式支持/
效果与十组合最终发布消费者接线由后续工单完成。
