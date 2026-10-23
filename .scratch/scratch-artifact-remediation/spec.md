# 测试输入与规范入口治理

Status: claimed
Type: maintenance
Updated: 2026-10-24

本轮实施[工单05](issues/05-active-tool-migration.md)，依据用户2026-10-24对第5A项的明确授权；工单04保持完成。[工单01](issues/01-retained-ordinary-inputs.md)、[工单02](issues/02-current-contract-and-entrypoints.md)和[工单03](issues/03-script-boundaries.md)保持已完成；唯一动态状态见[status.md](status.md)。

目标：Git 内容足以在干净工作区运行受影响普通测试。稳定输入进入 ai/e2e/testdata/retained；历史字节、固定锚点、缺输入失败和语义拒绝保持。runtime 只收实际消费的完整场景，派生清单与原件身份区分。

公共 seam：现有 Go E2E → verifier CLI / release 消费者；先在干净独立 Git checkout 复现缺输入 RED，再最小修改取得 GREEN。无自动下载、skip、重建历史证据或旧路径兼容。

本次不修改产品协议/目录/支持矩阵，仅迁移已确认的SOP工具/测试和Jev核验器，不恢复地区暂停轮次，不执行付费live/race/压力/真实清理。第5B项尚未实施。

## 第2项范围

统一基础spec、合同、追溯说明和地区工具README；保留旧文档原字节、accepted ADR、历史原件与recovery7暂停。按纯文档检查引用/格式/状态/来源，三份旧选定Markdown变化单独记录，不重写上一轮回执或声称新Go验证。其他治理项仍为规划。

## 第3项范围

以静态AST、源码引用及原生消费者接线区分318份独立脚本和2份压缩快照的身份。逐文件记录输入/输出/副作用、当前调用方、源码与历史pin适用范围、替代入口；合成测试对旧控制器的消费与真实轮次的关闭/暂停分别标注。保留原脚本和被封存文档字节，只在可编辑活动入口添加现行分类导航。不得运行被审查脚本来证明其身份。

交付入口：[脚本身份与执行边界](script-boundaries.md)、[完整逐文件清单](script-boundaries.json)及其中九份目录分册。当前工具README增加导航，原字节及基址保存在[本轮历史索引](history/phase3-before-20261024/index.json)。

## 第4项范围

保持349处原失效候选及原文不动，通过外部索引分类解析；原样快照固定原路径/相对基址、版本/Git/内容哈希及当前入口。按明确文件、Git版本和必要的有界归档定位查找，区分原哈希确认找回、仅历史版本可读、来源基址误解和仍未确认。原件缺失或原版本无法确认时，受影响结论与资格不放宽。

交付入口：[历史追溯总览](historical-traceability.md)、[349项处置](historical-links.json)、[快照来源](historical-snapshots.json)。原正文与临时路径保持；找到的永久副本和仍未确认的原版本分别记录。

## 第5A项范围

归属见[ADR-0053](../../docs/adr/0053-verification-tool-ownership.md)；先在公共CLI/Go消费者建立迁移失败，再迁移和验证受影响入口。历史原件/资格pin保留，活动旧路径删除，无兼容wrapper。独立来源选择器完整覆盖新路径和PY。recovery7暂停，本次禁止付费、race、压力与真实清理。
