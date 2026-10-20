# ZAI Messages GLM-5.2 窄思考形状的实施证据

本记录仅保留工单08的公共 Client/CLI 普通开发事实；功能当前状态只在 [spec](../spec.md#当前状态)。[自身四请求研究](zai-thinking-research.md) 先证明 exact enabled/budget1024/max4096 原生返回，再建立 RED；没有增加真实调用，也没有执行 race。工单08仍为 claimed，等待新的真实 candidate qualify。原始81调用中的 GLM-5.2 native-history FAIL 及原九个无 thinking 捕获保持不变。

实现只在 ZAI 私有 profile 独立 opt-in 既有精确子集，为自身 GLM-5.2 Messages 候选设置 Compat1024。版本统一提升至 `regional-candidates-2026-10-20.3`，只更新候选、快照与 pin；官方 source-files / 原件未改。无选项和 GLM-5.3 保留省略，simple 不推导 canonical effort。任意预算、错误输出配对、额外 thinking 字段以及 metadata 扩权仍被拒绝。parser 无改动，不补造原字段。

公共 E2E 保留三组逐字节 research-only request/SSE 原件：52显式、52省略、53省略；以独立原始 delta Hash 核对 thinking/signature，不将内容打印到失败消息。同源/Host 恢复后的第二轮是明确受控合成响应，只证明 Client 保留与重放，不能冒充真实 continuation。新的正式 full 采集从独立 ModelCompat 选择52精确形状，simple options.json 保存实际省略的等价请求形状，供原生 SSE replay 使用。

| 实际轮次 | exit | 事实 |
| --- | ---: | --- |
| red | 1 | 三个公共 Stream / Complete / OnPayload case 均因当前能力/最终请求边界拒绝而 RED |
| green-exact | 0 | 三 case PASS，实际 JSON 与自身研究请求相等，原生字段 Hash 相等 |
| native-boundaries | 1 | 新 E2E 使用了不存在的 ThinkingOff 常量，构建失败；保留原回执 |
| native-boundaries-fixed | 0 | 36 native case PASS，含省略对照、受控续接、预算 / metadata / hook / simple 边界 |
| red-harness | 1 | 实际受控 CLI18请求成功，但新断言确认52 full 缺精确 thinking 形状 |
| green-harness | 0 | 修复后原生进程18 case PASS，父消费者23 case PASS |
| affected | 1 | 父 native1037 PASS / 5 FAIL；独立 worktree 缺既有冻结依赖导致 PiDiff 拒绝，新增有限映射使独立合同 pin 拒绝 |
| pidiff-anthropic | 1 | 同一缺依赖原因，70个差分 case FAIL，未冒称成功 |
| red-harness-options | 1 | 新断言确认52两个 simple 场景保存的 replay options 与实际省略请求不一致 |
| affected-final | 0 | 1127 Go test PASS；父 native1060 PASS，六个采集进程81 PASS、其它受控子进程28 PASS；共1169 native case PASS，31份原生审计均 findings0 |
| pidiff-anthropic-final | 0 | 72 Go test PASS，70个实际 Anthropic frozen PiDiff native case PASS |
| vet / vet-live / vet-final / vet-live-final | 0 | 两轮双 vet 均通过；最终来源与最终普通验证一致 |

最终受影响普通、Anthropic PiDiff 与双 vet 的完整受测来源为 `507945547a32e3e9f0b4080be71026af4606bfc7335d6387288ccf9cb50a1b03`，每轮 source-before/source-after 相同。已通过的首轮检查后因实际 harness 记录行为修复启动最终 affected 集合；未新增无关全量测试。最终 affected 另含21个实际 Regional frozen PiDiff；旧 Pi/SDK/dependency pins 和差分豁免未改，临时依赖 symlink 已移除。

有限 trace 只在 R06 / R10 / R16 / R18 / RP13 / RP18 增加实际新 case，并补齐11已有三个 promotion/composition case；原字段、原 requiredCases、原场景、P0、差分与 live集合均未删改。独立 normalized release-contract literal 从 `836ebe0c…` 提升为 `c4daf796957695003fc1caacf49a2c95ef5529904f3d61291f764256f57e5c2d`，删除新增关键 case 的公共合同断言继续拒绝。旧 `cannot-grant-China-budget` caseID 为追溯保留，测试语义根据新自身证据改为拒绝未证2048，注释和测试名明确说明，不再声称1024仍被拒绝。

[全部实际 argv / 数字退出码 / UTC / 来源 / 原生指针与审计索引](../../../.evidence/provider-sites/final-acceptance-20261020/zai-fix-dev/index.json) / [有限集合只增不减回执](../../../.evidence/provider-sites/final-acceptance-20261020/zai-fix-dev/trace-additions.json)。所有中间失败保持原文件、原身份与实际 exit，不刷新旧证明。
