# Model admission audit (baseline 9584d3d6f095aefa4ab73bdc75c39a705b7931d6)

只读审计；未运行测试。范围 ai 顶层运行库、catalog 维护与加载；资格固定名单及 Host 示例授权不属于运行白名单。行号为基准版本。

## A. 型号名准入

- ai/kimi_catalog.go:11：Moonshot 两站仅允许 kimi-k3、kimi-k3.1-pro。删除 ID 比较，保留三个已实现 API、文本模态/容量、effort、兼容标志及价格规则。此函数同时被 catalog_validate.go:35 和 catalog_definitions.go:42 调用。
- ai/catalog_xai.go:11：仅 grok-4.7；删除 ID 比较，保留 Responses API 和其余能力/容量/思考配置约束。xai_binding.go:14-19 是操作、协议与服务端点限制，不是型号白名单。
- ai/image_pricing.go:28：仅 grok-imagine-image-2.0；删除 ID 比较，保留 image API、能力、per-image 价格表和有效性验证。
- ai/classifier_capabilities.go:34-40、50：TypeSafe SystemOne 图片、Box、位置能力以 jev-1.14.0 名称准入。用 provider/API 与已声明能力判断，保留图像数量/MIME、可用位置、Box 等实现边界；须确认 TypeSafe 适配器对任意同协议 ID 的具体 wire 能力。

关键空 ID：Catalog.validate（ai/catalog_validate.go:20-78）是直接 Config.Catalog -> NewClient 及 Catalog.Hash 的路径，没有通用非空身份验证；ValidateDefinitions（ai/catalog_definitions.go:19-27）只保护外部定义加载。移除 A 的型号判断会让符合其他条件的空 ID 通过直接 Host 构造路径。至少在共享 validate 保护非空 ID，且核对现存直接构造契约；外部加载也应继续拒绝空或空白身份。不要仅依赖 Host AllowedModels。

## B. 保留

- ai/call_runtime.go:84 Binding.AllowedModels 是 Host 逐租户授权；catalog_lookup（ai/catalog_query.go）按 operation/provider/API/ID 精确定位，均保留。
- ai/catalog_definitions.go:20-27 非空身份和已实现路由；ai/catalog_validate.go:22-25 四维身份去重；其余价格、模态和 API 能力约束保留。
- ai/native.go:25,100,142 及各 payload/stream 模型相等判断（如 chat_payload.go:16、responses_request.go:340、anthropic_sequence.go:25、responses_xai_stream.go:102）是目标身份、payload 防篡改或返回身份一致性，不是名单。历史同模型签名判断 gemini_request.go:342-347 保留。
- catalog/maintenance/official.go:36 显式四维身份、upstream.go:101 来源关联、cost_update.go:120 范围过滤均不是名单。ai/internal/testkit/live 等资格固定范围和历史凭据保持不变。

## C. 型号驱动行为，非简单准入

- ai/responses_usage.go:75-85：OpenAI priority/fast 对 gpt-5.5 用 2.5 倍，其余 2 倍，flex 0.5；删除会改历史计价。ModelCost 没有服务等级乘数字段，保留行为或新增明确的 JSON 服务等级费率/倍率并迁移原值，不能误用 context cost tiers。新增字段是行为保持且未来只改 JSON 的必要条件。
- ai/gemini_options.go:155-165：按型号正则/别名选择 thinkingLevel 或 token budget；192-201 以 2.5 系列名称选择默认预算。删除会改变请求语义，现有 ThinkingLevelMap 只映射等级，不表明 wire 形式或默认 token 数；需 JSON 声明 thinking wire 模式和预算映射才可无型号判断保持旧行为。
- ai/gemini_request.go:351-375：按 Gemini 版本或 claude-/gpt-oss- 名判断 tool-call ID 和函数响应图片嵌套。删除将改变历史/多模态 wire；现有 ModelCompat 未表达这两能力，需明确 JSON 能力字段并迁移旧定义。此类行为并不拒绝新型号，但与“同类只更新 JSON”目标有关，应独立说明范围。

## 公共验证建议

先公共 E2E 红灯：自定义新 ID 的 Kimi 两站三 API、xAI Responses 和 xAI image、TypeSafe 图像/Box；外部 JSON 加载和直接 Config.Catalog 路径，分别断言空 ID 拒绝、合法未知 ID 接受、无授权拒绝、非法能力/协议/价格拒绝、请求发出的精确 model、响应/usage/价格和目录身份。参考 TestCatalogCandidateContracts、TestCatalogCandidateDataOnlyAndCodeGap、TestCatalogCandidateCheckApply、TestBuiltinCatalogReleaseSnapshot，以及相应 provider E2E；C 如改动须追加 Gemini reasoning/tool history/image 与 OpenAI service tier 计价回归。测试证据以 SOP run.py 实际结果为准；真实资格不因受控 E2E 获得或扩大，live/race/pressure 按影响判定。目录维护文档要求 changelog 和外部 JSON/内置重建分别判定；旧程序外部 JSON 兼容需实际检查，不能推断。
