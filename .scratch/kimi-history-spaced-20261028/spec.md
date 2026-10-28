# Kimi 剩余四路：加大间隔后的独立付费批准

本次直接人类指令为“拉大调用间隔，继续验证流程”，承接开头的全部身份、预算、固定构建、原件保留、串行和失败即停边界。原两路Chat候选9＋1只读沿用，不重采。

仅依序新增kimi-cn-responses、kimi-intl-responses、kimi-cn-messages、kimi-intl-messages的kimi-k3/xai-origin-history；live/candidate、自有standard按量API。每路独立新schema2 grant、1调用/1HTTP/USD0.10；本次最多4/4/USD0.40。旧失败的授权和消费不恢复，不借用旧预算。每路采集及原九项＋新一项候选组合均通过才开始下一路；任何失败停止本次全部余下付费，不自动重试或再签ID。

冻结生产二进制、source、工具及SHA仍与开头指定freeze.json完全相同；进程内callIntervalMs=60000、4096输入/输出、call120s、process240s及精确唯一入口全部不变。额外在进程外冷却300秒：首路从本次批准记录时刻起计算，后续从上一原生进程结束时刻起计算。冷却期间不读凭据、不发grant；冷却结束后当前路线才签发30分钟有效的新grant，再启动冻结生产者，仍经历其60秒首等。仅当前站点凭据注入子进程，所有自动重试/重跑为0，无探活/预热/第二次请求。

本次在新目录保存批准、原件引用、冷却时间、四路新机器范围和未来grant；先重新离线核验原54项和当前来源/普通输入，再用冻结消费者只读检查两路Chat原组合。旧STOP、失败、账本和未知费用保留；缺失/partial/明确零值按Kimi自身合同，不补零、不采用MiniMax例外。

仅形成候选9＋1适用性记录；目录active=false。不运行正式资格消费、晋升激活、race、压力、其他Provider/xAI请求或暂停轮次恢复。每组独立保留两份测量、各自来源、时间、预算、退出、授权和账本，不要求六路/新旧同build。当前状态status.md，独立索引位于.evidence/kimi-history-spaced-20261028。
