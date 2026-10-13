# 07: 纳入冻结 Pi 的 Decisions 差分与有限差异账本

**Status:** needs-triage
**Publication:** draft，待用户批准后转 ready-for-agent；尚未发布到 issues。
**Blocked by:** 03: 交付 score/bool 混合答案、分布与逐题拒答；04: 证明 Decisions 用量完整性与输入阶梯计费；05: 打通分类图片、三重授权与完整资源边界
**Source tickets:** 06
**Requirements:** R02, R08, R15, R19
**Acceptance cases:** OD14, OD48, OD49, OD50
**Updated:** 2026-10-12

**What to build:** 维护者用唯一固定 Pi 1.1.0 对新 Classify 路线和目录重放共同场景，比较完整请求与基本答案，逐项解释有限差异，同时继续验证 Jev 原行为。

## Responsibility

负责已冻结 oracle 的 Decisions classify/catalog 分派、公共 fixture、完整投影和有限 case/path 账本。现有共享运行时足以启动差分，不必等待压力或真实服务；运行时发现交其责任票处理。协作时保留他人改动。

## Failure scenarios first

- 用聊天 Luna 目录做分类，type/source pin 错配或进程读了真实 key、endpoint/外网。
- 为通过比较删除整个 body/result/header，新增默认 protocol-wide extension。
- refusal、score 详情、JSON 数字精度、排序、用量/重试差异未各自证明。
- P07 原断言被删，或新增字段导致旧 metadata 比较靠全量排除解决。

## Acceptance criteria

- [ ] 先写 route/catalog/source/isolation 与 public differential 反例，再扩展现有 runner；不更换 frozen pins 或增加第二套 oracle。
- [ ] classifier 类型与 OPENAI_CLASSIFIER_MODELS 分派准确；只允许受控 loopback，清洁环境，不读 Provider 环境凭据。
- [ ] 比较完整共同 wire 和基本答案；数字键顺序及精度有固定对照，不按 index 冒充名称相等。
- [ ] 图片、拒答、score 分布、presence/total/reasoning、strictness、callback 时机、headers/错误/重试逐项登记有限 case/path/理由/需求与自身 E2E。
- [ ] 共同字段仍比较；未登记差异为 pending，fixed 再现为回归；零 pending 才完成。
- [ ] 新路线是冻结 Pi 自身路线，纳入 P10 确定性差分，不登记为没有 upstream 的 Extension Route。
- [ ] Jev 原 E2E/diff/官方 replay 保持；新增能力元数据投影只删除有来源的新增字段，不排除旧 catalog 内容。
- [ ] artifact 含完整 captures、pi/source/sdk/catalog Hash、请求/结果哈希、发现列表、实际 verdict、manifest/audit 和准确命令；门禁最终必须消费 P10 的此项证据。

## Completion

交付自己的协议差分与有限扩展证据；其他现有协议差分仍保留。仅映射或禁用差分得到的 NOT_RUN 不能勾选完成。
