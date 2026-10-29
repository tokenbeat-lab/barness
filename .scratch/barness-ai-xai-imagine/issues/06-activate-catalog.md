# 激活xAI图片内置目录

Type: task
Status: resolved
Blocked by: none

## Scope

用户于2026-10-29明确要求“激活xAI目录”。当前上下文为grok-imagine-image-2.0图片路线；Grok4.7聊天目录已经存在。内置图片目录限定自身真实六场景实测：最多5参考图/2输出，1k/1.5k/2k，low/medium/auto，1:1/16:9；单图编辑仍省略比例、编辑单输出。按张价目表保持，旧候选JSON/版本/hash与原grant/捕获/失败均不改写。

## Verification

用户已确认的公共NewClient/Catalog/GenerateImages边界先红后绿。新增默认Catalog可发现、价格/能力精确、目录深复制与版本/hash、新默认Client逐场景重放六份原始捕获且最终wire相等、超出当前内置上限预拒绝。无需厂商请求或真实凭据。验证相邻OpenAI/Google图片、xAI聊天、目录快照和相关普通包。由实际影响选择，不自动运行race/压力/新live。

## Evidence

原生真实candidate：`.evidence/barness-ai-xai-imagine/live/xai-images/20261029T012001.157782Z-n96g7rg4`，PASS/exit0，7图与USD0.51，源724e9881；不可变候选索引candidate-live-20261029-03.json。本次激活是内置目录源码接线，不把旧报告stage/mode/formalQualified重写，也不宣称race/压力或完整发布通过。

## Answer

受限内置目录、价格与根快照在`cec2bd09`实现；当前/历史版本测试耦合在`43821cb6`修复；中文评估新v2当前配置在`9549cf4c`修复；原ToolRoundTrip测试时序在`7d4edddd`修复，仅固定运行中的读取时刻。集成`0219357c`。能力、边界、适用性及回滚见[激活范围](../catalog-activation.md)，当前状态仅见[spec](../spec.md#唯一当前状态)。

公共入口红绿和失败原件均保留。最终冻结普通命令覆盖`./ai/e2e ./ai ./ai/examples/chineseeval ./ai/examples/chineseeval/cmd/chineseeval`，PASS/exit0，2058个测试事件，墙钟127.374s。原件`.evidence/barness-ai-xai-imagine/e2e/20261029T023159.242880Z-w8ibgpxm/receipt.json`，SHA256 `1912803cdef936263a0d3dc14068822850a20f110388f5bfe4ca98f18146e40f`。有限完整主包PASS集合及其他bundle/audit固定在record-20261029-04.json；未跑case不计通过。

两个轴的审查与修复复核见[目录审查](../review-catalog.md)，遗留行动问题0。原候选目录/6份捕获/授权/claim/FAIL保持。此次源码激活不提升正式原生资格，不新增厂商调用、race、压力或完整发布事实。
