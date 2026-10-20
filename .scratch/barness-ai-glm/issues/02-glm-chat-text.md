# 02: 通过公共 Client 完成 GLM Chat 文本与思考调用

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** Host 使用独立的 zhipu Chat Binding 和候选 Catalog，通过 full/simple 的流式与完整结果入口调用 GLM-5.3、GLM-5.2，并得到正确归属的文本和思考结果。

Blocked by: 01

依据：GLM 接入 spec R01、R02、R03、R04、R05、R06、R07、R11、R22；验收组 A01、A02、A08；User Stories 1、3、5、6、7、8、9、10、12、13、14、15、16、17、18、19、20、21、40、51、52、53、61。

- [x] 先写公共 Client 的文本/思考、鉴权和非法配置失败场景，再实现 ProviderZhipu=zhipu 及现有 Chat adapter 的厂商配置；不新增 GLM 协议或另一套生命周期。
- [x] Host 候选目录中的两型号具有独立的 chat × Provider × API 身份；候选上限有来源或按 spec 标明保守值，合成测试价格不得称为厂商价格，内置目录暂不纳入。
- [x] 标准 Chat Binding 最终生成正确官方资源地址与 Bearer 鉴权；仅使用显式 Credential Snapshot，不读取环境、SDK profile 或自动替换账户/端点。
- [x] Stream、Complete、StreamSimple、CompleteSimple、nil/full 选项及合法 WithHooks 都可验证；错误协议选项、Operation/型号越权和禁用 Binding 在读取凭据/发送前拒绝。
- [x] 请求使用 system、max_tokens 和合法 GLM thinking 形状；不自动生成 store、developer、max_completion_tokens、OpenAI 缓存字段或未核验的 include_usage。
- [x] GLM-5.3 不产生 disabled，仅声明 low/high/max；simple 对其他档位按既有规则夹取，full 的非法显式档位明确拒绝；GLM-5.2 的独立映射不得泄漏到 5.3。
- [x] 采样保持通用模型→canonical 档默认→调用覆盖→OnPayload 的既有顺序和 canonical/wire 区别；最终请求仍满足型号硬约束。
- [x] 读取正文和 reasoning_content，正确处理 Chat 终态及终帧后的合法用量；结果、事件和 Attempt 的 Provider/API/型号/账户归属一致。
- [x] 扩展现有公共 Client 测试/候选装配，使本路线后续可以在双开关、单组合 Key 和 spec 预算内捕获真实证据；离线票完成不要求缺 Key 时伪造 live 通过。
- [x] 留下两型号的可重放文本/思考请求、事件、Result 与错误产物，以及标准 Key Host 配置示例；声明工具往返和正式目录仍由后续票完成。

**独立验收：** 分别通过 full/simple Stream/Complete 演示两个候选型号的文本/思考调用，并展示 5.3 的非法档位与越权请求在本地被拒绝。

**依赖理由：** 需要 01 的厂商能力边界；本票同时提供后续两协议共享的 Provider 身份与候选 Host 装配。



## Answer

2026-10-17：已实现中国区标准 `ProviderZhipu=zhipu` 的 Chat 文本/思考候选。仍复用公共 Client、既有 Chat adapter/SDK、凭据快照与生命周期；Provider 形状为 system/max_tokens、默认省略 store/cache/include_usage，省略思考保留服务默认。两型号独立候选目录不进入内置目录；容量有来源和保守值说明，价格明确为合成 fixture 价格。

`RequiresThinking` 是可省略、可复制且参与 Hash 的型号约束。5.3 明确 low/high/max，simple 夹取、full 不支持档位在读取凭据前拒绝；5.2 有独立官方映射。最终请求在采样/回调后（无 hook 同样）按授权优先检查思考和输出预算，1..min(型号上限, schema 131072) 原样发送，错误字段/越界/非整数/错误类型零请求零 Attempt；具名 null/0 保持省略。合法回调可修复采样，但不能改写非法具名档位的语义。OpenAI/DeepSeek 行为保持。

公共验收最终 166 PASS、controlled live 装配 2 PASS、GLM race 76 PASS、受影响回归 2070 PASS 加独立 oracle guard 1 PASS、可信回调 48 PASS、适用冻结 Chat 差分 321 PASS/零 pending、vet PASS。已提交源重放与完整 red→green 产物、运行命令、manifest Hash、错误产物和来源索引见 [证据目录](../evidence/ticket-02/README.md)、[ADR-0033](../../../docs/adr/0033-barness-ai-zhipu-chat-candidate.md)、[标准 Key Host 配置](../../../ai/internal/testkit/glm/README.md)。

候选 live 装配使用独立 120 秒策略，只复用已有 Host 装配，不进入旧 32-call runner；本票无付费调用、未读 `.env`。11 负责双开关/单组合 Key/24-32 预算预留/完整 live 报告；03 负责工具往返与 clear_thinking，正式目录仍由 12 验收。
