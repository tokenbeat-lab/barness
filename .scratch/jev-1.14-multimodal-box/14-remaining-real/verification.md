# #14 全 Provider 真实验证与整版验收

最终公开发布门禁：**PASS**，实际CLI退出0。11条当前真实支持及4项效果完成，normal/race各9098 PASS、0 FAIL，17个输入审计零发现。

## 授权、凭据与预算

用户明确允许按根`.env`现有key执行剩余9条支持和2项效果，直到整票通过。
只记录变量名，不记录值、长度、前后缀或指纹；逐行解析选定变量，不source dotenv。
每个收费进程只注入当前路线的一个凭据；构建、核验和发布门禁不继承任何Provider key。
官方Endpoint、普通TLS、无ambient代理或endpoint覆盖。

| Provider | `.env`变量名 | 初轮上限 | 唯一恢复轮上限 | 本次总上限 |
| --- | --- | ---: | ---: | ---: |
| OpenAI | `OPENAI_KEY` | 116 | 116 | 232 |
| Anthropic | `ANTHROPIC_KEY` | 32 | 32 | 64 |
| Google | `GEMINI_KEY` | 36 | 36 | 72 |
| DeepSeek | `DEEPSEEK_KEY` | 64 | 64 | 128 |
| TypeSafe | `TYPESAFE_KEY` | 0，曾计划复用R2 | 54 | 54 |
| 合计 Calls/Attempts | — | 248 | 302 | 550 |

凭据存在不等于账户、区域、厂商权限或美元硬限已内省。相关字段保持unknown/UNVERIFIED；
不存在Host permissions文件或虚构spendLimitUSD。Decisions使用已授权的显式allow_unknown分支。
视觉dedicatedCredential只代表独立注入。未提供美元硬上限，目录估价不充当账单熔断。
所有原token、图片、题数、时限和重试合同保持。恢复轮只预约一次，各步骤独占intent，
外层超时/取消回收整个进程组，不能自动重发收费步骤。初轮已用120、唯一全量恢复已用179，
追加确认前共299次。随后唯一DeepSeek Chat同源码确认按用户持续授权，在原Provider/任务
总上限内另预约32次，最多可达331次；实际最终312次。不启用第三个全量轮，
此例外已向用户说明并记录在授权文件。

## 初轮及修复证据

初轮根：`.evidence/ticket14-remaining-20261017073529Z`。
源码清单607项（600Go+7配置），基准提交`d21aa99`；运行期间Go和配置未变。
构建/paid执行器为`d21aa99`；`43ac874`补强TypeSafe完整Go集合guard，并以无凭据
公开verifier再次核对R2支持、中文和视觉包。执行器两个版本按Git保存到初轮根，原报告不重写。

初轮8条支持和2项效果均完成，实际120 Calls/Attempts、零重试。唯一失败是正式Decisions：
9场景config FAIL、0请求、audit零发现。未设置可选代理却无条件调用仅接受显式loopback
的解析器，导致`invalid explicit loopback HTTP proxy`。这不是服务端权限拒绝。

续修04先列失败方式，再扩展原候选与正式进程E2E seam：候选Host记录直连、候选unknown
直连、正式unknown直连先3项RED；随后最小修复只在非空时校验代理，空代理明确直连且
不继承ambient代理。8项直连/loopback/非法remote-proxy回归GREEN。Go内容提交`9a8ea45`。
完整正式Jev源码快照包含测试文件，因此旧1.14正式协议/封装过期；中文和视觉适用性并不
机械包含所有测试文件。本轮选择统一重新采集全部11支持与4效果，旧包保持历史身份。

诊断根：`.evidence/ticket14-decisions-direct-20261017`，含RED/GREEN命令、原3文件快照、
完整case清单及审计。恢复前首次完整无凭据harness为206 PASS、62 NOT_RUN、1 FAIL，
仅FIFO进程2秒时限失败；保留原FAIL。源码与时限均不改，独立同例PASS；新CLI
`-help`约0.524s，实际FIFO拒绝约0.0114s、退出1。未捕获首个超时内部栈，根因不认定已证明。
同源码完整一次复核为207 PASS、62 NOT_RUN、0 FAIL，才允许恢复收费。两次harness各自
保留两个刻意崩溃子进程的缺audit事实，不将它们当已完成接受包。

## 恢复轮与当前整版结果

恢复根：`.evidence/ticket14-full-recovery-20261017074119Z`。
当前Go内容`9a8ea45`，600Go与7配置独立固定；任何变化都会停止下一步收费。
初轮消费Hash和唯一恢复预约保存在初轮根`recovery-reservation.json`。

