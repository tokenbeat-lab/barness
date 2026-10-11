# 02：迁移唯一 pi oracle 和基线校验到 1.1.0

Status: resolved
Type: task
Depends on: 01
Updated: 2026-10-11

## What to build

使用工单 01 的已验证发布来源迁移唯一 oracle；删除旧运行依赖和 guard，更新版本、commit、model-data hash、provenance、差分账本和发布基线校验。处理保持既有公共契约所必需的增量兼容修正，不实现与采样无关的新产品能力。

## Ownership and context

负责 `ai/internal/testkit/pioracle/`、差分账本、发布差分记录校验及其受影响 fixture；兼容修正仅限 01 inventory 中已明确归属的模块。现有 ModelPatch 已能携带新字段，不新建采样专用 runner 或第二套 oracle。协作时不要回退他人修改。

依据：[规范 R10](../spec.md)、[设计 §6](../design.md#6-冻结基线与目录迁移)、[ADR-0025](../../../docs/adr/0025-barness-ai-thinking-level-sampling.md)。正式主规范与 ADR-0004 的基线生效说明由工单 06 同步，此阶段处于同一未发布实现分支。

## Failure scenarios first

- runner 用 1.1.0，但 release validator、SDK 版本记录或数据 hash 仍校验 1.0.0。
- 安装解析到浮动版本、启动时接受错版本/commit/hash，运行依赖外网或环境凭据。
- 保留双 oracle，按场景偷偷选择旧基线；删除 payload 字段或放宽数字比较来隐藏差异。
- 原有 extension 失去依据或被扩大 case 范围；未登记差异被当作 PASS。
- 新采样失败被当作迁移完成，或以完成本工单为由宣称正式零 pending。

## Acceptance criteria

- [x] B02：唯一 frozen 1.1.0 安装可按 lock 重建；来源和逐文件证明可追溯；运行 guard 校验与验证报告一致。
- [x] 错 version/commit/model-data hash 用受控输入验证拒绝；沿用空环境、无外网、loopback 和凭据隔离，不新增启动兼容路径。
- [x] 全部原有差分子集执行一次，逐项核对工单 01 inventory；无未解释差异，记录 actual artifacts。
- [x] 非采样的迁移必需差异先列公共失败场景，再做最小修正；旧已修复项不退回 extension。
- [x] 新 profile/采样造成的预期失败显式登记 pending 并关联工单 03/04/05；工单 02 之外仍有 pending 时，发布保持失败。
- [x] 未解决的其他兼容问题有具体阻断和工单，不把新能力一律声明为本 feature 扩展；所有旧注册路线和豁免范围保持可审查。

## Verification and handoff

运行完整开启差分的 E2E，保存请求/结果/发现清单和审计；记录此阶段是“oracle 施工迁移”，不是 B03 的最终零 pending。隔离判定若必须测试，先列失败方式再实现，不补写镜像实现的单测。

交给[工单 03](03-catalog-profiles.md)验证后的发布元数据和精确 pending；交给[工单 05](05-isolation-and-differential.md)全量差分结果；交给[工单 06](06-document-and-release.md)全部基线 pins 与待更新正式文档清单。

## Comments

2026-10-11：方案阶段创建，当前正式 oracle 仍为 1.0.0。

2026-10-11 实施：单一 1.1.0 lock/guard/provenance 生效于实现；完整原子集和新增场景共 814 差分 PASS、pending 0。duration 字段及有限耗时差异获用户明确批准，三个预算差异 fixed。 命令、实际 artifacts、失败历史和未完成项见[专项证据](../evidence.md)与[来源核验](../release-verification/README.md)。
