# Grok 实施工单图

## Notes

集成分支：`codex/barness-ai-grok`。契约：[spec](spec.md)；唯一当前状态：[spec](spec.md#唯一当前状态)。

依赖：01 → {02, 03, 07}；{02, 03} → 04；{04, 07} → 08 → 09 → 05 → 06。07为独立候选入口兄弟工单。每工单在独立分支/工作树实现，合并到集成分支。

## Decisions-so-far

- [01 已解决](issues/01-foundation.md)：`04474e07` 交付候选/目录/目标授权；保留公共 E2E red/green，正式目录仍未激活。

- [04 实现已解决](issues/04-host-verification.md)：`c924c311`补齐48个公共Host/runtime及验收断言，定向普通回执PASS；已有生产实现通过，未虚构行为Red。全局两轴审查、完整来源冻结/普通包与后续资格仍由集成阶段完成。04不依赖07，真实资格05等待两者汇合。

- 01–04 交付候选实现与普通验证；05/06 按 SOP 另取 live/race/压力授权，未满足前不能关闭或宣称正式支持。

- [08 审查修复已解决](issues/08-review-fixes.md)：`51c95922`修复两轴审查五项发现，逐项保留公开 red/green；冻结定向范围605测试/564case通过。独立复核与完整集成普通回执仍待完成，05还需后续冻结及授权。

- [09 规范证据标签](issues/09-canonical-evidence.md)：真实原生消费者发现空 effort 产生中间空目录段；标签改为default并加实际producer→collector→CLI回归。旧错误原件保留，05等待新的集成冻结。

## Fog

- 真实流形状、非零 reasoning 算术、工具结果图片投影、request-ID 与实际返回型号仍需本路线真实证据。
