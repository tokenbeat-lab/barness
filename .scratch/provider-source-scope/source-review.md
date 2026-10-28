# 三条路线来源差异审阅（2026-10-27）

范围只含 TypeSafe Jev 1.14、智谱 Responses、Z.AI Responses。原测量及历史失败、授权、claim、账本、支持矩阵、候选目录和暂停状态不变。复用意味着当前消费者重新判断原测量适用，不是再次执行 live，也不增加厂商支持声明。

## 来源与逐文件依据

三份原生 bundle 见 `.evidence/provider-source-scope/selected-evidence-v1.json`。原测量提交分别为 `78843eb99905e95c8fa534a75897e0dc8787b1da`、`d2354144570da8aeda56862bbcaffe323061f986`、`1251199b6463d3319fff5488e77a648c7c31dce3`；原来源哈希对应的完整 Git 字节已留存，完整差异分别在同证据目录的三个 `*-runtime.patch`。不把当前 HEAD 当成原测量来源。

原保守拦截文件数为 25 / 35 / 30。`chat.go`、`anthropic.go` 的主入口只服务其协议，`responses.go` 只服务 Responses：前缀规则漏掉了这三个根文件，导致其他协议也失效。修正后需明确审阅的产品文件为 22 / 33 / 28；另对 Jev 效果评测的 `visualeval/host.go` 保留一次确切字节审阅。每个文件的前后 SHA 和分类在 `ai/internal/testkit/evidence/reviewed_sources.json`，不是按文件名永久豁免。

| 差异 | 实际影响和本次判断 |
| --- | --- |
| admission.go、call_runtime.go | ProviderID 进入本地 AdmissionRequest，属于产品变化；不进入厂商请求。用当前 Provider admission/隔离普通测试验证，不能称作注释变化。CostEstimate 的新增路径只为 Kimi/xAI。 |
| cache_key.go v1→v2 | Provider 加入缓存键，SessionID 非空时确实改变 wire。两条 Responses 原测量全部 40 个请求无 prompt_cache_key，全部 options 的 SessionID 为空；TypeSafe 不使用此缓存键。本次仅复用这些实测场景。当前普通缓存 v2 隔离/授权测试验证键的生成和隔离；不新增真实缓存命中、计费或跨市场缓存声明。 |
| history.go、旧 glm_history_ids.go → opaque_history_ids.go | 共用历史处理有变化；GLM Responses 在 opaque allocator 入口提前返回，不使用新分配器。两站同 Provider 原生历史/工具结果逐条重建请求，保持实测 wire。xAI 历史和跨市场语义不由这些旧输入推导。 |
| Responses capability/assistant/request/options/errors/stream/usage | 现有智谱和 Z.AI 仍走 GLM 分支；新增 Kimi/xAI 的 request ID、reasoning、usage、stream 状态和指令处理由 Provider/capability 条件隔离。两站 GLM 5.3 / 5.2 的实际目录投影及 SDK 不变，原完整场景继续由原生消费者检查。 |
| binding、catalog_validate、hooks、preflight、http_failures、SDK header、provider_usage_guard、retry、observer、replay、cost_estimate、call.go | 新 Provider 常量、限定 Provider 的目录/失败/header/cost/usage 分支，或已有 GLM 条件的等价 capability 判断；TypeSafe/本次同站 GLM 路径不改变实测请求与响应语义。Z.AI 原提交已有 ProviderID/usage guard/cache v2，不重复归为本次变更。 |
| catalog_models、estimate、transport、usage | 完整差异为注释；仍绑定确切前后字节，后续同文件代码变化会重新拦截。 |
| visualeval/host.go | 新增 finalbudget.Reserve，确实改变本地预算前置控制，不改变请求内容、结果解码或评分。保留历史预约/授权核验，当前受控预算测试验证；不得借此恢复旧授权或启动下一轮。 |

## 消费规则与边界

1. 原始完整来源图、快照、manifest/seal、审计、目录、原授权时刻、账本、预算和真实捕获继续独立核验。没有覆盖原文件、刷新旧时间或改写支持矩阵。
2. 差异审阅表只允许三个 Provider/API 的确切文件前后 SHA；无运行时参数、环境变量或调用方传入的“豁免”。未审阅产品差异、依赖、新增/删除运行时代码、相同文件后续改动仍拒绝；新的产品子包也必须进入来源采集。
3. 视觉效果消费者不再要求其他 Provider/测试工具/只读消费者的整个源码集合恒等；所有历史 source/ 字节仍按原 SHA 校验，实际 visualeval 执行/评分文件仍逐项比较，唯只读 verify.go 不属测量执行代码。未知生成器、评分或 host 修改继续拒绝。
4. 原生 mandatory cases、SDK、型号能力/价格、wire、Result、Observer 和历史来源校验没有替换成来源审阅表。补测只由实际新增厂商事实或缺原件触发。

