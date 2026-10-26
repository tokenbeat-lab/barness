# 身份、候选目录与前置授权

Type: task
Status: resolved
Blocked by: none

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R01–R04、R21；E01/E02/E04。公开 ProviderXAI；新增可明确装配的候选 Catalog（不激活 BuiltinCatalog），固定 grok-4.7 与全球目标。校验 Host 元数据只能收窄且非法组合在凭据前拒绝。

## Ownership

ai/binding.go、ai/catalog_xai.go、ai/catalog_validate.go、ai/call_runtime.go 等身份/目录/路由文件；对应 xai_foundation E2E。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

实施提交 `04474e07`，后合并最新集成分支；候选实现完成，未激活正式目录。

确认工作树 HEAD 为集成基线 `2671dacea14ea602f0f5bb1f96bbb2a697ebb372` 且包含集成分支。

先失败方式：自带目录扩大型号/协议/上下文/输出/模态/推理映射/默认档位/其他协议 compat；Binding 的其他操作、协议、型号或目标误读凭据；候选污染正式目录或可变快照。接缝仅公共 Catalog.Hash/NewClient 与 Client.Complete，所有调用观察凭据读取及 HTTP。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。

### 交付与回执

- 公开 `ProviderXAI`；`Catalog.Hash`/`NewClient` 收敛唯一 Grok Responses 元数据，允许 Host 收窄模态/容量/档位及独立价格。非聊天目录拒绝 xAI。
- `resolveBinding` 在凭据前检查聊天/Responses/全球标准目标；规范化末尾斜线、443 端口、主机大小写；没有 localhost 绕过。型号仍取 Binding 白名单与目录交集。
- `ai/internal/testkit/xai/candidate.json` 是独立候选快照；`xai.CandidateCatalog()` 为受控与有限授权 live 装配，正式 `BuiltinCatalog` 未改。Host 可加载该 JSON 到已有 Catalog 接口。
- 共用 E2E helper：`newXAIWorld(t, mutateCatalog, configure...)`、`xaiBinding()`、`xaiTarget()`、`xaiRequest()`、`xaiFrames()`、`xaiReply(t, frames...)`；Transport 保留官方 Binding，只在注入接缝转发至受控服务。
- 全部命令通过 `tools/verification/run.py --feature barness-ai-grok --stage e2e -- go test -json ./ai/e2e -run '<精确表达式>' -count=1`；实际 argv 在各原件 command.json。最终表达式为 `^TestXAI(CatalogBoundary|RejectsOtherOperationCatalogs|CandidateSnapshot|HostNarrowsCandidate|RouteRejectedBeforeCredentials|OfficialTarget)$`。
- Red（退出1）：`.evidence/barness-ai-grok/e2e/20261026T015908.655960Z-ch1yuy3j`（目录越界），`20261026T020024.989829Z-tw64wee3`（目标），`20261026T020124.131287Z-6ltb4vkd`（其他操作目录），`20261026T020246.875535Z-pi9zcund`（空端口/空fragment）；同级独立 PASS 重跑均留存。
- 一次验收原件 case ID 冲突：`20261026T020216.255583Z-no10q27e` 为 INVALID_EVIDENCE（子进程退出0）；已修复测试标识并保留失败原件。
- 最终定向普通回执：`.evidence/barness-ai-grok/e2e/20261026T020259.157917Z-yqsxii7a/receipt.json`，采集器/Go 实际退出0，PASS；唯一 `./ai/e2e` bundle、audit 及实际场景清单在该回执。它绑定实施前 HEAD 与完整 dirty 执行文件哈希；集成最终冻结后另取普通回执，不冒称此回执为正式来源。
- 未运行 live/race/压力；后续由功能唯一当前状态按 SOP 记录，不由此工单授予支持资格。
