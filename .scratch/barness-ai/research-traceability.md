# barness-ai 研究、规范与验收追溯

本表配套 [spec](spec.md)，恢复三份研究的实施细节、来源与验收对应关系。所有“保留/补齐/已映射”只表示规范覆盖，当前没有 barness-ai 实现或测试通过证据。具体行为以 spec 为准；本表不能独立放宽契约。

来源缩写：S = [原需求研究](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/spec.md)，D = [原技术设计](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/design.md)，V = [原 E2E 计划](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/e2e-test-plan.md)。冻结版本和三个文件的 SHA-256 见 spec 的 Further Notes；源文件保持只读。旧名称仅存在于历史来源链接中。

本表中的 I1–I10 指 spec 的 Implementation Decisions 第 1–10 节；E01–E11、P01–P06 指 Testing Decisions 的横向场景和协议专项。实施用例应以 E/P 标识关联原 T/C 条目，运行报告还必须指向实际请求、事件、结果与断言证据。

## 1. 范围与原需求验收

| 原研究条目 | 当前规范位置 | 处置及验收 |
| --- | --- | --- |
| S§1，S§2：独立库、本地/云端共用、单次生成轮次与 Harness 分工 | Solution、[I1]、[I2]、Out of Scope | 改为 barness 基础模块；协议转换保留，Agent/Session/工具执行仍属宿主；E01/E10 |
| S§2：文本、图片及工具结果图片、工具 JSON/校验 | [I6]，P01–P06 | 全部保留；图片支持与占位降级分别测，参数完整性与 schema 校验不等于执行工具；E03 |
| S§2：原生推理状态、统一及完整选项、流/结果 | [I2]、[I6]、[I7]、[I8] | 保留签名/密文、full/simple、Result 独立等待；E01–E04 |
| S§2：注册、模型元数据、更新/撤销、观测和有界资源 | [I3]、[I4]、[I9]、[I10] | 租户扩展与公共协议语义分开验收；E06–E11 |
| S§3：四类协议与六个首期接入组合 | [I4]，P01–P06 | 按研究建议纳入首期；DeepSeek Responses 属于扩展，不能冒充冻结 pi 路由已覆盖；E10 |
| S§3：官方云 API、服务身份、网关与订阅账号区别 | Out of Scope | 官方云 API/身份及网关后续独立规范；订阅账号路径不交付，不把服务身份 OAuth 等同订阅登录 |
| S§2：图片生成/音视频、批处理、WebSocket、动态模型目录、紧凑帧持久化 | Out of Scope | 按本次范围决策延期；不将这些能力算作首期已完成 |
| S§3：pi-messages、Cloudflare AI JS binding；S§2：JS 宿主入口 | Out of Scope | 明确排除，保留 Go 原生模块边界 |
| S§5：pi 可观察兼容语义与租户扩展 | [I5]–[I10]，Testing Decisions§5 | 事件、终态、选项、历史、usage 做差分；扩展单独断言，不把 SDK/文件结构当兼容目标 |
| S§6 验收 1–3 | [I3]、[I4]、[I6]，P01–P06 | 协议能力、A/B 并发隔离、失败前零推理请求；E01–E03/E06/E07 |
| S§6 验收 4–7 | [I3]、[I8]、[I9] | 更新/撤销、EOF/错误、限额/资源释放、脱敏；E02/E05–E09 |
| S§6 验收 8–10 | [I4]、[I6]–[I10]，Testing Decisions§5 | 全量差分、simple/full/reasoning/Usage、跨 Provider/协议路由；E03–E05/E10/E11 |
| S§7 R1–R5 | Further Notes 范围决策与 D1/D2 | 将原建议与本次采用的范围区分；D1 是新增选择，不能反写为原研究已确定 |
| S§8 来源 | 本表§6–§7 | 恢复直达源码、测试和探针的链接；不要求重新寻找原始行为出处 |

## 2. 多租户不变量逐项映射

