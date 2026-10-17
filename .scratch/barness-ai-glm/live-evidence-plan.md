# GLM 有界候选证据计划

日期：2026-10-17（Asia/Shanghai）。契约：[spec R21 / T03](spec.md)、[工单 11](issues/11-glm-candidate-live-evidence.md)。

## 前置与当前状态

当前状态：NOT_RUN。只检查了本地 `.env` 的变量名，未发现 GLM 凭据；已向 Host 请求标准 API Key 的文件路径/环境变量名与账户、区域别名。此计划不声称已发起任何请求，也不提供正式支持证据。

收费执行前必须：三条路线自己的公共 Client 离线验收通过；明确中国区标准 Key 用途；记录账户/区域别名；每进程只装配该组合的 Key；启用 live build tag 与显式运行开关；固定候选目录、SDK、Git 和来源 Hash。核心不读取 dotenv。不得使用其他 Provider Key、套餐 Key 或替代 endpoint。

## 固定预算与预留

每协议独立进程：24 Logical Call、32 HTTP Attempt、每次最多 4096 输出 token、1800 秒进程时限、单调用最多 120 秒且服从更短 Host 时限。Binding/SDK 重试关闭。发送前预留 Logical Call 和 Attempt；Provider 故障的场景至多重跑一次，重新预留并计入同一预算；到达上限即 FAIL，不动态放宽。

两型号各预留 9 次基础调用（18 次）：full Stream、full Complete、simple Stream、simple Complete、自动函数调用、工具结果续接、可信原生历史续接、流中取消、取消后恢复。工具相关调用使用 4096 输出 token，其他调用仍不可超过 4096；低预算造成 thinking 截断应保留失败。

预留 4 次候选单变量对照（每型号 2 次），总计 22 次；余下 2 次供至多一次的场景级厂商故障重跑。没有重跑时仍不得增加探针，未能收敛的事实保持未完成。每一对照在执行前登记型号、基准调用、仅改变的字段、预期和可接受结论；若 4 次不够覆盖必交未知，不运行猜参，先记录未完成并修订规范和证据计划。

## 按协议的证据和未知变量

| 组合 | 官方资源 | 基础场景中的对照 | 额外单变量候选 | 不能推断的结论 |
| --- | --- | --- | --- | --- |
| zhipu-chat | `/api/paas/v4/chat/completions` | 省略/显式 low/high/max 在不同基础场景分别捕获；5.2 独立记录 | 两型号各一次 temperature=0、各一次 stream_options.include_usage=true；对应无字段基础请求 | 接受字段不代表实际使用；5.2 不证明 5.3；缓存缺项不当零 |
| zhipu-responses | `/api/v1/responses` | 原生 reasoning/content 实际形状、store=false 完整历史、工具 call_id、无 DONE 终态 | 两型号各一次 low 和一次 max；基准 high 保持其他参数相同 | Chat 的 effort 不证明 Responses 的服务端强度；不得从一帧推断全部合法形状 |
| zhipu-messages | `/api/anthropic/v1/messages` | SDK 默认头、基本 SSE、工具 JSON/结果、thinking/signature/usage 原始字段 | 每型号最多两次针对基础结果揭示的未知变量；具体字段须事先登记有来源假设 | 无 Messages 字段证据时不试猜 budget/effort；纯文本成功不证明工具/签名回放 |

Messages 文档没有承诺的高级字段默认不发送；必须取得真实 thinking 与工具形状后再固化 fixture 和实现。禁止捏造签名。若无可观察 thinking，只有满足 spec R14 的官方/真实证据才能为该可选能力登记 UNSUPPORTED；工具往返和合法继续生成仍为必交。

## 固定输出

每型号每场景：合成任务、最终请求和原始 SSE（含截断标记）、事件、Result、Attempt、usage presence、错误、request id 来源、预算实际消耗、耗时、美元估算及其价格快照、断言。每包：运行命令、SDK/目录/Git/source Hash、账户/区域/Key 类别别名、完整场景集合、脱敏审计和校验索引。

缺 Key：NOT_RUN；真实失败：FAIL；不得自动换账号/协议/模型，不能删掉任一目标型号或必交场景。候选报告不能更新正式矩阵；正式目录纳入后必须以正式 Hash 重跑自身报告。现有十组合与十三组合发布合同按工单 12/13 保留，缺证据时不改变当前正式目录。
