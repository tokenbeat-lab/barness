# Claude Haiku 5.5 目录交付

仅新增 chat / anthropic / anthropic-messages / claude-haiku-5-5。
目录 80 条（74 chat、3 image、3 classifier），Anthropic 8 条、内容版本 11。原 79 条定义与价格保持，旧身份不恢复。
文本/图片、1M context、128k output，已有托管 adaptive effort 的五级强度和 prefix drop_block；禁发 temperature。
本库不声明 PDF、关闭推理、Batch 或原生浏览器/电脑工具；metadata 标注服务默认 medium，既有 SDK 托管省略 effort 策略仍为 high，调用者可显式选 medium。
基础 input/output/cacheRead/cacheWrite：0.1/0.5/0.01/0.125 USD/百万 token。
全部提示词 token（含缓存读写）严格超过 100k 后整请求为 0.5/2.5/0.05/0.625，1h 写为选定 input 的 2 倍。

版本：`catalog-73ff88ca97d69b37082bdda7ce5ab2cbd70fab98a8bb2483a4c3e749d3da8402`。Catalog.Hash：`sha256:d7ca061de668054669d4721038fa9cb29c94103815a8d2a318e9ee0025df9703`。
完整 JSON SHA-256：`f29ff70c8b0a70469ee3b924c156e3c33b0bda079a8dbd1de91b25029d59b468`。维护 revision：`sha256:8e46300c8052478e57982ed633b0682c6e2be4bd0ac9a5247653f9efae6e3c9e`。
[运行 JSON](catalog-73ff88ca97d69b37082bdda7ce5ab2cbd70fab98a8bb2483a4c3e749d3da8402.json) · [变更记录](CHANGELOG.md) · [普通验证](verification/ordinary-summary.json) · [候选检查](verification/candidate-checks.json) · [状态](verification/status.md)。

HEAD/实际受测 HEAD：`665862ff90074eecda0a573050114de9ce88c80d`，起点包括上一轮未提交的 Anthropic 下架。实际基准另由 before-source/before-builtin 固定。
本次 SDK 核心、运行依赖、schema 变化均 NO；仅维护数据/完整基线、相关 E2E、示例 v8 和文档变化，正式验证目标未改。
models.dev 的 PDF 映射实际被拒绝；改用官方原捕获与完整人工定义/决定，经 add/check/apply、导出、三处同步后归集。
候选、完整受影响普通、生成、构建、vet 和 live 标签编译/vet 均通过。原编译失败、预期 RED 和 launcher 启动失败均保留。

外部加载：仅本次受测程序取得兼容依据，可只分发原 JSON，Host 显式加载并创建新 Client；现有 Client 不热更新。其他旧版本未验证。
编译内置：更新默认目录须重建并分发程序。live/race/压力按影响不适用；未发布、未部署，不产生新厂商实测资格。

review/ 与采用 baseline 相等；补充官方文档位于 verification/supplemental-sources。测试胶囊、CLI、未改写的 models.dev 响应和工具仍在外部原路径，见 external-originals.json/候选索引，不声明全部证据自包含。运行 JSON 可独立使用。
恢复前置目录需 withdraw 新身份并重新 check/apply/export、同步三处和另建配置 pin；旧历史不重写。本次未执行回滚。
