固定审查：`56301103ae1e35a9e0354e8c1b5574a7aa32fda1...9ea3b9d3d72c37c538971c7fafa40ade5c883fa7`，工作树HEAD与pin一致。Spec发现2项，最高P2。

- **[P2] 空摘要part错误取得可见推理通道优先级。** R11要求两种reasoning文本事件正确合并；设计第237行明确“首个有效通道驱动增量”。`ai/internal/testkit/supportmatrix/xai_wire_parts.go:65–69` 在忽略空`reasoning_summary_part.added`之前已设置primary。合法顺序：空summary part added→reasoning_text.delta为`Detail.`→终态同时含summary=`Summary.`、content=`Detail.`。公共运行时在`ai/responses_xai_text.go:66–67`提前返回，因此实际Thinking为`Detail.`；新消费者却选择summary，继而在`xai_output.go:33–34`拒绝这份合法资格原件。应让空part保持无通道语义，并以公共Client/CLI正例固定此行为。

- **[P2] 新旧路线历史证明未核对函数内容。** R12要求“工具调用与结果同步重命名”，R13要求降级“保留可见文本与工具配对”；工单14要求实际目标wire证明碰撞/复用正确。固定Host历史在`ai/internal/testkit/grokhistory/history.go:66–75`包含同名函数与`first/second/first`参数标签；但`grokhistory/wire.go:32–41、53–58、66–73、89–96`仅提取调用ID，随后第102–129行只比较ID序列与结果标记。将前两个历史调用的label互换，保留ID、结果文本和不可变源历史，生产检查及复用该helper的原生消费者仍会通过；这会掩盖降级时参数错配。需按协议独立核对三个调用的函数名、完整参数与相应结果，而非只核对关联字符串。

v5的5项Spec发现均已修复：完整流块/增量、原始输出到有序Result与可见推理、第二轮none及精确结果文本、默认收费语义、可空密文及专门续接证明。未发现额外范围扩张。新增授权仍绑定本次scope/source/binary/目录与持久化claim；旧完整能力为追加，历史Plan13/closed ledger仍保留为历史事实。

本次只读静态审查，未运行测试、构建或厂商调用。尚未授权/执行的formal live、race、压力不计为软件缺陷；最终完整普通验证与发布判断仍由集成阶段完成。