| 原条目 | 当前契约 | 对应验收与不能遗漏的证据 |
| --- | --- | --- |
| T01 所有入口显式 TenantID | [I2]、[I3] | E07：full/simple、Stream/Complete 缺 TenantID 时服务器零推理请求；本地也显式归属 |
| T02 可信宿主身份 | [I3]，Testing Decisions§1 | E10：伪造 body/header 租户和历史引用在宿主契约边界拒绝，不宣称库完成登录认证 |
| T03 租户+binding 定位与归属 | [I3] | E06/E07：同名 binding、同模型的 A/B 实际发送各自 key，绑定越权零请求 |
| T04 不允许普通请求覆盖鉴权/目标 | [I3]、[I7] | E04/E07：Authorization、endpoint、代理/项目、payload/headers 变换不能形成授权旁路 |
| T05 无凭据兜底 | [I3]、[I5] | E06/E07：污染环境变量、缺 key、禁用及后端故障，服务器不收到其他身份请求 |
| T06 调用快照与共享实现隔离 | [I3]、[I5]、[I8] | E05/E06/E07：交错重试/取消不串配置、key、结果，已有调用保持快照 |
| T07 历史、结果、原生状态及引用归属 | [I2]、[I6]、[I9] | E03/E10：同名 session、恢复封套、汇合流逐事件归属；缓存/服务端引用为后续前置约束，不虚构首期缓存测试 |
| T08 独立取消 | [I8] | E06/E08/E10：A 取消不影响 B，下游断开/发送失败释放本次资源 |
| T09 观测关联与脱敏 | [I2]、[I9] | E09/E10：租户/调用/尝试对应、实际厂商 request ID、无秘密泄漏；未解析状态不伪造身份 |
| T10 更新与撤销新鲜度 | [I3]，D2 | E07：K1→K2、新调用见撤销、缓存传播契约、在途请求保持原快照、解析冲突前置失败 |
| T11 credential-dependent 操作同一路径 | [I2]、[I4] | E07：首期全部入口无绕过解析器的网络入口；以后发现/图片生成也必须携 scope，未实现能力不记 PASS |
| T12 租户与厂商账户计量区分 | [I9]、[I10] | E05/E09/E11：每次尝试的已知/未知 usage、账户归属、价格版本，不将初始化零值当免费 |

## 3. 技术设计细节与取舍

| 原设计条目 | 当前位置 | 保留/调整的细节与验收 |
| --- | --- | --- |
| D§1、§2 调用路径与接口草图 | [I1]、[I2]、[I8] | 保留四入口、显式 scope、失败最终消息、单消费者与并行 Result/Close；函数草图及目录形式不作为固定实现；E01/E02 |
| D§2.1 ID 含义与粒度 | [I2]、GLOSSARY | RequestID 不等于 RunID/TurnID/JobID；Actor 不是 worker；重试新 Attempt、再次调用新 Request；E05/E09 |
| D§3 绑定字段、解析、一致性 | [I3] | 版本/账户/权限/凭据引用独立，Resolver 可信，缺失/禁用/读取失败保留分类上下文；D2 收窄为不一致时失败；E07 |
| D§3.1 Provider/API 关系 | [I4] | 协议简称与 pi API 标识对应；DeepSeek 两 binding 可授权引用同账户 key，不混淆模型配置；E10/P05/P06 |
| D§3.1 DeepSeek Responses 限制 | [I4]，P05 | 保留 previous_response_id/conversation/store 不受支持及 reasoning_text.delta/done；按研究快照核对请求能力，不从 OpenAI 推导 |
| D§4 本地/云端装配 | [I3]，Testing Decisions§1 | 本地显式秘密来源、云端任务不带秘密、可信历史；最小调用程序验证宿主交接，完整云端服务属后续宿主模块；E10 |
| D§5 密钥生命周期与重试 | [I3]、[I8] | 无 key/provider 切换；0 次默认、Retry-After 优先级、500ms/8s/25% 退避、60s 单项上限和取消；E05/E07 |
| D§6 数据模型、TypeScript→Go 表 | [I1]、[I2]、[I6]–[I8] | 显式合法变体、presence、JSON Schema 运行时校验、live PartialView、最终错误消息；不复制 TS 类层级；E01–E04 |
| D§6.1 历史转换 | [I6] | 同模型空文本签名/redacted；跨模型删除非空 thoughtSignature、text 只保留文本；图片占位/工具 ID/缺结果/system 顺序/error 历史；E03 |
| D§6.2 统一与完整选项 | [I7] | reasoning 等级/禁用/预算、输出余量、samplingParams 覆盖与其他 API 忽略；明确 cacheRetention/metadata；E04 |
| D§6.2 输出元数据及恢复 | [I2]、[I6] | Result/流身份、逐事件信封、AccountScopeID、可信恢复而非 trusted=true；E03/E10 |
| D§6.3 回调时序与 Harness 桥接 | [I7] | transformHeaders 合并头后且 adapter 前；payload 修改/替换区分；onResponse 在 start 前；Anthropic/Gemini 差异；before_payload 与 after_response 分工；E04 |
| D§6.4 SDK 优先与直接 HTTP 条件 | [I5]，本表§5 | H1–H5 保留原含义；先扩展/原始响应，按组合回退；旧探针不替代目标库；E01–E11/P01–P06 |
| D§7 流与资源 | [I8]、[I9] | 先返回流、后台设置、成功 Err=nil、Result 无消费依赖、Close 阶段差异、终态预留、无静默丢 delta；D1 新增强制策略；E01/E02/E08 |
| D§7 transport | [I5] | 无 CookieJar/带凭据重定向；共享对象无鉴权状态；未来不同 mTLS/代理身份分 transport profile；E06/E07 |
| D§8 状态隔离、观测、费用 | [I9]、[I10] | 应用缓存/原生句柄后续约束、账户维度、观测丢失计数、不提供账本保证、Usage 数值与完整性分开；E09/E11 |
| D§9–§10 批次与参考构建 | [I10]，Testing Decisions§6，Further Notes | 先建失败场景和基线，各批有通过门槛；生成模型数据在独立副本补齐并记哈希，不修改收集源码 |
| D§12 待评审选择 | Further Notes、ADR-0002/0003 | 范围采用与新增决策单列原因、代价和重议条件，不伪称原研究已有定论 |

