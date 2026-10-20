# 11: 取得六个候选型号组合的真实 API 与价格证据

Status: resolved
Type: task

由 2026-10-17 implement-spec 实施指令发布。

**What to build:** 维护者通过公共 Client，在三种 GLM 官方协议上分别验证 GLM-5.3 和 GLM-5.2 的必交能力，并获得可追溯的容量、用量和美元价格候选快照。

Blocked by: 09

依据：GLM 接入 spec R03、R06、R07、R13、R14、R17、R21；验收组 A02、A09、A15；User Stories 13、14、38、39、54、55、56。

- [x] 先登记每协议两型号的必交能力、未知变量、预期、预算和证据格式，再扩展现有候选 live 验收；正式支持仍由公共 Client 证明。
- [x] 核验中国区标准 Key 和三条官方端点；每进程仅本组合 Key，显式 SDK 配置和双开关生效，账户/区域/Key 类别用别名记录，不读取套餐或其他环境凭据。
- [x] 为六个候选身份固定有日期的来源和容量上限；1M/128K 不冒称未经核验的精确整数，需要保守值时明确其 barness 支持子集性质。
- [x] 从官方人民币报价取得两型号的输入/输出/缓存/阶梯费率，并以有来源、日期、固定换算率转成美元候选 Price Snapshot；无来源的零价格不能作为免费。
- [x] 逐协议完成两型号的文本、工具往返、合法历史、可用思考、取消/用量等必交场景，专项收敛 Responses effort、Chat temperature=0/include_usage、Messages thinking/签名与默认头的未知事实。
- [x] 遵守每组合每次运行最多 24 Logical Call/32 HTTP Attempt、每次最多 4096 输出 token、进程 1800 秒/调用不超 120 秒及更短 Host 时限；Binding/SDK 重试关闭，场景级故障重试至多一次且计入预算。
- [x] 先预留预算，一次只改一个待确认变量；缺 Key 为 NOT_RUN、真实失败为 FAIL，已发生失败保留，不隐式加预算或主动制造欠费/套餐故障。
- [x] 返回六个候选的实际结果、用量/估算、SDK、request id 来源、截断标记及脱敏审计；候选报告不更新正式支持矩阵。
- [x] 任何必交型号/协议失败都阻止候选整体验收，不删目标或把必交能力改成 UNSUPPORTED；可选可观察思考的限制只按 spec 的证据条件登记。

**独立验收：** 获得三协议、两型号各自完整的候选报告和有来源的容量/价格快照；若缺凭据或能力失败，结果如实保留而不宣称交付。

**依赖理由：** 付费候选验收必须建立在三路线完整行为、故障和隔离/资源检查之上；可与 10 的固定基线核验独立推进。


## Answer

2026-10-17：六个候选身份全部完成自身真实验收，三协议各两型号11/11场景 PASS。每协议22 Logical Call /22 Attempt预留/22实际HTTP，零重跑；总计66/66/66。Chat131.53s、Responses111.79s、Messages108.99s，均满足24/32、4096输出、120s调用、1800s进程及更短Host时限，SDK/Binding重试0。直接官方中国端点、单路单Key与双开关隔离；Host只提取已授权root.env GLM_KEY，账户/区域/权限/账单不冒充核实。

收费前固定并提交计划/完整官方来源/六条独立USD快照，实际源冻结06975520f8f1a76278644dd26d7926bfef24a1d2、binary87cf6b3ed542e7132ca6b399a74305ea0f0a3928a4e08ad88a09addac75f79d7，sourceManifest837a83c4711dc4914b8110341403c0d6e43cac254b3101f601d6d538b209c907。旧synthetic费率未重标。CFETS2026-10-16固定6.7220换算，缓存写计算按普通输入的推断与保守容量边界均独立注明。

六条路线均有自己的文本四入口、Host工具校验/执行/结果继续、可信原生历史、取消与恢复、协议专属单变量对照。Responses5.3自身InputReasoning对象回放已实证；Messages5.2 native/unsigned保留enabled1024/max4096，5.3默认思考及各自signed/unsigned回放通过；canonical Messages maps全null。全部无截断，343文件/协议inventory及脱敏审计通过，实际request-ID来源与原始用量缺项保留。报告已知估算合计USD0.08818744，不是完整消耗或账单；Chat/Responses各20complete+2unreported，Messages22partial。

原始真实ZIP封存后才合并10的账本/文档/测试，并补全无需真实Key/网络的公共Client精确捕获回放。三协议各22 replay attempts、0外部HTTP，原始wire/Result/Observer按完整上游ID、工具配对、原生签名逐项比较，仅Host本地ID/时钟/耗时的明确路径归一化。Mode=replay和独立报告不会伪装live。缺失/改写/错路归档、凭据、窄化入口及provider-ID/signature突变均按公共子进程反例验证。normal11.36s、race190.50s及live vet通过；所有历史RED与raw child包保留。

完整来源、六型号报告、Host receipts、raw ZIP SHA256、可重复离线命令与12接续说明见[证据索引](../evidence/ticket-11/README.md)和[handoff-12](../evidence/ticket-11/handoff-12.md)。候选/回放未发布BuiltinCatalog、支持矩阵或正式route；12负责当前正式Hash自身重跑，13负责发布门禁。
