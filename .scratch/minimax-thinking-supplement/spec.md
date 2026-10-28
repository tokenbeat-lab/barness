# MiniMax thinking-on 单项与固定十三加一

Status: implementing

2026-10-28 用户授权实施；[批准设计](../minimax-thinking-supplement-plan/plan.md)为完整实现合同。职责、失败清单、预算及不扩范围约束全部沿用，特别是第6节先失败验证清单。

固定身份 minimax-messages/minimax/anthropic-messages/MiniMax-M3/standard。只增加单项采集与专用只读组合，不改公共Client、SDK、目录和价格。不签发付费授权、不请求厂商、不启动 race/压力、不恢复暂停。合成 grant/claim 仅私有 Git common namespace，写前断言。

公共入口：processgrant 固定 profile 模板、精确 TestMiniMaxThinkingOnSupplement、releasegate inspect/consume/record。失败清单：范围/身份/来源/基包/binary/入口/count/时限/expiry/USD错误；重复 reserve/Attempt；重试；断流/缺终态/未知写价；缺/重复/额外/篡改原件；混合mode/账户/SDK/跨包工具历史；旧失败冒充成功；单包冒充完整；来源与selectedInputs缺失/冲突；旧canonical/claim改变。全部在实现前建立公共进程测试，保留 RED/GREEN。

普通范围沿用前次七包并增加 livegrant，相关 live-tag 子进程均无厂商调用；race/压力的claim及取消候选仅列 NOT_RUN。唯一当前状态：[status.md](status.md)。

原件核验发现：旧 FAIL producer 的 bundles=[]，stderr 有唯一精确原生路径。仅 baseline FAIL/exit1 允许固定 parent/bundles/session（受控为parent/session）布局与原日志标记、claim及封存包共同归属；成功单项不允许该分支。新公共进程先重现该真实布局再实现。MiniMax未实现Kimi专有CostEstimate，单项以原生四个计费counter完整性、现有独立usage/cost重算及最高费率预留判断；不为此修改产品API。
