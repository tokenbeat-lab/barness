# OpenAI chat 目录下架交付

已完成 37 条接口身份下架及内置采用。仅保留 GPT-5.6、GPT-6 及以上当前 8 个型号；
Responses 8 条、Chat Completions 7 条。目录总计 88 条，OpenAI Provider 版本 39。

版本：`catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2`。Catalog.Hash：`sha256:9dff935e34e42b96adc8afeb34c2db689447c936f1c584928ea08b3ed0a7b65a`。
完整 JSON SHA-256：`e706fdc1dd13aed8799478da342467a810005e769420abce1ed0493dd8c0d9e6`。
维护 revision：`sha256:f5bf85ac328703c01a4019bebec41c5b779b1601b2f02eee0ae533881ed99a3e`。

[版本化运行 JSON](catalog-21027c1121f250321e83789b00daa0b4573b372f72258256da1d0378d21fa5b2.json) · [变更记录](CHANGELOG.md) ·
[普通验证](verification/ordinary-summary.json) · [候选检查](verification/candidate-checks.json) ·
[数据核对](verification/data-audit.json) · [当前状态](verification/status.md)。

比较基准、目标 HEAD 和实际受测 HEAD 均为 `b041eff5ca74393a88f66916b13ce1aaeec12ee8`，本次修改尚未提交；
实际字节由普通回执固定。核心运行代码/运行依赖/schema 变化均 NO。
生成数据、维护源、完整基线、相关 E2E 装配、示例配置 v6、验证工具默认目标和文档有变化。
验证工具删除已下架 pro 的厂商探针；旧报告/测量保留，不产生新型号真实资格。
37 次候选检查、完整受影响普通、生成、构建、vet 均实际通过；历史失败保留。

外部加载：在本次实际受测 b041eff 程序上可只分发原 JSON，Host 显式加载并创建
新 Client；现有 Client 不热更新。其他旧程序版本未验证。
编译内置：需要重建并分发程序以更新默认目录。

未运行厂商调用、race、压力、软件发布或部署。前三项按影响不适用；
未声明新厂商资格或完整产品发布。
完整审阅包位于 review/，其文件库存和原字节已经核对。原始测试胶囊、CLI 操作、
二进制及检查原件保留在 external-originals.json 和检查索引指定的外部路径；
本交付并非全部验证证据的自包含归档。运行 JSON 本身可独立分发。

恢复通过新的 restore/rollback 候选重新 check/apply，再导出及采用整组来源、
快照和基线。本次未执行恢复；所有下架定义和历史保留。
