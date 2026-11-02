# 最终交付 metadata 独立 Spec 复核

发现两项待修正：

1. 唯一当前状态 `.scratch/catalog-kimi-k31-pro/status.md` 仍写「完整普通 v3 当前运行」「目前没有终态」，但同文最终验证结果已记录 PASS / exit 0；首行状态也停在待重验。SOP 要求唯一当前状态与真实阶段一致，须改为历史时间描述及当前结果。
2. `provenance/changelog-source-final.md` 是原 changelog 原字节快照，其相对链接 `[本次状态](../../.scratch/catalog-kimi-k31-pro/status.md)` 在交付目录内无法解析。保持原快照字节时，应在交付文件提供正确的状态入口并明确快照相对链接按原位置解析，避免误导；如更改快照，须重新绑定摘要。

其余已审范围未发现不一致：六项检查、原件路径/摘要、普通回执摘要、125 条及旧 119 项、内外快照与完整 baseline、代码/依赖和旧程序 120 条探针范围、203 必需场景、226 NOT_RUN、来源差异及固定汇率表述均与索引一致。完整胶囊留外部原路径；PROCESS/draft 为历史。未运行测试，delivery.json 尚待审查完成后封存。
