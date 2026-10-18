# MiniMax 交付状态

Delivery: PARTIAL
Updated: 2026-10-18
Integration branch: `codex/minimax-global`
Historical CN checkpoint: `codex/minimax-cn` / `4295ee8`

国际站 MiniMax 接入、M3 三协议自身候选和自身正式验收已完成。完整 spec 仍未通过：旧能力续期存在真实失败，17 身份矩阵未完成，原始完整发布门禁没有产生发布报告。spec、10、09 保持开放；所有已批准的一次性实际进程均已结束。[机器可读状态](evidence/ticket-09/international-delivery-status.json)记录当前结论与来源边界。

| 范围 | 结果 | 可核验依据 |
| --- | --- | --- |
| 三协议、M3/M3.1 六身份离线约束 | PASS，限定离线来源 | 来源8d56410，国际站完整 MiniMax 普通/race 各1340 PASS，定向各1372 PASS，双 vet、审计与两轴审查通过；见[收敛证据](evidence/ticket-10/international-convergence/README.md)。 |
| M3 三份自身完整候选 | 三项 PASS | 来源7605332，各13 calls/12 Attempts；[完整候选胶囊](evidence/ticket-10/international-complete-candidates/README.md)。 |
| M3 三份自身正式与目录 | 三项 PASS，目录已激活 | 来源8900034，合计39 calls/36 Attempts；目录2026-10-18.1-minimax / `3718603f`；[正式胶囊](evidence/ticket-10/international-formal-glm/README.md)。三份263文件执行映射仍适用，公共验证通过。 |
| M3.1 三协议线上候选 | 三项 NOT_RUN，0/0 | 缺少获准 M Plan 资格；没有读取 Key、建立付费 intent 或虚构公开价格。离线约束已完成。 |
| GLM 唯一获批补正 R1 | Chat/Messages PASS，Responses FAIL | R1 三路各22/22消费，Responses为21 PASS/1 FAIL；完整上游 SSE 缺正文。连原轮累计132 calls/132 Attempts；[实际补正胶囊](evidence/ticket-12/actual-corrective-1/README.md)。217文件执行映射仍适用，无 R2。 |
| 其余15项原授权身份/效果 | 2 PASS / 13 FAIL，168/168 | 两项 DeepSeek PASS；其他失败及全部消费保留，见[实际证据](evidence/ticket-10/international-legacy-final/README.md)。来源3d3a1fb为历史来源，不给后续 Go 修改刷新证明。 |
| 原始完整 ordinary/race 发布门禁 | 两套各12385 PASS / 6 FAIL，审计0；无发布报告 | 来源3d3a1fb，实际运行双 vet、两套完整测试及 pi/压力；随后读取 bundle 失败。最终 verdict 为 null，见[原始完整证据](evidence/ticket-09/international-final-actual/README.md)。 |
| 发布证据消费者修复 | 定向普通/race各93 PASS，审计0；双vet通过 | 见[修复与复核入口](evidence/ticket-09/release-consumer/README.md)。这不授予新全量认证。 |

## 实际失败与消费边界

R1 的 `glm-5.2/native-history` 返回 HTTP200、完整 SSE 与 completed 终止，但只有推理、没有正文或 refusal；适配器没有丢失内容。不能将推理改作正文，也不能拼接原轮次。严格 raw/views 续期拒绝，四个旧 selector 字节未变。捕获和场景清单齐全、`complete=false` 的 FAIL 报告可以如实导入临时矩阵，但不能因此授予发布支持。

剩余15项的失败包括连接前超时/未取得 HTTP 响应、Anthropic 实际 HTTP403、TypeSafe 实际 HTTP451“Typesafe is not available in your region.”。403 不证明 Key 无效；未取得响应的记录不足以区分 DNS/TCP/TLS。20个 Attempt 用量完整，148个未上报，未上报不能记作免费。[只读诊断与原始审计](evidence/ticket-10/international-legacy-final/audit.md)保留这些限制。

旧能力本次续期共300 calls/300 Attempts（GLM132/132加其余168/168），批准上限440/464。M3候选阶段保留全部历史49 calls/45实际Attempts（49 reserved），正式阶段39/36（39 reserved）；中国区两次401仍计入候选历史。未消费额度不授予另开轮次；所有既定实际进程已关闭。

Jev1.14 的部分失败无法派生正式支持包，17项临时矩阵导入因此拒绝；默认矩阵未写入，SHA256仍为 `cacb9dd13d75e44ae6ad22103504bfa2c2f6679ae7fba66764eb679d8889519b`。本次不能完成17身份与四业务效果的当前支持认证。

