# 普通测试固定输入交付

Status: claimed
Type: maintenance
Updated: 2026-10-24

当前仅实施[工单01](issues/01-retained-ordinary-inputs.md)，依据审查 F01、2026-10-24 执行方案第1项及用户本次明确授权。唯一动态状态见[status.md](status.md)。

目标：Git 内容足以在干净工作区运行受影响普通测试。稳定输入进入 ai/e2e/testdata/retained；历史字节、固定锚点、缺输入失败和语义拒绝保持。runtime 只收实际消费的完整场景，派生清单与原件身份区分。

公共 seam：现有 Go E2E → verifier CLI / release 消费者；先在干净独立 Git checkout 复现缺输入 RED，再最小修改取得 GREEN。无自动下载、skip、重建历史证据或旧路径兼容。

本次不修改产品协议/目录/支持矩阵，不迁移活动工具，不恢复地区暂停轮次，不执行付费live/race/压力/真实清理。其他治理单元仍为规划。