## 4. E2E 计划逐项映射

| 原条目 | 当前场景/章节 | 保留的验收要求 |
| --- | --- | --- |
| V§1–§2 边界、装置、证据 | Testing Decisions§1–§3 | 真实 SDK/HTTP、本地脚本、假 key/环境污染、屏障/时钟/ID、版本/哈希；边界调整只缩减宿主产品范围 |
| V§3 差分规则 | Testing Decisions§5–§6 | presence、数组/事件顺序、一一 ID 映射、原始请求/帧哈希、首个不同位置和处置；禁止宽松删除字段 |
| V§4 OpenAI Responses | P01 | completed/incomplete/failed、缺终态 EOF、reasoning replay、工具分片、usage presence |
| V§4 Anthropic | P02 | message_stop、空文本签名/redacted、交错块、缓存及 1h 计价、SSE error |
| V§4 Gemini | P03 | thought/thoughtSignature、缺 finishReason EOF、function result 与图片结果路由 |
| V§4 OpenAI Chat | P04 | finish_reason、流末 usage、工具 JSON 分片及 reasoning 扩展 |
| V§4 DeepSeek 双协议 | P05/P06 | 两 binding 分别验收；不支持字段、reasoning_text/reasoning_content、[DONE] 前 usage、模型配置独立 |
| C01 生命周期 | E01 | start/块事件/唯一终态、live partial、Result 不消费、Complete 无队列积压；补阻塞解析和成功 Err=nil |
| C02 失败与截断 | E02 | setup/HTTP/流内错误、断连/半 JSON、length、缺终态 EOF；保留部分消息与 usage |
| C03 历史与工具 | E03 | 原生状态及封套、降级、ID 关联、system/tools 顺序、输入不改写、部分 JSON/修复/schema |
| C04 参数与回调 | E04 | 全部等级及 presence、预算/maxTokens、回调修改/替换/失败/次数/时序；补明确 header、metadata/cacheRetention |
| C05 重试 | E05 | 判定优先级、退避/Retry-After、服务端延迟上限、取消、SDK 不额外重试、流中不重放、固定快照 |
| C06 租户隔离 | E06 | A/B 同名配置交错、认证/取消/超时独立、环境污染，实际收到的 key 与目标而非仅 race 结果 |
| C07 授权与更新 | E07 | 身份/绑定/模型/秘密故障、轮换/撤销/在途重试；另验 D2 无内部重解析及恢复 |
| C08 资源与关闭 | E08 | 全部限额、慢消费者/只等 Result、Close 并发、终态预留、释放 I/O/许可；另验 D1 构造校验 |
| C09 观测与脱敏 | E09 | 成功/拒绝/重试/中断/Close、慢/失败 Observer；秘密与正文不记录，授权原生结果完整 |
| C10 路由与宿主 | E10 | 跨 Provider 共享协议、同 Provider 双协议、可信身份/历史、下游取消；补合流事件信封 |
| C11 用量与成本 | E11 | 缺失/null/零值、缓存/1h、reasoning 子集、阶梯价格、每尝试完整性；不把 unknown 当免费 |
| V§6 live 约束 | Testing Decisions§2、§5 | 六组合低权限测试账户、CI secret store 逐进程注入、token/次数预算；故障注入只在本地；PASS/FAIL/NOT_RUN/UNSUPPORTED |
| V§7 批次与门禁 | Testing Decisions§6 | 本地/差分/live 分报告；B/C 批门槛、race/vet、压力/脱敏、对应组合独立通过 |
| V§8 来源与边界 | 本表§6–§7 | 迁移具体场景并保留测试适用范围，不将 SDK 测试/探针改名后当模块 E2E |

