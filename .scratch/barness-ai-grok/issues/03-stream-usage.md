# 权威流快照、工具合同、用量与估算成本

Type: task
Status: resolved
Blocked by: 01

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R11/R12、R14/R15、R17/R18/R20；E05/E06/E10–E12。xAI 私有状态绑定 Responses adapter，严格关联、done/final-only、重复与冲突、真实终态及原始数字；独立成本完整性。

## Ownership

Responses 流接线、xAI stream/usage/cost 文件及对应 E2E；不拥有请求/history 文件。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

已实现并提交至 `codex/grok-stream`；核心变更提交 `9b21b4ed`、`672d7cd9`、`9f059499`、`d1b80111`。已合并请求/history 的 `119ddcc1` 与最新集成 `8e70cc49`，不修改正式目录或 canonical spec 状态。

- Responses 读路径仅对 xAI 使用每调用独立的原始 SSE 状态；保留 output/part 索引、item/call ID、完成状态与原始 reasoning JSON。支持 done-only/final-only、双推理通道、独立内容分段和终态密文回填；完全重复快照不重复发事件，冲突/孤立 delta/丢失或错误索引/终态后新内容均 protocol 失败。
- 函数调用使用请求/history 工单的 `xar1|` opaque codec；正常 completed 严格兑现 required/none/parallel:false 和完整 JSON object。截断的 function 保留 raw 参数与 length 状态，不发 ToolCallEnd，不将显示修复当作可执行参数。
- 原始用量区分 missing/null/合法0，严格检查精确整数、边界、加减溢出与关系。P/T 确定全部输出 B，接受 B=O 或 B=O+R，缓存从 P 分离，推理仅作 B 子集；缺总量不补造 TotalTokens，缺项保持 partial。
- 成本使用本次 Catalog 快照及既有 tiers；P 含缓存且 ≥200000 时全量换档。未知缓存按全部普通输入/全部缓存的端点计算有界范围，计数保持未知；Usage、Attempt 与 Observer 同步保存成本及完整性。failed usage 沿 ADR-0010 保持 unreported。
- xAI 流/回调失败保留已收到的 HTTPStatus，request-ID 头在自身捕获确认前保持空；初始 HTTP 与流内上游错误仅输出稳定安全诊断。旧 Provider 分支与公共接口、依赖未变，新增生产文件均低于500行。

### 验证与原件

公开接缝为已确认的 Client/Catalog，经受控 HTTP/SSE 服务验收。每轮命令均由 `tools/verification/run.py --feature barness-ai-grok --stage e2e -- go test -json ./ai/e2e -run '<精确表达式>' -count=1 -timeout=5m` 执行；实际 argv、数字退出及场景清单见各独立 attempt。

- 可复现行为 Red（退出1）：`20261026T020934.028118Z-rfh_5prr`（final-only 丢内容）、`20261026T021100.156590Z-k6337vaf`（快照/推理合并）、`20261026T021327.682193Z-rwfbawju`（工具）、`20261026T021457.187155Z-w_012giv`（HTTP状态/请求ID）、`20261026T021739.373250Z-jea7q8j3`（原用量路径少计推理、缺成本区间）、`20261026T021859.723287Z-b2atzf02`（推理权威冲突）、`20261026T022008.198374Z-9piptxjr`（HTTP回显）、`20261026T022219.111549Z-r2bcokx_`（截断工具）、`20261026T022412.213092Z-gclt1zyb`（结构）、`20261026T022623.498731Z-5cvpbo6g`（completed矛盾字段）。各轮独立 Green 均留存在同级 `.evidence/barness-ai-grok/e2e/`。
- 测试基础设施失败同样保留：`20261026T021443.609573Z-yuqhgc4n` 是测试编译错误；`20261026T021606.542630Z-gjs35s12` 与 `20261026T021644.752322Z-yjtba5e8` 是 Observer 原件文件名含路径分隔符，不能作为用量行为的 Red。修正标识后临时回到原用量路径，取得上述 `021739` 行为失败，再恢复实现取得独立 Green；未覆盖原件。
- 最终干净源码普通回执：[receipt.json](../../../.evidence/barness-ai-grok/e2e/20261026T022734.642299Z-a17jb0sl/receipt.json)，命令 `go test -json ./ai/e2e -run '^TestXAI' -count=1 -timeout=5m`，实际退出0、PASS，400个原生 case，唯一 `./ai/e2e` bundle 与 audit 通过。受测 SHA 为 `d1b80111257b075f2cd834aa1f7957185c2c96f2`，receipt SHA-256 为 `819b99a310d0be02cad699e7250892e7275d625fd7216cac522d90812aa818c8`。
- 原生 bundle `bundles/20261026T022736.029602000Z`；manifest SHA-256 `929a1099ec1afbbb0b95b2c6fecad5362926585186555ebd5ee8cd45b025ea5f`；audit SHA-256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。
- 随后集成合并及本文收尾仅涉及 Markdown；[完整执行来源/选定输入集合相等性](../../../.evidence/barness-ai-grok/ticket03-final-source-equality-20261026T0230.json)绑定最终提交与受测 SHA，未重写旧回执。
- 未运行完整受影响包、live、race、压力；由集成负责人按最终冻结来源与 SOP 另行处理。候选普通验证不产生正式支持或 release.pass=true。回滚为撤回 xAI Binding/候选装配并回退本工单实现提交，既有 Provider 合同保持。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
