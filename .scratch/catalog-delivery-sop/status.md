# Catalog CLI 交付规范状态

状态：首轮独立审查指出两处范围/不可覆盖缺口，已修订 SOP 并生成独立新批次，待复核。起点 `4831e3544945f667d9a817be8dfb92dd08856575`。

范围：维护 SOP、changelog、统一交付目录说明及当前 Terra 的版本化交付归集。
没有 SDK、CLI 或协议运行代码变更。既有未跟踪 `.catalog/`、`local-catalog.json`
和 `.scratch/provider-integration-status-20261102/` 保留。

普通验证：交付库存/文件摘要、真实 Catalog.Version 与文件名、目录原字节、
review 完整性、changelog、本地引用及代码变化报告；公共 cataloggen --check。
原完整普通回执保持为
`.evidence/catalog-gpt61-terra-builtin/e2e/20261102T013403.060605Z-n5h34x6a/receipt.json`，
实际受测 `d4dbafaa1537ada3f6a56ade817dc63ad20d9f9d`；收尾核对 2,159 个执行来源
和 87 个选定输入的完整集合及字节，没有新模型/协议行为测量。

live/race/压力：NOT_APPLICABLE。仅流程文档和原字节归集，无新厂商事实、共享状态
或设计负载变化。本次不作软件发布或厂商资格结论。

原始 3.3 GiB 验证树和历史失败包按测试 SOP 原位保留；交付目录放入验证摘要及
包含原件路径/摘要的索引，明确不是自包含的完整验证胶囊。运行 JSON 可独立分发。

## 本次归集

批次：`20261102T021202Z-b9f7df2a`。
版本：`catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6`。
统一交付：`artifacts/catalog/catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6/20261102T021202Z-b9f7df2a/`。
完成索引 `delivery.json` SHA-256：`e51524e08a2cb82aeeddb4a3715d181aaf9be52bff17cae4f3aad4363118766f`。
当前 53 个交付文件（含完成索引），运行 JSON 与原内置/基线/本地输出同字节，
完整文件 SHA-256 保持
`04107af45ba03e3a4e7660aa5255997303b8c1442b74fc6d905b37fba0ba80ca`。

实际执行：编译工作 CLI、核对完整复制后的原 store、review 导出 dry-run/confirm、
公共 `cataloggen --builtin --out <版本化快照> --check`，5 项实际退出均为 0。
原命令/日志在交付 `verification/commands/` 及工作批次 `commands/`，
归集脚本本身已按原字节存入 `provenance/assembly.py`。
脚本整体实际 exit0，原件为 `artifacts/catalog/work/20261102T021202Z-b9f7df2a/assembly-command.json`。
首次脚本写入因 Python 引号错误未启动，未创建交付目录；诊断保存在同工作目录
`setup-failure-v1.json`，不冒充成功。

代码判定基准 `37a283ee4ef743f93d3a12132d1bc2c3211fb1e5`，目标
`4831e3544945f667d9a817be8dfb92dd08856575`，核心运行代码与依赖均 NO；
其他 Go 变更明确列为 ChineseEvaluation 示例默认配置路径及测试。
外部 JSON 仅对已实际验证的外部加载程序给 JSON-only 结论；内置默认需要重建。
旧版本兼容与厂商资格不扩张。工作 store 复制到 `artifacts/catalog/work/store/`，
原 `.catalog/` 及全部旧证据保持。

## 独立审查修订

首轮 Standards 通过；Spec 发现两项：SOP 用 mkdir -p 可能复用交付目录，及
旧 changelog/随包说明对有加载能力的程序承诺范围过宽。改为工作/交付目录独占
创建、失败即停止；JSON-only 明确只对实际受测程序，不推断其他旧版本兼容。
首批目录保留且未重写，在统一入口标为已被新批次替代。

当前交付批次：`20261102T022319Z-67bee386`。
路径：`artifacts/catalog/catalog-94a5402bdf07d2d75497c3531671126f07b887c274f2cf3cb7e98e238f1e50d6/20261102T022319Z-67bee386/`。
`delivery.json` SHA-256：`21330a2138f28be8e85a6be57b9d2537bc3696ea406ac5d796ddf82286b4fd5d`。
完整 54 文件库存（含索引），增加原 changelog 原字节快照，避免后续追加使原文不可取。
代码报告同时记录未提交内容不影响运行代码判定的完整来源/输入相等性依据。
五项公开命令再次全部 exit0，脚本原件及实际整体 exit0 记录独立留存本批次 work。
目录版本与运行 JSON 原字节保持，完整文件 SHA-256 仍为 `04107af45ba03e3a4e7660aa5255997303b8c1442b74fc6d905b37fba0ba80ca`。
