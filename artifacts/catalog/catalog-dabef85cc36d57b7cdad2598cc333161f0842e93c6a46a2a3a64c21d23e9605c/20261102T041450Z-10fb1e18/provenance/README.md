# 来源与验证副本阅读说明

本批次运行 JSON 和 review 按批准旧包原字节复制，并非新导出；review 中的历史 checks 不能代表本次型号准入策略。`operation-v1/index.json` 与 `final-execution-v1/mapping.json` 记录原路径、原件与 SHA；完整 case 胶囊留在仓库原路径，包并非自包含验证胶囊。执行脚本副本不是原执行位置。

`changelog-source-final.md` 保存最终 Changelog 来源原字节；其相对链接按原路径 `docs/barness-ai/CHANGELOG.md` 解释，不作包内导航。包内可阅读 [变更说明](../CHANGELOG.md)、[交付说明](../delivery.md) 和 [验证结果](../verification/result-v1.json)；最终完成以 `delivery.json` 为准。

[封存前状态快照](../verification/current-status-before-seal.md)保留验证完成、交付尚待最终审查时的原字节。仓库唯一当前状态为 `.scratch/catalog-model-id-policy/status.md`；快照不持续更新，交付完成以 `delivery.json` 为准。
