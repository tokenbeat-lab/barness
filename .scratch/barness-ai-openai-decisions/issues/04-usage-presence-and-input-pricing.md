# 04: 证明 Decisions 用量完整性与输入阶梯计费

**What to build:** Host 从 Classify 的结果和 Attempt 区分未上报、部分上报与完整零用量，按 Decisions 自己的输入费率得到估算；响应答案失败后仍能核对已唯一上报的消耗。

**Blocked by:** 02: 打通 Decisions 文本单选与最终请求闭环

**Status:** resolved
**Type:** task
**Source tickets:** 04
**Requirements:** R11, R12, R19
**Acceptance cases:** OD35, OD36, OD37, OD38, OD41
**Updated:** 2026-10-12

## Responsibility

负责 Decisions usage 独立转换、presence、一致性、Checked Sum 和输入费率验证；通过公共单选即可完成，不依赖 score/bool 或图片。复用公共 Usage/Attempt/Price Snapshot，不复制聊天 cache 拆桶，也不新增原始明细到 Observer。协作时保留他人改动。

## Failure scenarios first

- usage 缺/null 被报为 complete 零；Reasoning 未报与显式 0 混淆。
- duplicate/负数/小数/溢出被选择最后值或修补成完整计数。
- input 扣掉或重复加 cached/write tokens，阶梯只应用于超出部分。
- 输出有报告但免费就被丢弃；答案错误后丢 usage，ResponseModel 替换授权型号。
- 缓存/总量/reasoning 矛盾被接受，使用另一操作的同名型号价格。

## Acceptance criteria

- [x] 先写缺/null/partial/全零、重复/溢出/矛盾和坏答案公共 E2E，再修改 usage 路径。
- [x] complete/partial/unreported 按设计的独立 Decisions 字段 presence 判定，Reasoning unset 与 0 区分；存在 usage 不自动 complete。
- [x] Input 保留完整 input_tokens，Output 如实报告，canonical cache 桶为 0；合法 total 保留或已知项 checked sum 并标 partial。
- [x] 缓存合计不得超已知 Input，reasoning 不超已知 Output，全计数明确时 total 一致；token 只接受非负 int64 整数。
- [x] ambiguous/invalid usage protocol/response 且不选最后值；仅答案损坏时保留唯一合法 usage 到 Result 与 Attempt。
- [x] 271999/272000 用 .1、272001 用 .2 USD/百万的全请求输入价；非零 output/cache 不新增费用，不推算图片 token。
- [x] 授权 ModelID、价格 CatalogVersion/Hash 不被 ResponseModel 或同名 chat 目录覆盖，区域 premium/账单未核对状态明确。
- [x] 2xx 错误仍一个 Attempt，保存 expected/actual cost、原始 presence、终态、manifest 和审计；Jev 用量及费用对照继续通过。

## Completion

通过合成明确计数核验边界，不向服务主动提交超长输入，也不宣称合成 cache 统计已与真实账单核对。

## Comments

2026-10-12：用户批准 13 张草案并替换旧票。此票按新编号发布；下列原票评论按来源保留，ready-for-agent 不表示已实施或依赖已完成。

### 原票 04

2026-10-12：待实施；Refusal 是有效分类结果，整体协议损坏仍为失败。

### 本票实施（2026-10-12）

实现提交 `49f119c`，审查修复 `1e01c9a`。独立 Decisions parser 验证六项 presence、非负 int64、
checked sum 和已知项一致性；Reasoning unset/0、完整全零与未上报分开。保留 inclusive Input 和
免费 Output，canonical cache=0；按授权宿主目录的输入阶梯计费，不修改正式目录或其他协议。
原始 presence 在合成 fixture/capture 保留，不新增 Observer 明细。文档与 ADR-0027 同步。

编码前写失败清单和公共 E2E，首次 294 case 为 168 FAIL/126 PASS；最终新增 307 case 全部 PASS。
两轴 code-review 修复具名 presence 索引及 OD41 的 Decisions 自身生命周期补证据，最终 Standards 0、Spec 0。
最终专项 race 开启冻结 Pi 为 1258 PASS。最终全套开启 Pi 退出 0：7027 E2E case，7018 PASS、
9 既有压力 opt-in NOT_RUN；815 个既有差分全部 PASS。编译检查、普通/live-tag vet 和整包审计通过；
固定 Jev 655 case 核验通过，历史 release FAIL 保留。

合成输入覆盖 271999/272000/272001 × 无/非零/全缓存，全请求 .1/.2 USD/百万。
未发超长真实输入，未推算图片 token；区域 premium、实际账单和真实非零 cache 未核对，live 为 NOT_RUN。
可重放捕获、expected/actual cost、源码/成员 SHA-256、验证收据与审查记录见
[用量证据](../usage-evidence/README.md)。
