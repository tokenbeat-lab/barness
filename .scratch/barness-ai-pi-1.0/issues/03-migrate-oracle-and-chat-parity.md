# 03: 切换唯一 oracle，完成五项聊天对齐

**What to build:** 让现有聊天调用以 pi-ai 1.0.0 为唯一冻结基线，并交付该版本的五项聊天修正；维护者能在同一次绿色合入中看到基线来源、全部行为变化、目录更新与零待处理差分。

**Blocked by:** 01 — 拆分目录职责，保持现有聊天行为；02 — 核验 pi-ai 1.0.0 发布件。

**Status:** resolved

**依据：** 规格“基线迁移”“随基线迁移的五项聊天行为”、实施设计第 4–5 节，以及 ADR-0004、ADR-0006、ADR-0010、ADR-0011、ADR-0018。

**已确认的规划决定（2026-10-07）：** 新基线要求差分零待处理，而五项聊天修正会直接影响该差分。发布件等价核验先由 02 独立完成；启用新 oracle 与五项修正在本工单一起合入，不产生无法通过门禁的中间基线提交。

- [x] 先更新或新增五组公共入口失败场景 fixture，再修改行为。每个受影响用例写明原行为、1.0 行为与冻结 pi 依据。
- [x] 唯一 oracle 固定为 1.0.0，使用 02 的等价证明更新 provenance 和锁定依赖；runner 拒绝版本或模型数据哈希不符的副本，正确过滤聊天型号并适配 1.0 的入口与目录导出。
- [x] 无法解析或非有限的 Retry-After 值回退到带抖动的指数退避：基数 0.5 秒、上限 8 秒、抖动 0–25%；有效的后续响应头仍按 pi 顺序处理，有效服务端等待受绑定上限约束，全部等待受调用截止时间约束。假时钟证明 Infinity、无法解析值和取消行为。
- [x] Responses 按响应报告的服务等级、缺失时按调用选项估价：flex 为 0.5 倍，priority 与 fast 为 2 倍，gpt-5.5 的 priority 与 fast 为 2.5 倍；逐位符合 pi 的计算顺序。
- [x] Anthropic 后续增量中的 1h 缓存写入明细覆盖原值，不累加；缺明细时保留已有值，用量完整性规则不变。
- [x] Responses 与 Chat Completions 在 full 和 simple 入口只合并一次 samplingParams，顺序为命名字段、型号默认、调用级；调用级覆盖同名默认值，型号默认与调用级均不能覆盖保留字段。
- [x] Responses 在终态事件检查之后核对每个工具块的完成事件；缺失、null、数字输出索引独立匹配，重复索引覆盖的旧块也受检查，未完成工具调用产生 protocol 错误，正常完成对照组通过。
- [x] 内建聊天目录按 1.0 模型数据逐字段核对，依 ADR-0018 评估型号硬约束和已批准扩展；内容变化升目录版本并重建价格与目录快照。
- [x] 全量新基线差分零 pending；五项以外的新发现经来源核查和维护者决定处理，不通过扩大已有扩展登记隐藏差异。
- [x] 删除旧 0.87.1 oracle 副本与运行路径，修订相关 ADR、基线说明和差异登记；交付 vet、race、离线 E2E、全量差分、快照核对及脱敏审计证据。

## Comments

2026-10-08：五项行为、唯一 oracle 与目录快照已实现。代码评审发现缺失输出索引和显式 0 合并会漏报未完成工具调用；已先补失败 E2E，再保留 missing/null/number 的独立键，并以 3×3 矩阵复核 full/simple、stream/complete 和冻结 pi 差分。Standards 无 actionable finding，Spec finding 修复复核通过。

全量新基线另外发现 OpenAI Node SDK 7.19.0 带来的两类错误文本变化：HTTP 对象体缺少/null error 的包装，以及具名 SSE event:error 的处理，共四个用例、八个 pending。按本工单“新发现经来源核查和维护者决定处理”的要求，等待维护者选择同步文本或逐用例批准扩展，当前未扩大 extension 登记。来源、具体前后文本与复现命令见 [额外差异](../chat-parity-evidence/extra-differences.md)；最终验证记录见 [证据](../chat-parity-evidence/README.md)。保持同一次绿色合入，决定前不提交失败门禁的中间基线。


2026-10-08 维护者决定：用户回复“按建议同步”，批准将两类额外错误文本同步到 pi 1.0.0。先更新四个受影响 fixture，并新增对象／数组／标量 HTTP 与具名／未具名 SSE 对照，再实现 SDK 错误归一化；只增加 fixed 登记，原 extension 保持不变。Responses 使用 SDK 原始 SSE decoder 保留 event 名，再解码 typed event；初始 typed stream 继续负责响应体关闭。复核发现 Chat 对非对象具名错误先做 DTO 解码会漏报，已先补数组、字符串、数字、false、null 失败场景，再将错误判断移至 chunk 解码前。单用例 replay 还暴露 fixture 在合成 key 注册前写入造成审计失败，已将已知 fixture key 的注册移至 TestMain，使证据不依赖测试顺序。


2026-10-08 完成：最终完整运行 3345 E2E PASS，611 差分 PASS／0 pending，7 policy pressure PASS，race、vet、快照逐字节比较与脱敏审计 PASS；Standards／Spec 复核无剩余 finding。一次中间压力运行的单个 loopback 流中断未在最终全量及单独重放复现，失败与重放证据均保留，未放宽限制或重试。全部验收项完成；按规划以唯一绿色提交交付。