恢复轮15任务实际179 Calls/Attempts，零重试：除DeepSeek Chat一次推理回放断言外全部通过。
该Chat原始请求明确thinking enabled/effort low，原始SSE非空reasoning块为0且reasoning_tokens=0；
首轮同场景请求字节完全相同，有18块非空推理并成功回放。客户端未丢弃服务端状态。
原12次失败包完整保留，未修改Go、提示、型号、断言或UNSUPPORTED规则。
`recheck.py`在既有550总额与DeepSeek128内独占预约一次完整路线≤32次，实际13次、
11/11 PASS；使用新目录`deepseek-chat-recheck`，旧失败和原allCommandsSucceeded=false保留。
最终选定包仅采用这次新的完整有效Chat报告，不拼接不同进程场景。

本次新增总计312 Calls/Attempts：初120 + 全量恢复179 + 单路线确认13；
各包SDK/场景内retry计数均0，另行安排的整路线复核13次明确计入总额，不隐藏为零消耗。
之前TypeSafe102次属于独立历史授权，不计入本次新增。

| Provider | 本次实际 Calls/Attempts | 本次总上限 | 补充授权阻断 |
| --- | ---: | ---: | --- |
| OpenAI | 143 / 143 | 232 / 232 | 当前必交调用均通过；账户/区域/privilege内省未知 |
| Anthropic | 40 / 40 | 64 / 64 | 当前必交调用均通过；同上 |
| Google | 20 / 20 | 72 / 72 | 当前必交调用均通过；同上 |
| DeepSeek | 58 / 58 | 128 / 128 | 当前必交调用均通过；空推理失败原包保留 |
| TypeSafe | 51 / 51 | 54 / 54 | 当前必交调用均通过；账户/区域/privilege内省未知 |

11个当前包经公开supportmatrix命令整批原子更新，提交`6eb440a`；全部行独立更新，
Jev1.13、Jev1.14和Decisions的型号身份保留。没有用候选或其他型号替代必交能力。
OpenAI Chat的reasoning-history仍为原合同明确的UNSUPPORTED，其余必交场景均PASS。

当前Jev/Luna中文各24/24正确，两路线protocolStatus均COMPLETE；正式视觉每路线12样本
COMPLETE。两路线共有8样本的分类/存在性全正确；Jev的4个local样本与Decisions的4个
supplemental context样本各自全正确，但两组不是局部能力等价对照。
Jev当前Box present2、absent2，MeanIoU=0.7123740342276741；消费后absent误报0，
仍保留2个raw absent box被存在性门控抑制的事实。分数只描述固定合成小任务，无新增质量阈值。
两文本、两视觉及Jev正式封装均通过公开严格verifier，视觉/封装均require-live。

完整无凭据releasegate已完整结束，输入恰为11个当前支持、2文本和2正式视觉包；
normal/race实际manifest均为9098 PASS、0 FAIL，各含861 Pi及33 UC14，完整压力与两入口
图片负载在其中；两阶段审计均0发现。实际报告`gate/release-report.json`为`pass=true`，
CLI退出0，执行报告提交`fc8f02b`（Go内容仍`9a8ea45`），29个本轮命令状态逐一保留。
两种vet、P0、差分、live、脱敏、追溯、快照、required-artifacts等10项门禁全部PASS。
11条live行全部PASS；4项文本/正式视觉效果和原要求的资源/撤销/负载证据均被接受。
17个指定输入包聚合审计零发现；normal约493.18s，race约1858.04s，均完整执行，
没有使用`-bundle`复评来替代本轮normal/race。

目录、发布计划、有限classifier扩展表和原#07快照均current。原#07保持其既有resolved
和Hash `25bacdc0fddb101a79ebde7c6077dadfaddd1a9ff029529a876fd5335fc7a052`，
没有重写其原始因果证据或追认旧报告PASS。#14仅在全部真实包和本次整版门禁满足后
勾选最后两项并resolved；续修04随真实9场景及完整回归验收关闭。父规格/ADR的用户
未提交内容未被代改，未执行部署。

封存器另从原始报告重算120+179+13，校对550任务及128 DeepSeek上限、每路线预算、
实际步骤退出码、source-at-start的607项与Git对象、600个完整Go路径、必需唯一生产包、
两阶段真实9098/861/33覆盖、各case清单、视觉/文本/派生支持完整性及所有保留树。
历史TypeSafe两轮按原字节与旧索引保存，当前只接纳本轮新源码包。
所有收费调用已结束，不再自动补跑。

## 可重复核验

以下verify无收费调用；只在最终回执已封存后使用：

```sh
python3 .scratch/jev-1.14-multimodal-box/14-remaining-real/seal.py --verify
```

全新收费轮需新的明确范围与有限预算，不能重用已占用的intent或旧失败目录。
本轮recovery的`prepare/check/live/merge/gate`命令与实际argv/退出码均按步骤留存；
`gate`调用原公开releasegate完整执行normal/race，不能用`-bundle`复评替代。
当前/历史完整目录的字节索引存于`.evidence/ticket14-remaining-integrity`；
原TypeSafe两轮索引仍在`.evidence/ticket14-typesafe-integrity`。
索引只排除每根3个可重建二进制，嵌套捕获、错误包、故意损坏夹具与link身份均保留。
不覆盖旧封存器、回执、失败normal、主动中止race或历史pass=false。
