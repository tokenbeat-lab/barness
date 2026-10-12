# 03: 交付 score/bool 混合答案、分布与逐题拒答

**What to build:** Host 在一次文本 Logical Call 中提交 choice/score/bool，得到完整命名答案、score 分布与图例；混合或全拒答仍成功，任何协议损坏使整次 Answers 为空。

**Blocked by:** 02: 打通 Decisions 文本单选与最终请求闭环

**Status:** resolved
**Type:** task
**Source tickets:** 04
**Requirements:** R02, R04, R08, R09, R10, R11, R19
**Acceptance cases:** OD13, OD22, OD25, OD27, OD28, OD29, OD30, OD31, OD32, OD33, OD34, OD39, OD41
**Updated:** 2026-10-12

## Responsibility

负责 score/predicate 请求与最终 guard、完整答案转换和封闭公共结果。用量计数及费率专项由计费切片拥有，保持已有唯一合法用量的保留路径。TypeSafe 舍入和严格 decoder 独立。协作时保留他人改动。

## Failure scenarios first

- bool 被变为阈值判断，true/false 说明丢失，score levels 被排序或合并。
- 返回数组按 index 回配；一次 refusal 清空其他答案或触发错误重试。
- 多/缺/重复题或字段、错误 kind、错 score label/index 与概率损坏被修补或部分发布。
- Decisions 套用 Jev 两位小数例外；TypeSafe 未知 native refusal 被误接受。
- Hook 改题型或 levels 后仍验证原题；全拒答缺少最终问题的完整名称。

## Acceptance criteria

- [x] 先写三题型、合法并列、乱序结果、混合/全拒答及所有损坏响应 E2E，再实现。
- [x] bool 映射 predicate；固定 true/false 拼接规则，省略/空说明对照；score 保留有序重复 label 和公共 2–10 范围。
- [x] Hook 可在已实现类型间合法改名、改种类、重排候选与 levels；完全依据最终冻结记录校验，原对象不被修改。
- [x] 普通与拒答 Answers 完整原子发布；refusal 无编造原因，混合和全拒答 stop/nil，2xx 不重放。
- [x] 缺/多/重复名称、case-alias/重复字段、错 discriminator、非字符串 choice、非法数值整体 protocol/response。
- [x] score 概率覆盖整数 0..n-1、labels 精确，概率和及期望绝对容差 1e-6；保留原数值、Probabilities/Legend，无归一化。
- [x] 已知答案语义严格，benign additive metadata 有界忽略；拒答夹带矛盾普通答案与非法语义字段有明确拒绝证据。
- [x] 任何答案错误保留唯一合法 usage；TypeSafe 原拒答拒绝、宽 JSON、图例与有界舍入 replay 不变。
- [x] 从两个公共入口生成请求、结果、错误、用量、Attempt、断言、manifest 与审计证据。

## Completion

完整文本题型与拒答闭环可单独演示；不宣称已完成图片、usage presence 专项或真实服务验证。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 04

2026-10-12：待实施；Refusal 是有效分类结果，整体协议损坏仍为失败。

### 本票实施（2026-10-12）

实现提交 `aa9da73`。Decisions 支持一次文本调用中的 choice/score/predicate；固定 bool 说明拼接，score ordered 重复 label、原分布和独立 JSON string 图例完整保留。初始和最终 Hook 检查型号 kinds、公共范围及能力交集；Hook 合法改名/种类/候选/levels 后只按最终冻结集合校验。

按 name 匹配完整答案后原子发布，混合与全拒答 stop/nil，矛盾语义字段/重复/别名/非法数值与 1e-6 外的分布整体拒绝，唯一合法 usage 保留，2xx 不重放。benign vendor metadata 在总响应预算内忽略；Jev 独立严格 decoder、宽 JSON 和有界舍入保持。

先列[失败清单](../mixed-answer-failures.md)，再记录公共 E2E 红→绿。新增 292 case 在普通/race 下全部 PASS。专项最终 882 PASS、1 既有 Pi opt-in NOT_RUN；保留初次 250 FAIL 及后续两项测试装配失败，不改写历史。开启冻结 Pi 的完整测试退出 0：6,711 PASS、9 既有压力 NOT_RUN，815 项既有差分全部 PASS。完整 race 退出 0：5,894 PASS、826 默认 opt-in NOT_RUN；两种 vet 通过，全部执行的审计无发现。

两轴 code-review 与验证器补充审查均 0 实质发现，见[审查](../mixed-answer-evidence/REVIEW.md)。固定包包含两入口请求/Provider 脚本、结果/错误、usage/Attempt、Observer、断言、源文件/版本、准确重放、完整 manifest 与审计，见[证据与命令](../mixed-answer-evidence/README.md)。Go 实现从全套测试到最终交付未变化，之后仅补文档与证据。

本票交付完整文本混合题与逐题拒答闭环；图片、usage presence 专项、Decisions 自身差分与厂商真实请求属于后续票，live 仍为 NOT_RUN，未修改正式内置目录或支持矩阵。
