# barness-ai 未发布变更

### 2026-11-04：Anthropic chat 仅保留 4.8 及以上型号

按维护者要求，`chat / anthropic / anthropic-messages` 保留 Fable 5、5.1，
Opus 4.8、5、5.5，Sonnet 5、5.5，共 7 条身份。Haiku 当前无满足门槛的型号。
逐条下架 `claude-haiku-4-5`、`claude-haiku-4-5-20251001`、`claude-opus-4-5`、
`claude-opus-4-5-20251101`、`claude-opus-4-6`、`claude-opus-4-7`、
`claude-sonnet-4-5`、`claude-sonnet-4-5-20250929`、`claude-sonnet-4-6`，共 9 条。
全部执行 CLI withdraw 预览/确认、check、apply 预览/确认；旧定义和追溯保留。
总目录 88 → 79 条（73 chat、3 image、3 classifier）。Anthropic 内容版本 1 → 10。
其他 Provider 版本及全部保留定义、价格、相对顺序不变。

| 标识 | 值 |
| --- | --- |
| Catalog.Version | `catalog-043a11039c792ad6f448e840652aa4776830f57732abddb80e6fe874027b2ec5` |
| Catalog.Hash | `sha256:40d4c5fe7551fb06e79210b676e69e96a47b6b1b200dde5f6c07e243f970b2f0` |
| JSON 文件 SHA-256 | `44d7caa8f36ce1b5b7914d014215be1e3625b453677cabe54d891b59ac5a9bb8` |
| 维护 revision | `sha256:2047fad03a8020c866481cb8549e9e0a28f4a0a767646e65f4e13b41a381349f` |

**代码与分发影响：** 比较基准/当前 HEAD 为 `665862ff90074eecda0a573050114de9ce88c80d`，
实际受测为本次工作区。SDK 核心运行代码、运行依赖和 schema 均无变化。
维护源、完整审阅基线、生成 JSON、相关 E2E、验证工具型号和示例配置有变化。
正式验证默认 Haiku 4.5 调整为 Sonnet 5.5；历史报告及原测量不改写，不授予新真实资格。
ChineseEvaluation 新建 v7，仅变更配置版本与目录 pin；v1–v6 原字节保留。
外部加载使用者可向本次实际受测程序分发原 JSON，再显式加载并创建新 Client；
现有 Client 不热更新。任意其他旧程序版本兼容性未验证。
编译内置目录使用者需要重建并分发程序，才能取得本次默认目录。

**验证：** 9 次候选检查全部 PASS/exit0；公开 Client 下架 E2E 从预期 RED 转为 PASS。
完整受影响普通回归及最终同源生成/构建/vet 的结果由完成索引记录。
live、race、压力按纯下架影响不适用；软件发布、部署均未执行。
[唯一当前状态](verification/status.md) ·
[本次交付](delivery.md)。

完整普通回归实际 PASS/exit0；两段并集覆盖全部 934 个已发现顶层测试，固定必需场景 54 项 PASS，原生 audit 均为 0。最终同源生成、构建、vet 和 live 标签编译/vet 均 exit0。