## 完整测试与后续修复的适用范围

来源3d3a1fb的原始公共 CLI 确实执行了两种 vet、完整普通和完整 race；每套12391个 case 中12385 PASS/6 FAIL，两套生产者均记录审计0。六项失败为旧 GLM 正式导入/已知用量的来源不适用。每套869个 PIDiff 标记 case、29个 Pressure 标记 case 均 PASS；这些是原生产者记录的case状态，后续消费者对两份记录的拒绝见下文。内部各命令数字退出码没有持久化，保持未知；不能从静默日志补写为0。CLI最后因 bundle 目录校验失败退出，没有 `release-report.json`，gate级最终审计与判定未到达。

来源7754c7f修正了空effort case的单尾分隔符、递归artifact inventory与安全读取边界：保留原case ID及manifest字节，拒绝canonical目录别名、软链接、目录置换、文件篡改、重复枚举与超限输入。判定文件在解析前经有界、rooted、预期Hash验证；不改producer或支持断言。失败先行证据与每轮审查均保留。最终定向普通/race各93 PASS、审计0、exit0，双vet均exit0；独立只读导入测试1 PASS，原两套各12391场景与原Hash/6FAIL不变。公共CLI随后显式采用原始普通/race及17身份/效果输入，只读诊断176.607秒后生成报告：pass=false、outer exit1、commands=null；没有执行四个完整命令或新增付费。加载、回滚、多模态和六MiniMax加载消费者均PASS，保留缺失Jev支持目录和当前17身份矩阵失败。另有两项差分记录诊断，原因与后续范围见[只读原因及后续生产者修正](evidence/ticket-09/release-consumer/README.md)：原目录差分缺少实际Go版本来源；原目录查找控制没有Compare/Record/HTTP却使用PIDIFF前缀。严格消费者的拒绝正确，旧原始记录保持不变。

后续来源ad20469仅补齐目录差分的实际Go版本来源，将lookup-only控制更名为`MM06-frozen-catalog-metadata`，并更新7行当前追溯。全部断言、artifact与requiredCase数量保持；真实差分的SDK篡改负例在重算合法inventory后仍被严格拒绝。新的producer→consumer普通/race（含frozen子进程race）和双vet均exit0，两轴审查0发现。旧A诊断没有重跑或改写；这是后续生产者的定向证明，不授予新完整认证。

后续 release 源码修复不在三份 M3 的263文件及三份 GLM R1的217文件执行映射内，相关自身实际证据仍适用；L的1034文件广域来源与五项 Jev 广域 Go 来源则只能保留历史适用性。后续定向验证或采用旧 bundle 的只读诊断都不能充当新源码的完整 ordinary/race 执行认证，原3d3失败记录不改写。

[最终两轴审查索引](evidence/ticket-09/final-reviews/README.md)分别保留固定基点ecfd4b4以来的源码审查、便携校验器修复复审、关闭后实际证据审查、消费者修复与生产者修正复审。源码审查通过不替代真实支持与完整发布认证。

## 证据保留与复核

全部实际候选、正式、GLM原轮/R1、其余15项、原始完整普通/race及失败先行修复均有不可变归档、逐文件 Hash 和无 Key 复核入口。大原始归档保留在主仓库 `.evidence/minimax-retained/`，有限索引和完整压缩清单纳入 Git；各 README 明确外部保留文件、复核命令和可迁移路径。校验 Hash 通过只证明保存完整，不升级原来的 FAIL/NOT_RUN。

早期中国区和国际站研究、历史失败/中断、目录激活及受控检查点仍在[票09历史证据](evidence/ticket-09/README.md)、[票10历史证据](evidence/ticket-10/README.md)与[票11](evidence/ticket-11/README.md)中。原CN401、9481a6a的12304 PASS/52 FAIL/exit-15，以及后续12378 PASS/11 FAIL均保持原状态。GLM便携校验器的有界解压修复保留[原程序及 red/green](evidence/ticket-12/retention-bounds/README.md)，旧索引继续对应原字节，当前校验器 Hash 独立登记。

两处实现工作树已在原字节保留全部忽略输入后清理，[清理收据](evidence/ticket-09/worktree-cleanup.json)记录恢复位置和分支。

完整交付仍为 PARTIAL。剩余真实支持缺口需要可用且获准的服务配置/上游结果与另行明确的有限调用合同；不能重置历史消费、降低验收或自动新增补正轮次。
