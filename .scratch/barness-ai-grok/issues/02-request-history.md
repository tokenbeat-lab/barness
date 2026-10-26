# 无状态请求、选项、最终守卫与可信历史

Type: task
Status: resolved
Blocked by: 01

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R05–R10、R12/R13、R16、R19；E02–E04/E07–E09/E13。复用 Responses profile，明确最终字段子集；预算/采样前置；完整 transcript、opaque ID、来源降级与严格图片。

## Ownership

Responses 请求/选项/profile、xAI request/guard/history/ID codec 及对应 E2E；不拥有流与 usage 文件。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

已实现本票请求与历史范围：独立 xAI Responses profile；默认/显式 effort 的无状态完整请求；精确预算/采样与凭据前 preflight；Hook 后无条件授权及字段子集守卫；完整 native reasoning JSON、`xar1|` opaque 二元组、来源降级；严格 PNG/JPEG 与 20 MiB/Host 限额；工具结果图片紧随自身结果；派生缓存与回调顺序。无公共签名、Client 接口或依赖新增，旧路线配置分支未改动。

实际受测提交为 `119ddcc10aaf74303146d1beff7c48e925f69597`，已包含最新集成基线 `815bb3a00e35d8179de34ff51b328e1909aec015`。最终命令：

```sh
python3 tools/verification/run.py --feature barness-ai-grok --stage e2e -- \
  go test -json ./ai/e2e -run '^TestXAI' -count=1 -timeout=5m
```

退出 **0**；唯一主包 `./ai/e2e` 的 **290** 个原生 case 全部通过，audit findings 为空。最终[receipt](../../../.evidence/barness-ai-grok/e2e/20261026T022404.466029Z-kkfrh3ui/receipt.json)（SHA256 `4ea66ed5d77c958c6711d6ecdf919ae4850be2a45c3c2f6de4ae64458a65c970`）；[manifest](../../../.evidence/barness-ai-grok/e2e/20261026T022404.466029Z-kkfrh3ui/bundles/20261026T022405.583327000Z/manifest.json) SHA256 `fe2e8fe37064b531eec8f1c2beeb387834e57553bc9fa9d37dfd9fe6e5b35117`；[audit](../../../.evidence/barness-ai-grok/e2e/20261026T022404.466029Z-kkfrh3ui/bundles/20261026T022405.583327000Z/audit.json) SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。

每个 red/green attempt 均保持原件：

| 切片 | Red（退出1） | Green（退出0） |
| --- | --- | --- |
| E02/E03 请求与最终守卫 | `20261026T020937.330364Z-idzqaqjy` | `20261026T021022.970912Z-q1m84gqs` |
| E07/E08 原始关联与可信历史 | `20261026T021157.893937Z-213r7i3n` | `20261026T021300.389207Z-m7aluumb` |
| E04 命名选项、sampling、Full/Simple | `20261026T021343.260558Z-i_fixagi` | `20261026T021433.782114Z-n4k3rb69` |
| 输入/工具子集、Hook native 授权 | `20261026T021417.192811Z-86zx478z` | `20261026T021506.829967Z-5vtgfffv` |
| E09 图片与工具结果顺序 | `20261026T021552.751768Z-dy4yvcr5` | `20261026T021615.417252Z-9rddnwly` |
| native 精确大数保留 | `20261026T021917.039156Z-iqswfpqw` | `20261026T021937.686201Z-_qi3oxq_` |
| null ciphertext 回放一致性 | `20261026T022304.761950Z-z5_yzjdt` | `20261026T022350.278316Z-pi_gn9m0` |

以上均位于 `.evidence/barness-ai-grok/e2e/<attempt>/receipt.json`。额外通过的缓存/回调验收为 `20261026T021738.243127Z-qb3dtqke`、Provider 缓存隔离 `20261026T022019.289011Z-7p5d0laj`；中间组合验收为 `20261026T021826.118611Z-aqiu1cf9` 与 `20261026T022045.066600Z-zq0lcdhp`，均退出0，未覆盖或修改既有失败。

本票仅完成受控公共 Client/Catalog E2E 范围；最终集成的旧 Provider 回归与完整受影响普通包由集成阶段执行。live、race、压力均 **NOT_RUN**，无新增权限；正式目录/支持资格与发布结论未改变，`release.pass=false`。流终态/用量的 R12 部分归 ticket03；功能唯一当前状态仍由 [spec](../spec.md#唯一当前状态) 统一维护。

源码已分职责拆为低于500行文件。回滚可撤销本分支 xAI 请求/历史提交，并由 Host 禁用对应 Binding；不删除历史原件或价格快照。本票文档收尾与受测执行来源相等性另保留在[核验原件](../../../.evidence/barness-ai-grok/ticket02-source-equality.json)。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。

- 2026-10-26：request_history 在 `codex/grok-request` 负责请求、原生历史与图片边界；当前进入独立 E2E 回执收尾，未修改功能唯一当前状态。

- 2026-10-26：已完成并解析为 resolved；公共接缝已由 spec 预先确认，依 tdd 垂直切片留存失败再实现。与 ticket03 对齐 `encrypted_content:null` 的原样回放，id/status 仍在存在时限定字符串。
