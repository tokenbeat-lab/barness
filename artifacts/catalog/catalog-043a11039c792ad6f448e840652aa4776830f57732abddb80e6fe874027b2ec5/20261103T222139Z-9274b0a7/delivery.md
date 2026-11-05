# Anthropic chat 下架交付

已完成 9 条旧身份下架及内置采用。保留 Fable 5/5.1、Opus 4.8/5/5.5、Sonnet 5/5.5，共 7 条；当前 Haiku 无符合门槛的型号。
目录共 79 条（73 chat、3 image、3 classifier），Anthropic 内容版本 10；其他 Provider 和保留定义不变。

版本：`catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5`。Catalog.Hash：`sha256:40d4c5fe7551fb06e79210b676e69e96a47b6b1b200dde5f6c07e243f970b2f0`。
完整 JSON SHA-256：`44d7caa8f36ce1b5b7914d014215be1e3625b453677cabe54d891b59ac5a9bb8`。维护 revision：`sha256:2047fad03a8020c866481cb8549e9e0a28f4a0a767646e65f4e13b41a381349f`。
[运行 JSON](catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5.json) · [变更记录](CHANGELOG.md) · [普通验证](verification/ordinary-summary.json) · [候选检查](verification/candidate-checks.json) · [状态](verification/status.md)。

比较基准、目标 HEAD 和实际受测 HEAD 均为 `665862ff90074eecda0a573050114de9ce88c80d`；本次工作区改动尚未提交，实际字节由回执逐文件固定。
核心运行代码、运行依赖、schema 变化均 NO。维护数据、完整基线、相关 E2E、示例配置 v7、验证工具型号及文档有变化。
正式验证默认改为 Sonnet 5.5；历史测量保持，不产生新的厂商资格。
9 次候选检查、完整受影响普通、生成、构建、vet 和 live 标签编译/vet 均实际通过；预期 RED 原件保留。

外部加载：仅本次实际受测程序取得兼容依据，可仅分发原 JSON，Host 显式加载并创建新 Client；现有 Client 不热更新。其他旧程序版本未验证。
编译内置：必须重建并分发程序，才能更新默认目录。
live、race、压力按本次影响不适用；软件发布、部署未执行，不声明完整产品发布。

review/ 为完整审阅包，与采用基线逐文件相等。测试胶囊、CLI 命令/二进制和检查原件仍位于 external-originals.json 及候选检查引用的原路径；本交付不是全部验证胶囊的自包含归档。运行 JSON 可独立使用。
恢复须创建新的 restore/rollback 候选，再 check/apply 并同步来源、快照和完整基线。本次未执行恢复，旧定义和历史保持。
