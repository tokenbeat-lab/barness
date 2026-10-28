# 当前状态

2026-10-28：前置脚本账本副本类型错误已修复。本次仅离线修复，不恢复付费链；此前两次失败、STOP、已完成的离线54项审阅及六份不可执行机器范围均保留。

原因：preflight.py始终把preservedCopy当路径字符串；国际Responses/Messages的旧记录实际为{path,sha256}引用对象。上次仅offline_review.py处理了该格式，实际入口没有接入。本次在common.py增加一处明确的账本记录读取校验，preflight.py与offline_review.py均调用；字符串格式从外层取哈希，对象格式从引用内取路径和哈希，随后独立核验副本和权威账本。非法格式/缺字段明确报错，哈希不符仍拒绝，不改写历史格式。

先保留实际入口读取段的失败进程回归（退出1），再修改代码；34项离线子进程检查全部通过（退出0），包括两个入口分别读取六路线共12条真实账本记录、两种格式，以及错误类型/缺字段/副本或权威账本篡改拒绝。仅执行实际入口的账本读取段，未重跑会写STOP/scope的完整前置流程，也未重跑厂商采集。

2051来源与冻结仍一致，187普通输入保持；46个既有执行证据文件及8165个授权/账本文件集合和字节不变。付费live、race、压力、正式消费、目录激活和暂停恢复均未执行，新grant/claim/凭据读取/厂商请求均0。并发/生命周期/负载未变，race/压力本次不适用。

修复与证据：[verification.json](../../.evidence/kimi-history-paid-restart-20261028/type-error-fix-v1/verification.json)。先前原采集/停止事实见[completion-v1.json](../../.evidence/kimi-history-paid-restart-20261028/completion-v1.json)，本次不改其测量或状态。回滚仅恢复本次三个本地脚本的before快照；历史账本和证据无需回滚。
