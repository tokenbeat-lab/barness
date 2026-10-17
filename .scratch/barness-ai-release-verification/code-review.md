# 两轴复核：当前发布续修

固定比较点 `b61c6a5`；源码改动为 `4e4d99d`、`f7c3aff`、`cc78cfd`。
诊断归档 `f29703b`；完整门禁执行期间不改 Go、认证、负载、预算或比较。
两个独立只读 reviewer 按 code-review skill 分别审查 Standards 与 Spec；
标准为根 AGENTS.md、原票 14、续修票 01/02 和压力原 #07。

## Standards

源码与已归档自然反例：无剩余发现或需要记录的 smell judgement call。
追加发现一项 P2：诊断脚本在 try/finally 前启动 ambient，随后哈希/receipt 失败可能
遗漏回收；主实验超时不记录终态。已把启动与后续操作纳入清理范围，finally 回收、
关闭日志，并记录异常/超时/中断终态及完整 600 Go 文件不变性。独立只读复核确认关闭。
未在完整门禁期间重新运行竞争缓冲诊断；此前最终脚本正常路径的实际执行已保留。
封存器又发现一项 P2：必需 native 根缺失会被空集合误收；已在 record/verify 双向
拒绝缺失、非目录与链接根，并固定父目录精确条目集。独立只读复核确认关闭。
最终CLI原包已独立核对：normal/race各9,098PASS、0FAIL，六受控gate通过，17输入audit0，
600个Go源Hash一致；无新增Standards发现。封存完成后独立执行--verify。

## Spec

源码无剩余发现。两次顺序、独立的自然 RED 均定位到解析前相同首差，与连接内原生
ENOBUFS/0-reported 和同 Hash 不安全续写关联；不注入错误或正文，实际 FAIL 原样保留。
同条件安全对照32次调用各一次 Attempt，22显式失败、10成功、零首差、资源0、audit0。
公开 regression RED/GREEN 与 single-P 原复现已保留。因此原 #07 除完整本版门禁外的
因果、回归要求已满足；未声称反向认证历史 call-04/call-11 的原始字节。

原 #07 只能在完整 normal/race 及本缺陷相关质量、完整性和审计验收通过后关闭。CLI 在启动时固定的未关闭 blocker 快照必须保留，
允许在最终回执记录验收后的前/后原文、Hash、时间与对应证据；不能追认旧 report PASS。
真实支持和效果仍 NOT_RUN，整体发布保持 BLOCKED。

最终Spec复核确认完整normal/race及相关质量/完整性/审计条件已满足，可以关闭原#07
及续修01/02。原发布报告SHA-256固定为
`085add91f8ec6c2cefe54c74770298164a7cbd44e86b64252c9d1f9e7f5ee313`；
当前原工单仅在该报告结束后转换，前后snapshot/Hash/时间均保留，NOT_RUN/整体阻断不变。

最终剩余发现：Standards 0；Spec 0。两项归档脚本P2均已修复并独立确认关闭。
