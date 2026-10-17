# #14 全 Provider 真实验证与整版验收

当前状态：恢复轮真实采集中，整版门禁尚未完成；不得据本文件草稿关闭工单。

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
外层超时/取消回收整个进程组，不能自动重发收费步骤。初轮已用120，因此本次可执行最大
为120+302=422，仍低于550总边界；余额不授权第三轮。

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

待收齐实际15任务、严格核验、矩阵合并及完整发布结果后，替换本段并关闭#14剩余验收。

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
