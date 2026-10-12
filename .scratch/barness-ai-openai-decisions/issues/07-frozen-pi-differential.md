# 07: 纳入冻结 Pi 的 Decisions 差分与有限差异账本

**What to build:** 维护者用唯一固定 Pi 1.1.0 对新 Classify 路线和目录重放共同场景，比较完整请求与基本答案，逐项解释有限差异，同时继续验证 Jev 原行为。

**Blocked by:** 03: 交付 score/bool 混合答案、分布与逐题拒答；04: 证明 Decisions 用量完整性与输入阶梯计费；05: 打通分类图片、三重授权与完整资源边界

**Status:** resolved
**Type:** task
**Source tickets:** 06
**Requirements:** R02, R08, R15, R19
**Acceptance cases:** OD14, OD48, OD49, OD50
**Updated:** 2026-10-13

## Responsibility

负责已冻结 oracle 的 Decisions classify/catalog 分派、公共 fixture、完整投影和有限 case/path 账本。现有共享运行时足以启动差分，不必等待压力或真实服务；运行时发现交其责任票处理。协作时保留他人改动。

## Failure scenarios first

- 用聊天 Luna 目录做分类，type/source pin 错配或进程读了真实 key、endpoint/外网。
- 为通过比较删除整个 body/result/header，新增默认 protocol-wide extension。
- refusal、score 详情、JSON 数字精度、排序、用量/重试差异未各自证明。
- P07 原断言被删，或新增字段导致旧 metadata 比较靠全量排除解决。

## Acceptance criteria

- [x] 先写 route/catalog/source/isolation 与 public differential 反例，再扩展现有 runner；不更换 frozen pins 或增加第二套 oracle。
- [x] classifier 类型与 OPENAI_CLASSIFIER_MODELS 分派准确；只允许受控 loopback，清洁环境，不读 Provider 环境凭据。
- [x] 比较完整共同 wire 和基本答案；数字键顺序及精度有固定对照，不按 index 冒充名称相等。
- [x] 图片、拒答、score 分布、presence/total/reasoning、strictness、callback 时机、headers/错误/重试逐项登记有限 case/path/理由/需求与自身 E2E。
- [x] 共同字段仍比较；未登记差异为 pending，fixed 再现为回归；零 pending 才完成。
- [x] 新路线是冻结 Pi 自身路线，纳入 P10 确定性差分，不登记为没有 upstream 的 Extension Route。
- [x] Jev 原 E2E/diff/官方 replay 保持；新增能力元数据投影只删除有来源的新增字段，不排除旧 catalog 内容。
- [x] artifact 含完整 captures、pi/source/sdk/catalog Hash、请求/结果哈希、发现列表、实际 verdict、manifest/audit 和准确命令；门禁最终必须消费 P10 的此项证据。

## Completion

交付自己的协议差分与有限扩展证据；其他现有协议差分仍保留。仅映射或禁用差分得到的 NOT_RUN 不能勾选完成。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 06

2026-10-12：仅为新增路由扩展当前 oracle，不升级已经冻结的 Pi 版本。


### 实施与验证（2026-10-13）

实施 `dbdd77c`，两轴审查修复 `6acff6d`。冻结 Pi 1.1.0/commit/model-data/SDK lock 不变；
Decisions classify/catalog 使用其自身 classifier 路线，受控 endpoint 与来源反例通过。
P10 46 份完整差分 PASS、pending 0，49 条有限 case/path/kind 账本；普通 PASS、
NOT_RUN、删映射、SDK/source/result/capture 漂移反例继续失败。旧 Jev 全部 E2E、
12 个差分、官方 replay 和整个旧目录序列化保持。

最终全套退出 0，7268 E2E（7259 PASS、9 既有压力 opt-in NOT_RUN），861 差分
全部 PASS，pending 0；专项 race 347 PASS，Node 9 PASS，两种 vet 通过。
初次全套发现旧聊天 Costs 被同名 classifier 价格遮蔽，已保留原 chat 查询语义并补反例。
Standards 与 Spec 复审各 0；完整 raw captures、pi/source/SDK/catalog/request/result
hashes、原始失败收据、实际 gate verdict、manifest 与脱敏审计见
[固定证据](../frozen-differential-evidence/README.md)。

实际门禁消费全部七个子集，differential PASS、审计 PASS。本票不以此替代完整发布；
正式目录、十组合 live、设计负载与独立效果仍由后续票交付。