P01–P06 均需运行适用的 full/simple、Stream 完整消费、Stream 仅 Result、Complete 和错误终态。矩阵是一项发布承诺的完整证据集合，不意味着所有输入都做无意义的笛卡尔积。

## 5. SDK 历史证据与尚未完成的门槛

以下为[选型验证记录](/Users/cyber/RestoX/harness/docs/research/pi-ai-go-sdk-validation.md)的报告内容，本次仅迁入上下文，没有重跑测试。锁定版本：OpenAI v3.66.0、Anthropic v1.75.0、Google GenAI v1.71.0。

- 报告记载 28 个离线顶层测试及 race 通过、5 个 live 顶层测试默认跳过；四协议简单文本流实际请求 JSON 与冻结 pi 相同。这不是全量公共事件、结果或历史的差分。
- 六个 Provider×协议组合的文本、强制工具和首帧取消共 18 个 live 子场景通过，另有 DeepSeek 推理和历史回放 4 个子场景。通过只适用于当时账户/模型/版本。
- 探针显式装配过 store:false、include_usage:true、Anthropic cache_control 等字段，SDK 不自动保证 pi 默认值；DeepSeek Responses 的不支持字段不能照搬 OpenAI 配置。
- DeepSeek reasoning_text、reasoning_content 和工具历史有正常路径证据；incomplete/failed 的真实触发、服务端拒绝字段及边界组合未逐一实测，P05/P06 必须补齐本地断言并如实记录 live 覆盖。
- DeepSeek Chat thinking 模式下不能强制 tool_choice=required；强制工具探针关闭 thinking，自动工具选择的 thinking+tools 与推理回放另有验证，不能合并成一个笼统通过结论。

| 原硬门槛 | 研究能证明什么 | barness-ai 仍必须证明什么 |
| --- | --- | --- |
| H1 凭据/endpoint/header 隔离 | SDK 显式配置入口样本可避开环境污染 | 绑定/凭据授权、撤销、一致快照及并发隔离；E06/E07/E10 |
| H2 字段、签名、工具和 pi 差分 | 简单请求/工具通过；OpenAI 原始字段 presence、Anthropic 签名可读取；Google typed 保留已知字段但丢失未知字段 | 全量事件/结果/历史/选项差分；必需字段经原始通道保留或按组合改直连；E01/E03/E04/E11/P01–P06 |
| H3 正确终态与部分消息 | SDK 不替库检查所有异常 EOF | 自有归并逻辑识别终态、保留 partial、独立 Result；E01/E02/P01/P03 |
| H4 pi 等价重试 | SDK 内置重试可关闭，探针 HTTP 500 只发一次 | 默认零重试、Retry-After、延迟上限、取消、快照固定及尝试计量；E05 |
| H5 前置资源限制 | 读取前截断 HTTP body 可行 | 请求/帧/工具 JSON/队列/最终结果及全部时限有界；E08 |

H1–H5 均不能因 SDK 样本通过而标为目标库 PASS。优先核查公开扩展/原始响应/受限 transport，仍无法满足时只对受影响组合改直接 HTTP，继续执行相同 E/P 验收。

## 6. pi 源码与具体测试入口

这些链接指向已收集的研究仓库；实施时将固定提交及必要模型数据整理为 barness 自身可重建的参考产物。链接方便核对语义，不要求复制源码目录或把机器绝对路径写入 CI。

