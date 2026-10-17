# #14 TypeSafe 当前真实验证

2026-10-17，用户要求按缺口清单验证，并明确指定根 `.env` 的 `TYPESAFE_KEY`。
仅执行 TypeSafe 四项。首轮 51 Calls/Attempts 后，完整门禁发现 UC13 历史矩阵测试缺陷；
用户另行明确授权修复并用同一 key 再执行四项一轮、最多 54 Calls/Attempts。
修复后的第二轮也实际 **51 Calls、51 Attempts、零重试**；两轮合计 **102/102、零重试**，
每轮各不超过 54，独立计账，不从首轮余额自动补采。其他厂商凭据未读取或注入。
所有真实调用已结束；各轮未用的 3 次额度不转给其他任务。

正式目录：`2026-10-17.1`，
`sha256:ae3315ce504af2019ce6476f3b021093e13428a5992b3e5fc7d3208576eb49c7`。
第二轮源码提交 `541ef2d`，支持矩阵提交及完整门禁执行提交 `4d191fa`。
完整 Go 集合仍为 600 文件，仅矩阵 E2E 相对首轮改变；目录和固定效果任务未调整。
首轮执行起点 `d4e61e2`、矩阵提交 `701bd69`，其 Go 内容为 `cc78cfd`，只作为历史证据。

## 实际结果

| 独立任务 | 结果 | 实际 Calls / Attempts | 原上限 |
| --- | --- | ---: | ---: |
| Jev 1.13 协议 | 3 个原场景 PASS | 3 / 3 | 6 / 6 |
| Jev 1.13 中文效果 | COMPLETE，24/24 答对 | 24 / 24 | 24 / 24 |
| Jev 1.14 正式协议 | 12 个场景 PASS，含 plain/hooked | 12 / 12 | 12 / 12 |
| Jev 1.14 正式视觉效果 | COMPLETE，固定 12 样本 | 12 / 12 | 12 / 12 |

1.13 保留原 `context-422` 场景名；实际是原规则允许的 HTTP 400 / max_tokens_exceeded，
不声明服务返回了 422。协议用量为 714 输入、92 输出 token。
中文固定四类任务为每类 6 条，准确率 1、ECE 0.0004166667、Brier 0.0000083333；
输入 12,843、输出 1,080 token，目录已知估价 USD 0.000539406。

视觉公共 8 样本及局部 4 样本的分类、存在性均全部答对。Box 的 2 个存在样本都有框，
按存在样本作分母的平均 IoU 为 **0.6840596383**；2 个不存在样本无被消费的误报。
原响应对不存在样本仍给出 2 个框，由独立存在性结果阻止消费，原始输出完整保留。
视觉输入 10,015、输出 716 token，目录已知估价 USD 0.000420630。
这些是固定合成任务的实测分数；没有质量通过线，也未重跑样本挑分数。
两项效果估价之和不代表本轮四项的完整费用，更不代表厂商账单。

1.14 协议实际累计 20 题、18 sources、26 references，源图 44,784 bytes、
展开图 64,688 bytes、State 1,670 bytes，均在原限额内。
视觉实际 28 题、16 occurrences，源图 27,959 bytes、展开 41,603 bytes。
两项正式任务及中文效果均零重试、终态和用量完整。

## 核验和支持记录

证据根目录（仓库相对路径）：
`.evidence/ticket14-typesafe-r2-20261017T061932.423266Z/`。

- `protocol113/*/`：原始 1.13 协议、捕获、Attempt/Observer、清单与审计。
- `text-jev/*/`：中文效果、24 样本、捕获重放、usage 和完整性清单。
- `protocol114/*/`：原始正式协议、12 场景、预算预留/使用、清单与审计。
- `visual-jev/`：固定视觉任务、响应、评分、Box 消费、usage、清单与审计。
- `support/`：原协议和视觉逐字节复制的正式封装及独立完整性清单。
- `matrix-before.json` / `matrix-after.json`：仅 TypeSafe 的两行发生改变，其余 9 行不变。
- `*-command.json` / `*.log`：各步的实际参数、凭据变量名、时间、退出码和脱敏日志。

无凭据的中文 verifier、视觉 `-require-live` verifier、正式封装和封装
`-require-live` verifier 均退出 0。支持矩阵随后合并两条各自当前的正式支持记录。
原始包和正式封装审计均为零发现；封装与生产者文件保持一致。

账户、区域和厂商权限未做内省，别名明确标为 `unverified`。
`dedicatedCredential` 只描述本轮隔离注入选中的 key，不证明账户专属或厂商权限。
用户授权了这轮有限调用；未提供美元硬上限，授权记录的 `dollarCap` 为 null。
所有构建、verifier、封装、合并和完整门禁均未继承或注入任何 key。

## 完整门禁

完整无凭据发布命令已完成；normal/race各 **9,098 PASS、0 FAIL**，
每阶段包含861个冻结Pi差分、33个UC14用例，均PASS。原local/cloud压力及两公开入口
图片设计负载完整执行，原64MiB、4并发等限额未缩减。无凭据live-tag harness为
204 PASS、62 NOT_RUN，审计零发现；其中未执行的真实场景仍保留NOT_RUN。

