# Provider 审计历史归档

本目录保存 2026-11-02 的历史审计。原审计源码提交为
`37a283ee4ef743f93d3a12132d1bc2c3211fb1e5`，目录为 118 条身份：
112 条聊天、3 条图片、3 条分类。报告里的“当前”均指该历史上下文。
此处不表示最新目录、当前源码资格或完整产品发布状态。

## 归档内容

| 原件 | 用途 |
| --- | --- |
| [report.md](report.md) | 当时的 Provider 接入、证据引用及交付范围核对 |
| [inventory.json](inventory.json) | 当时的目录统计、原件引用与核对范围 |
| [catalog.json](catalog.json) | 当时通过公共 BuiltinCatalog 读取的完整目录 |
| [audit.py](audit.py) | 当时执行的审计脚本源码 |

四个原件按原路径、原字节提交。[archive.json](archive.json) 固定原件 SHA-256、
历史来源和目录身份。本次只增加归档说明、摘要与导航，不重跑历史审计，也不改写
历史测量、失败、授权、支持矩阵或唯一当前状态。

## 阅读与复核边界

报告与清单引用的完整证据仍位于原 `.evidence/` 路径；本目录不是自包含证据包。
原相对链接保持，后续可变状态文件须结合历史提交阅读，不能作为冻结历史字节。
最新目录查看 [Catalog 交付入口](../../artifacts/catalog/README.md)。
后续变化查看 [Terra 状态](../catalog-gpt61-terra/status.md)、
[Kimi K3.1 Pro 状态](../catalog-kimi-k31-pro/status.md) 与
[型号准入策略状态](../catalog-model-id-policy/status.md)。

`audit.py` 会写回同目录的 catalog.json 和 inventory.json。
它只保留为历史执行源，不是当前审计入口。需要重放时，必须在独立副本中使用原
源码与留存依赖，并把新输出写入新批次；不得覆盖本归档或把历史 PASS 当作新执行。