| 行为 | 原实现入口 | 可迁移的具体测试 | 对应场景 |
| --- | --- | --- | --- |
| 公共入口/选项/身份扩展边界 | [models](/Users/cyber/RestoX/harness/pi/packages/ai/src/models.ts)、[types](/Users/cyber/RestoX/harness/pi/packages/ai/src/types.ts)、[auth types](/Users/cyber/RestoX/harness/pi/packages/ai/src/auth/types.ts)、[auth resolve](/Users/cyber/RestoX/harness/pi/packages/ai/src/auth/resolve.ts) | [stream](/Users/cyber/RestoX/harness/pi/packages/ai/test/stream.test.ts)；本项目另建租户隔离场景 | E01/E04/E06/E07 |
| 历史与跨模型降级 | [transform-messages](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/transform-messages.ts) | [跨协议历史](/Users/cyber/RestoX/harness/pi/packages/ai/test/transform-messages-copilot-openai-to-anthropic.test.ts)；仅迁移历史变换断言，不引入订阅接入 | E03 |
| system/tools 重放 | [transcript](/Users/cyber/RestoX/harness/pi/packages/ai/src/utils/transcript.ts) | [system replay](/Users/cyber/RestoX/harness/pi/packages/ai/test/system-message-replay.test.ts)、[tool changes](/Users/cyber/RestoX/harness/pi/packages/ai/test/transcript-tool-changes.test.ts) | E03 |
| simple reasoning 与成本 | [simple-options](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/simple-options.ts)、[models](/Users/cyber/RestoX/harness/pi/packages/ai/src/models.ts) | [stream](/Users/cyber/RestoX/harness/pi/packages/ai/test/stream.test.ts)提供用户路径；按规则补预算/presence/计价 fixture | E04/E11 |
| 流与 lazy 错误 | [event-stream](/Users/cyber/RestoX/harness/pi/packages/ai/src/utils/event-stream.ts)、[lazy](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/lazy.ts) | [event-stream test](/Users/cyber/RestoX/harness/pi/packages/ai/test/event-stream.test.ts) | E01/E02/E08 |
| Responses 终态及工具 JSON | [Responses adapter](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/openai-responses.ts)、[validation](/Users/cyber/RestoX/harness/pi/packages/ai/src/utils/validation.ts) | [terminal event](/Users/cyber/RestoX/harness/pi/packages/ai/test/openai-responses-terminal-event.test.ts)、[partial JSON cleanup](/Users/cyber/RestoX/harness/pi/packages/ai/test/openai-responses-partial-json-cleanup.test.ts) | P01/E02/E03 |
| Anthropic SSE/签名 | [Messages adapter](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/anthropic-messages.ts) | [SSE parsing](/Users/cyber/RestoX/harness/pi/packages/ai/test/anthropic-sse-parsing.test.ts) | P02/E02/E03 |
| Google 签名 | [Gemini adapter](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/google-generative-ai.ts) | [thinking signature](/Users/cyber/RestoX/harness/pi/packages/ai/test/google-thinking-signature.test.ts) | P03/E03 |
| 重试判定与协议调用点 | [provider-retry](/Users/cyber/RestoX/harness/pi/packages/ai/src/utils/provider-retry.ts)、[google-shared](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/google-shared.ts)、[Chat adapter](/Users/cyber/RestoX/harness/pi/packages/ai/src/api/openai-completions.ts) | [provider retry](/Users/cyber/RestoX/harness/pi/packages/ai/test/provider-retry.test.ts)、[Chat retry](/Users/cyber/RestoX/harness/pi/packages/ai/test/openai-completions-retry.test.ts)、[Google retry](/Users/cyber/RestoX/harness/pi/packages/ai/test/google-shared-retry.test.ts) | E05 |

旧探针入口：[TS validation](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/ts-validation)、[SDK validation](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/sdk-validation)。可迁移固定响应与请求样本；不能原样作为 barness-ai 的验收，因为它们尚未调用目标模块。

## 7. 官方 SDK 测试的适用边界

以下链接与结论来自 V§8 及 [SDK 测试来源核对](/Users/cyber/RestoX/harness/.scratch/pi-ai-go/sdk-test-research.md)，并非本次在线重新核验。锁定版本的源码引用属于证据索引；不将其内部类型或测试结构变成公共契约。

