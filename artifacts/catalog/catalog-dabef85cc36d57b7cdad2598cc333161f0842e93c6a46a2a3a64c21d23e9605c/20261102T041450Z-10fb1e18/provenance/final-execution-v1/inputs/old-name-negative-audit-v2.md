# 型号名旧负例补充审计 v2

只读静态审计，未执行测试。工作区在并行修改，以下反映检查时所见；这不是测试已通过的结论。

## 范围与方法

扫描 `ai/e2e/**/*_test.go` 和 `ai/*_test.go`；后者当前无匹配的顶层测试文件。对 `.ID` 赋值、复合赋值和追加、字面 `ID:`、JSON `id`、`strings.Replace/ReplaceAll`、`model/alias/old/version/future/name` 表驱动分支、NewClient/Hash/catalog.Load/LoadFile 调用作交叉检索，并追踪相关上下文。原始机器命中列表临时存于 `/tmp/barness-id-hits-v2.txt`、`/tmp/barness-name-hits-v2.txt`、`/tmp/barness-construction-files-v2.txt`。这些列表含响应 ID、测试 case ID 等非模型噪声，不能当成完整语义分析。

排除 `ai/internal/**/*_test.go`、`ai/live/**/*_test.go`、非 Go 测试文件和运行时动态生成的外部输入；未逐字人工审阅全部巨量响应字面量、跨文件间接数据流及非字面反射修改。因此不能保证无其他遗漏。真实资格模型范围不变。

## 修改：旧名称本身不是拒绝依据

- `ai/e2e/kimi_foundation_test.go` `TestKimiCatalogBoundary` 原 `model` 只把 Kimi ID 改为 `kimi-k2.5` 后要求 NewClient 拒绝；应改为空 ID 真负例，保留协议、模态、容量、价格、thinking/compat 负例。检查时已见 `empty-id`，verifier 正处理。
- `ai/e2e/jev_context_runtime_test.go` `TestJevContextCandidateContract` 的 `alias`/`old` 只改为 `jev-latest`/`jev-1.13.0`，应接受本地定义，不宣称真实厂商支持。`webp/local/duplicate/omitted` 涉及独立能力边界，保留真实断言；检查时已见接受分支。
- `ai/e2e/jev_box_catalog_test.go` 的 `jev-1.13.0`、`jev-preview`、`jev-1.14.1` 只改名称，应接受；`decisions/text-only/criteria-only/duplicate` 保留能力拒绝。builtin 固定 ID 仅是内置快照事实；此前 verifier 已修改。
- `ai/e2e/kimi_k31_catalog_test.go`、`xai_foundation_test.go` 原未来名称负例已换空 ID；`local_catalog_classifier_taxonomy_test.go` 未来 TypeSafe SystemOne 已转接受；`catalog_maintenance_official_test.go` 未来名称外加图片模态构成真实未实现能力 code_gap。均已修改，维持这些语义。

## 其他 Model.ID 写入命中及判定

- `catalog_future_model_test.go` 38/50/60/70：四类未来名称是正例；186/193/200 空 ID 负例；保留。115 响应 model 替换是返回身份不匹配；131 missing-target 和 149 hook-model 分别是找不到目标和请求防篡改，保留。
- `local_catalog_classifier_taxonomy_test.go:35` 正例；`local_catalog_operations_test.go:27-39` 先增加 ID 前缀只是隔离不同条目，负例实际为未知字段、类型、价格、API、身份缺失或重复；保留。
- `local_catalog_snapshot_test.go:18` `not-whitelisted` 可加载，调用被 Binding 拒绝，属于 Host 授权；保留。`catalog_concurrency_test.go:26` 重命名产生独立正例；保留。
- `operation_catalog_snapshot_test.go:81-103`、`local_catalog_test.go:31`、`catalog_snapshot_test.go:44`、`client_routing_test.go:37`、`mixed_snapshots_test.go:30`、`regional_snapshots_test.go:90`、`glm_shared_callbacks_test.go:114,120`、`minimax_shared_snapshots_test.go:23` 是快照/拷贝隔离或 hash 内容变化，不断言名称不准入；保留。
- `mixed_world_test.go:50-61` 同名模型用于验证操作身份隔离；`google_images_fixture_test.go:16`、`classifier_image_pressure_test.go:44`、`multimodal_runtime_fixture_test.go:34`、`decisions_usage_test.go:46` 是合成支持模型；`jev_context_fixture_test.go:40`、`jev_state_snapshot_test.go:54` 固定已有能力夹具；均保留。
- `google_images_capabilities_test.go:32` 的 `models/host-google-image` 是路径结构校验，非名称白名单；保留。
- `scenario_fixture_test.go:249` 把场景目标赋给 classifier；`evidence_applicability_test.go:111` 构造无关身份作来源/目录适用性对照；保留。
- `catalog_maintenance_revise_test.go:62` 不同 ID 验证 revise 不得更改四维身份；保留。

## 模型替换、表驱动及来源资格命中

- `xai_stream_faults_test.go:40`、`kimi_messages_faults_test.go:22`、`xai_request_id_test.go:90` 操作响应 model 或用于响应场景，不是 Catalog 准入；前两者测不一致，应保留。
- `images_editing_hooks_test.go`、`images_hooks_test.go`、`jev_context_hooks_test.go`、`jev_instructions_hooks_test.go`、`jev_box_hooks_test.go`、`decisions_images_hooks_test.go`、`decisions_image_inputs_hooks_test.go`、`multimodal_runtime_retry_test.go` 的 model 分支修改请求或 hook payload，校验目标防篡改；保留。
- `xai_route_test.go`、`xai_imagine_contract_test.go` 测 target/路由或请求，非 Catalog 仅改 ID 拒绝；保留。
- `xai_formal_process_test.go`、`xai_live_process_test.go`、`xai_images_live_process_test.go`、`kimi_history_admission_test.go`、`kimi_composition_promotion_test.go`、`minimax_thinking_supplement_test.go`、`regional_release_tamper_test.go`、`chinese_routes_process_test.go` 的 model/alias 分支是授权、原生证据或历史资格固定身份；保留。
- `glm_messages_replay_test.go`、`thinking_sampling_lifecycle_test.go`、`kimi_responses_cache_snapshot_test.go`、`kimi_chat_snapshot_test.go`、`glm_shared_callbacks_test.go` 的 model 分支为历史、目标或回调快照，不是目录准入；保留。
- `catalog_candidate_source_test.go` 使用未实现 API 判 code_gap；`xai_images_catalog_test.go` 负例为能力/价格；`local_catalog_test.go` 未来名称伴随未知字段；均保留。
- 其余匹配的 JSON `id` 大部分是 response item/tool call、grant/fixture case ID 或固定正式原件；不得以运行准入变更为由修改其真实性。