## 离线回放与限制

重复执行入口为 `.scratch/provider-source-scope/replay.go`，固定只读两条选定 Responses bundle；只构造 synthetic credential，通过内存 RoundTripper 返回留存响应，没有 dialer，也没有 grant/claim 写入口。助手历史必须与同 bundle 的实测 Result 完全对应、租户和账户一致，才通过现有 TrustNativeState 恢复原来源。

智谱 control-low/control-max 原测量在 high 基线上通过 hook 仅替换 reasoning.effort 为 low/max；回放复现这个原 hook，不用整包替换 payload。最初未提供显式 ResourcePolicy 的回放失败及随后漏复现这四个 hook 的失败均保留，不当作厂商失败或消除原测量。

最终回放预期：40/40 当前编码请求与捕获结构相同，36 个正常完成的内容、用量、stop、response ID 与原 Result 一致。4 个取消场景只有截断 SSE，不要求“读完流”等价于原取消时序；请求编码仍核对，取消/部分用量沿用原生消费者的已观测帧/Result/Observer 校验和当前受控取消普通测试，不能把回放补成新的真实取消测量。

Jev 原 12 个正式 protocol 场景通过现有当前公共 Classify 解码回放，12 个效果样本通过完整效果消费者及评分核验；其 OnPayload 使用留存最终 payload，因此请求编码等价不能仅依靠该回放断言。本次 TypeSafe 编码器字节未变，其他共用差异逐项审阅，当前普通 TypeSafe 编码/图片/hook/预算测试另行验证。

## 验证和后续

先写失败清单及 E2E 反例，再实现：三条实际旧来源为 RED；根协议主入口反例为 RED；遗漏未知产品子包采集为 RED。新规则只放过已审阅差异，错路线、改旧来源、后续编辑、依赖、未知新增/删除、评测生成/评分改动保留反例。完整消费者仍保留大量捕获/授权/场景篡改测试。

最终完整受影响普通包、工具进程和 vet、当前标准消费者及保护核验结果统一在 `status.md` 与独立版本索引。无需新增厂商事实的本次三路线付费补测为 0；live 执行、race、压力、暂停轮次恢复均未执行。此前 Kimi 六个 xai-origin-history 场景和 MiniMax 国际 Messages incomplete/FAIL 的独立缺口不在本次范围，保持原义。

## GLM 旧采集清单的两个遗漏

当前采集器增加产品子包后，智谱旧 `glm-execution-source.json` 暴露出 `ai/internal/clock/clock.go`、`ai/internal/probe/probe.go` 未列入旧路线清单；这不等于它们在当时不存在。Responses 的 `.evidence/minimax-current-eleven-glm-2/source-at-start.json` 完整执行快照含两文件，原授权 sourceManifestSHA256、原 intent、命令回执、二进制清单和实际留存 live.test 哈希相连；旧提交原字节与当前相等。全程只读取二进制并计算哈希，没有运行它。

为避免相同采集规则的修正使已交付智谱 Chat/Messages 失效，只读回归这两条路线。它们各自原执行 `receipt.json` 的完整 source.files 同样含相同字节；不是拿今天的 checkout 替代旧来源。逐份绑定见 `.evidence/provider-source-scope/source-omission-custody-v1.json`，新增原件保护清单为 `supplemental-protection-before.json`。

恢复规则仅覆盖三个确切旧清单的 canonical SHA、Provider/API 和两个固定文件的固定原哈希；修改旧清单的任一字段、后续修改 clock/probe、新增其他内部包都拒绝。审阅表 90 行中，83 行是三个目标路线的确切产品差异，1 行是效果 host 的预算前置变化，6 行是三个旧清单的两个缺项恢复。原清单和报告不补写，历史标签/授权不变。

首次扩大采集后的智谱 Responses 消费 FAIL、据此取消的第二个未完成全包 attempt、遗漏恢复的 RED/GREEN 均保留。不是为绿色结果放开所有 missing source，也未建立来源恢复框架或新的豁免入口。目标三条及两个智谱回归最终在 `consumption-v3` 统一通过；最终当前公共回放为 `public-replay-v5`，原 v3/v4 保留。
