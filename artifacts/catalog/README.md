# Catalog CLI 交付物

每次交付位于 `<Catalog.Version>/<UTC批次>/`，运行文件使用
`<Catalog.Version>.json`。同一版本的不同追溯/验证批次也使用独立目录，不覆盖历史。
具体步骤见 [Catalog CLI SOP](../../docs/barness-ai/catalog-maintenance.md)。

## 已归集交付

| 目录版本 | 批次 | 入口 |
| --- | --- | --- |
| `catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6` | `20261102T021202Z-b9f7df2a` | [运行 JSON](catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6/20261102T021202Z-b9f7df2a/catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6.json) · [交付索引](catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6/20261102T021202Z-b9f7df2a/delivery.json) · [代码变化与分发判定](catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6/20261102T021202Z-b9f7df2a/delivery.md) |

本次为 GPT-6.1 Terra 的已批准目录：119 条，OpenAI 内容版本 2，原 118 条定义及
相对顺序保持。核心运行代码相对新增前基准 `37a283ee` 没有变化；有外部加载能力且
通过检查的程序可以只接收 JSON，编译内置数据的程序需要重新构建才能取得新默认。
具体受测程序、比较提交和实际证据以该批次 `delivery.json` 为准，不推断旧版兼容。

## 文件职责

- 版本化 JSON 与 `.sha256`：可独立分发的原受测运行字节和完整文件摘要。
- `review/`：完整原审阅包，包含维护来源、revision、snapshot 和所有已索引输入。
- `CHANGELOG.md`：随本次交付冻结的自包含变更说明；仓库汇总记录在
  [Changelog](../../docs/barness-ai/CHANGELOG.md)。
- `delivery.md` 与 `delivery.json`：核心/依赖变化、其他代码变化、两类分发结论，
  以及全部交付文件库存。`delivery.json` 最后写入，是本批次完成标记。
- `verification/`：已执行检查和验证摘要；原测试胶囊仍在索引所指的不可变路径，
  当前目录不是自包含的全部验证胶囊。目录 JSON 本身不依赖这些证据文件运行。

`work/` 存放显式选择的工作库、原输入、工具、日志与操作脚本，Git 忽略该目录。
已完成的版本化交付目录保留入库。历史 `.catalog/`、无版本本地 JSON、`.evidence/`
及源码采用位置不移动或重写；新操作按 SOP 显式选择统一根目录下的路径。
目录接纳、软件发布、重建分发与厂商真实资格是分别记录的事实。
