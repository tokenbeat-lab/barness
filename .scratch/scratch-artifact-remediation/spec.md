# 测试输入与规范入口治理

Status: resolved
Type: maintenance
Updated: 2026-10-24

当前实施[工单02](issues/02-current-contract-and-entrypoints.md)，依据用户2026-10-24对执行方案第2项的明确授权。[工单01](issues/01-retained-ordinary-inputs.md)保持已完成；唯一动态状态见[status.md](status.md)。

目标：Git 内容足以在干净工作区运行受影响普通测试。稳定输入进入 ai/e2e/testdata/retained；历史字节、固定锚点、缺输入失败和语义拒绝保持。runtime 只收实际消费的完整场景，派生清单与原件身份区分。

公共 seam：现有 Go E2E → verifier CLI / release 消费者；先在干净独立 Git checkout 复现缺输入 RED，再最小修改取得 GREEN。无自动下载、skip、重建历史证据或旧路径兼容。

本次不修改产品协议/目录/支持矩阵，不迁移活动工具，不恢复地区暂停轮次，不执行付费live/race/压力/真实清理。其他治理单元仍为规划。

## 第2项范围

统一基础spec、合同、追溯说明和地区工具README；保留旧文档原字节、accepted ADR、历史原件与recovery7暂停。按纯文档检查引用/格式/状态/来源，三份旧选定Markdown变化单独记录，不重写上一轮回执或声称新Go验证。其他治理项仍为规划。
