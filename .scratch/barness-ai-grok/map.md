# Grok 实施工单图

## Notes

集成分支：`codex/barness-ai-grok`。契约：[spec](spec.md)；唯一当前状态：[spec](spec.md#唯一当前状态)。

依赖：01 → {02, 03, 07} → 04 → 05 → 06。每工单在独立分支/工作树实现，合并到集成分支。

## Decisions-so-far

- 01–04 交付候选实现与普通验证；05/06 按 SOP 另取 live/race/压力授权，未满足前不能关闭或宣称正式支持。

## Fog

- 真实流形状、非零 reasoning 算术、工具结果图片投影、request-ID 与实际返回型号仍需本路线真实证据。
