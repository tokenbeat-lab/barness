# Kimi 单场景补采方案当前状态

设计已完成，状态为 **DESIGN_READY / NOT_IMPLEMENTED / NOT_AUTHORIZED**。本轮没有采集新厂商事实、生成可执行授权、claim、目录晋升或完整当前资格；Kimi 仍为六条各九项历史事实、缺一项的候选状态。

[方案](plan.md) 推荐在付费前完成 Kimi 自身源码差异审阅及旧 54 项的当前公共 Client 回放，再用一个固定单场景入口与一个扁平的两原件组合消费接口补齐。原九项真实性/原件完整与当前产品适用性分开判断；不得因已有原九项 PASS 就跳过后者。

| 项目 | 已完成/提议范围 |
| --- | --- |
| 离线原件复核 | 六原生 bundle 的 manifest/artifact/seal、54 项 PASS、原授权/命令/原生退出/账本引用及当前候选型号投影均有效 |
| 当前来源适用性 | 本轮未完成 Kimi 差异放行；前次保守差异 Chat/Messages 各11、Responses各19。实施前逐项审阅/54项回放，未知或真实受影响场景单列，不自动 paid |
| 单项补采（拟新增） | 每路线仅 kimi-k3/xai-origin-history；固定合成 xAI Host 历史，一次 Kimi 请求，不访问 xAI 或执行工具，无旧九项前置请求 |
| 新授权提案 | 六份独立新 schema2/purpose/contract，绑定各自原 baseline manifest 与冻结新来源/二进制/SDK/目录/预算/账户；本轮未生成 ID/expiry/grant/claim |
| 单路线预算提案 | calls1 / HTTP1 / retry0 / input reservation≤4096 / requested output4096 / USD≤0.10 / call120s / 首等60s / 进程240s |
| 六路线预算提案 | calls≤6 / HTTP≤6 / output≤24576 / input reservation≤24576 / USD≤0.60；按冻结价格 token 上限预留合计约0.5139718321938432美元。不是实时报价或已有付费批准 |
| 组合（拟新增） | 每路线旧9+新1；各自来源、时间、候选标签、授权/账本和费用分开保留，10项映射回原manifest/case；不伪造单进程十项报告、不递归、不复制胶囊 |
| 正式激活/资格 | 当前 catalog active=false；完整候选证据、正式数据审阅/激活、正式只读消费、整体发布各自判断，不能通过改旧stage取得支持 |

证据与预算见 `.evidence/kimi-history-supplement-plan/route-evidence-and-budget-v2.json`，原核验日志为 `inspection-command-v1.log`。其中 4,096 输入限额依据现有受控 StreamSimple/64 夹具的 2,594–2,953 输入预留；新入口 full/4096 的最终编码必须在发放付费 grant 前用无 Key E2E 确认，不够则离线修订方案，不能现场抬预算。

现有入口只支持整组十项、单包完整资格，不能直接执行本方案。本轮仅新增本设计目录和只读核验材料，未修改任何运行时/采集/消费者/授权实现、ADR/SOP 或 Kimi 激活数据。普通/受控实现测试未执行；本轮纯设计校验引用、预算算术、原件与来源集合。后续实施先失败后实现，再按 SOP 验证对应普通范围；新增 grant/claim 的 race 范围待另行人工选择，本轮及实现授权均不自动执行 race/压力。

本轮实际付费 live=0、凭据读取=0、新 grant=0、真实 claim=0、race/压力=NOT_RUN、暂停恢复=0。所有历史失败和账本保持，保留核验见 history-protection-v1.json；本状态不改写此前 current-qualification-index 的结论。
