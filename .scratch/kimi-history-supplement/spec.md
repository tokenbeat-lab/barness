# Kimi 历史缺项最小补采与组合消费实施

Status: claimed
Type: implementation

用户已授权按 ../kimi-history-supplement-plan/plan.md 实施：先离线审阅六条 Kimi 来源差异并重建/回放54项旧捕获，再实施固定单场景入口、schema2 单项范围与只读两来源组合消费者，完成受控及完整普通验证。禁止真实 paid grant、凭据读取、厂商 live、race、压力、目录激活、暂停恢复；旧事实/授权/账本/失败全部保留。

先失败再实现：旧单包9项/新1项都不应独自获得十项资格；准确9+1受控组合验证机制但不产生真实资格；错误模式/身份/授权/预算/来源/原件/传递库存/真实退出必须拒绝。每份新单项绑定固定场景和原 baseline manifest，exact entry/240s、calls1/HTTP1/retry0/input4096/output4096/USD0.10；旧 full/schema1 不兼容启动新单项，反向也拒绝。全部合成授权及claim写入私有 Git common namespace，写前断言，前后保护公共历史存储。

公共入口：既有 processgrant、精确 TestKimiHistorySupplement、releasegate -inspect-bundle 与 consume_live.py/record。不增加通用场景图、续期/恢复控制器或并发状态服务，不修改产品 adapter/SDK。受控公共进程先RED/GREEN，完整受影响普通包、工具过程与vet收尾；race/压力候选按共享claim输入影响记录为NOT_RUN。