| 发布门禁 | 实际结论 |
| --- | --- |
| go-test、go-test-race | PASS / PASS |
| go-vet（含live tag）、P0、冻结Pi差分 | 全部PASS |
| 聚合脱敏审计 | PASS，17个实际输入审计均零发现 |
| snapshots | PASS，原压力#07以resolved当前状态和Hash进入本次快照 |
| live | BLOCKED，TypeSafe两条PASS，其他9条NOT_RUN |
| required-artifacts | BLOCKED，仅中文Luna和正式视觉Decisions仍NOT_RUN |
| traceability | BLOCKED，保留依赖尚缺真实路线/效果的条目 |

实际`gate/release-report.json`的`pass=false`，CLI退出1；收集脚本自身退出0只表示流程收尾，
不代表整体发布通过。七项内部门禁全部PASS，剩余三个阻断门禁均源于未获授权的其他路线/效果。
完整命令、参数、开始时间、耗时与退出码保留在根命令回执和报告中，不改写历史报告。
该命令显式选择两条真实 TypeSafe 支持输入、两项真实 TypeSafe 效果输入，
其他 9 个支持输入和 2 项效果输入为本轮独立的零调用 NOT_RUN 包。
`no-key/` 中另有未被选作 TypeSafe 当前证据的 NOT_RUN 预检；保留这些占位事实，
不会拿它们覆盖本轮真实成功，也不会用真实成功改变其他路线的状态。

## 首轮失败和修复边界

首轮目录 `.evidence/ticket14-typesafe-20261017T055619.971418Z/` 原样保留，实际四项均完成，
51/51、零重试，文本24/24、视觉12/12分类和存在性正确，Box平均IoU为0.7014543034。
该分数不替代新源码第二轮的0.6840596383，两个原始包均保留。

首轮完整normal为9,097 PASS、1 FAIL，唯一失败为
`UC13-model-specific-matrix-history`。原测试将初次迁移的快照逐字段比较施加于当前矩阵，
错误禁止1.13合法重测；单例也退出1，独立Spec核对确认本轮合并合法。
用户追加授权后停止该已确定失败的旧门禁，Python外层退出130，命令回执为INTERRUPTED，
底层exitCode为null；race未完成、manifest/audit均MISSING，没有完整发布报告，不追认完成。

修复保留原before fixture，after取自最初纳入提交`527f633`，JSON值完全一致；
完整旧行状态/时间比较准确落到一次迁移快照，当前两型号身份独立。
公开CLI在临时矩阵验证1.13后续合法重测只改变自身，其他十行逐字段不变，
重复陈旧报告拒绝且文件不写，仓库矩阵不被离线用例修改。单例GREEN退出0。
缺陷、来源Hash和RED/GREEN证据见
[续修03](../../barness-ai-release-verification/issues/03-jev-matrix-history-current-runs.md)。
首轮源码按当时Git对象核验，第二轮按当前600文件核验，旧包不能借新一轮继续声称当前PASS。
当前公开`jevformal -verify OLD_SUPPORT -require-live`也已无凭据退出1，明确拒绝
`formal source snapshot drift`；命令、退出码和脱敏日志保存在RED/GREEN诊断目录。

## 剩余授权范围

#14 继续 `needs-info`：两项整体验收条件仍需其他路线和效果的当前真实包。
TypeSafe 这四项无需再采集；新增 Go 源码、目录或固定任务变动后须重新判断包是否仍有效，
不得直接改写已经完成的包或自动再花费。

| 剩余路线 / 任务 | 有限一轮上限 | 凭据与授权依据 |
| --- | ---: | --- |
| openai-responses、openai-chat | 各 32 Calls/Attempts | 清单 A，各自完整型号集合和 API 授权 |
| anthropic-messages | 32 | 清单 A，六个固定型号 |
| google-gemini | 32 | 清单 A，固定型号 |
| deepseek-responses、deepseek-chat | 各 32 | 清单 A，各自固定型号集合 |
| openai-images、google-interactions-image | 各 4 | 清单 A，各自生成/编辑场景和图片额度 |
| openai-decisions | 12 | 清单 B，gpt-6-luna，权限记录或显式接受 unknown 分支 |
| 中文 Luna | 24 | 清单 D，独立固定 24 样本 |
| 正式视觉 Decisions | 12 | 清单 E，独立固定 12 样本 |

剩余合计上限 **248 Calls/Attempts**，其中 9 条支持路线 212、2 项效果 36。
这只是待授权的原额度，不是本轮继续执行的授权。需要明确每项凭据来源、账户/区域、
型号/API 权限或接受其未知状态、一次执行窗口、有限调用及账户金额约束。
金额不能由调用数推导；现有通用 harness 没有统一美元熔断器。
原有可重试路线的有限重试计入表内上限，Decisions 和两项效果保持零重试。
完整映射、endpoint、token/图片/时间上限及环境变量见
[授权预算清单](../14-live-authorization-checklist.md)。

## 重复核验

重复核验命令（不会联网调用模型）：

```sh
python3 .scratch/jev-1.14-multimodal-box/14-typesafe-real/seal.py --verify
```

封存覆盖两轮全部存储字节、实际命令/退出码、原始 manifest/audit、正式封装、
效果清单、支持矩阵转换、源码/配置和实际门禁。仅排除证据根目录下三个可重建二进制，
`no-key/` 中的嵌套二进制仍纳入完整索引。
故意崩溃的 live harness 子包保留 MISSING audit 的事实，不能成为完成的验收包。
旧失败、旧压力状态快照和旧发布报告不改写；旧回执针对其记录版本，
当前源码及矩阵变化由两轮各自source pin、before/after和新回执单独封存。