| SDK | 具体测试来源 | 可迁移的行为 / 不能继承的结论 |
| --- | --- | --- |
| OpenAI v3.66.0 | [工具累计](https://github.com/openai/openai-go/blob/v3.66.0/streamaccumulator_test.go#L30-L137)、[SSE EOF/关闭](https://github.com/openai/openai-go/blob/v3.66.0/packages/ssestream/ssestream_test.go#L338-L470)、[Responses 参数](https://github.com/openai/openai-go/blob/v3.66.0/responses/response_test.go#L18-L133) | E01–E04：分片、结束、presence；参数测试依赖 mock server，SDK 的 Err=nil 不证明 barness-ai 终态成功 |
| Anthropic v1.75.0 | [累计与签名](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/message_test.go#L252-L439)、[流内错误](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/error_type_test.go#L18-L87)、[取消/重试](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/client_test.go#L224-L330) | E01–E03/E05：交错块、input_json_delta/signature_delta、redacted、deadline；单租户累计测试不证明绑定授权、pi 用量或跨模型历史 |
| Gemini v1.71.0 | [回放](https://github.com/googleapis/go-genai/blob/v1.71.0/table_test.go#L319-L505)、[HTTP 流故障](https://github.com/googleapis/go-genai/blob/v1.71.0/api_client_test.go#L230-L647)、[函数参数流](https://github.com/googleapis/go-genai/blob/v1.71.0/models_test.go#L152-L229) | E01–E03/E06：key/path/body、CRLF/空流/非法 JSON；该函数参数流真实测试只覆盖 Vertex，部分图片 bytes 回放被禁用；P03 必须独立覆盖 Developer API 与图片 |

## 8. 新增决策与补齐范围登记

| 项目 | 处置 |
| --- | --- |
| 名称及项目归属 | 按用户要求统一 barness-ai，属于 barness 基础模块；历史证据链接保留原始位置 |
| 测试边界调整 | 公开 Client 主入口保持完整模块验收；宿主仅验证身份/历史/取消交接，实际应用 E2E 由未来宿主模块承担 |
| D1 资源策略 | [ADR-0002](../../docs/adr/0002-barness-ai-resource-policy.md)：新增强制有限策略，记录装配成本、误限风险和默认值重议条件；E08 |
| D2 一致快照 | [ADR-0003](../../docs/adr/0003-barness-ai-snapshot-consistency.md)：收窄为前置失败，记录轮换可用性代价和有界重解析重议条件；E07 |
| 精确行为遗漏 | 已补签名删除、samplingParams 忽略、cacheRetention/metadata、header 时点、异步设置和成功 Err=nil；E01/E03/E04 |
| 归属契约遗漏 | 已补 Result/流字段、AccountScopeID、逐事件信封、未解析状态和 Observer 字段；E03/E09/E10 |
| 协议与 SDK 细节遗漏 | 已补 DeepSeek 字段限制、Google typed 丢字段、SDK 测试适用范围和 P01–P06；不宣称新测试已运行 |
| 来源与构建线索遗漏 | 已恢复源码/测试/探针入口、历史证据范围、生成模型数据缺口及独立副本重建要求 |

| Review 契约调整（2026-10-01） | [I2]：Complete/Stream.Result 返回 `(Result, error)`，失败时 error 非 nil 且 Result 有效，Stream.Err 与之一致；[I6]：库只校验原生状态封套，可信性由宿主经不可反序列化的构造入口担保，无封套/账户不匹配按跨模型规则降级并记录到 Result 元数据与 Observer（[ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)）；[I7]：删除每调用 transport 配置；[I9]：只取 Result 的流仍受队列上限、不采用背压，准入接口携 AccountScopeID 而账户聚合准入由宿主注入；E01/E02/E03/E08/E10 |
| 术语统一 | 本地/远端统一为“本地/云端”，见 GLOSSARY |

以上是此次文档补齐的追溯记录。未来发现原研究还有未覆盖的实际行为，应先登记来源、受影响契约和 E/P 场景，再更新规范；不能仅以“参考原文”作为已完成迁入的证明。

[I1]: spec.md#1-定位依赖与交付形态
[I2]: spec.md#2-公开调用与标识契约
[I3]: spec.md#3-租户授权服务绑定与配置快照
[I4]: spec.md#4-provider协议与模型目录
[I5]: spec.md#5-sdk-与传输边界
[I6]: spec.md#6-消息历史与工具语义
[I7]: spec.md#7-完整选项统一选项与可信回调
[I8]: spec.md#8-流终态取消与重试
[I9]: spec.md#9-资源准入错误与观测
[I10]: spec.md#10-用量成本与交付完成条件
