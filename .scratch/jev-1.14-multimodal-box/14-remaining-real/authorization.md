# #14 剩余真实路线授权与有限预算

2026-10-17，用户要求核对根`.env`可用key，只报告变量名，按Provider整理缺失授权与有限预算，
明确允许发起真实调用直到整票通过。发现非空且可解析的相关变量名如下；不记录值、长度、前后缀或指纹。

| Provider | `.env`变量名 | 本轮任务 | Calls / Attempts上限 |
| --- | --- | --- | ---: |
| OpenAI | `OPENAI_KEY` | responses32、chat32、images4、Decisions12、中文Luna24、正式视觉Decisions12 | 116 / 116 |
| Anthropic | `ANTHROPIC_KEY` | messages32 | 32 / 32 |
| Google | `GEMINI_KEY` | gemini32、interactions-image4 | 36 / 36 |
| DeepSeek | `DEEPSEEK_KEY` | responses32、chat32 | 64 / 64 |
| TypeSafe | `TYPESAFE_KEY` | 本轮复用已验证当前源码的两支持、两效果，不额外调用 | 0 / 0 |

首轮上限248 Calls/Attempts，各路线原token、题数、图片、时间和重试限额不扩大。
通用聊天/图片最多每场景一次既有环境故障重试，计入原额度；Decisions和两效果零重试。
Host为聊天进程设置30分钟测试时限、35分钟外层进程组时限；图片10/11分钟；
Decisions180/210秒；中文工作600秒/外层630秒；视觉240/270秒。
显式取消或外层超时都保留状态并回收进程组，不能算作完成。

授权来源是本次用户指令。没有账户ID、区域、型号/API权限内省或金额硬限记录，
因此账户/区域/权限均记为UNVERIFIED；Decisions使用已有显式allow_unknown分支，
不编造permissions JSON或spendLimitUSD。视觉dedicatedCredential只代表独立注入。
所有实际服务端授权拒绝和型号不可用都保留FAIL，并记录所需厂商侧解锁条件。
没有提供美元硬上限；Calls/Attempts、输出token、图片与时间构成本机执行硬边界，
目录估价不冒充账单或金额熔断器。官方Endpoint、正常TLS、无环境代理或endpoint override。

为使持续修复有全局边界，最多预留一轮全量重验证302次：OpenAI116、Anthropic32、
Google36、DeepSeek64、TypeSafe54；仅在已诊断并完成所需修复后使用，不自动原样重试。
本次新增总上限550：OpenAI232、Anthropic64、Google72、DeepSeek128、TypeSafe54。
之前TypeSafe的102次独立历史调用不计入本次新增额度。各轮余额不跨任务挪用。
硬权限/不存在型号等外部阻断不能靠重复调用解决；不通过换型号、扩大UNSUPPORTED或改验收凑PASS。

## 实现前失败清单

- 输出key值、指纹或完整.env；shell source执行.env；未选Provider秘密进入子进程。
- 构建/核验/门禁继承凭据，或真实子进程携带多个key、代理或endpoint override。
- 入口选择过宽意外触发其他Provider；Decisions走错候选或正式身份。
- 单次步骤被重复执行、SDK/外层暗中重试，预算未记已失败或状态不明的进程。
- 型号、必交能力、目录Hash、固定效果任务或质量通过线被改变以冒充成功。
- 源码变化后复用过期TypeSafe/Decisions/效果包；重写旧失败与已经封存的历史回执。
- 超时/取消后残留子进程，包缺清单/审计却被接纳，或错误/usage/Observer不完整。
- 不完整或失败报告污染全局矩阵；其他行被连带更新；收集脚本退出0被误作整版PASS。
- 实际CLI未执行完整normal/race、Pi、压力、两个公开入口、审计和快照，却关闭#14。

先无凭据prepare，固定600个Go文件及目录/配置；每步独占intent标记、白名单环境、
独立进程组及实际退出码，保留所有原始包。所有真实步骤结束后清除进程key引用，
再进行严格公开verifier、按路线合并当前PASS及完整releasegate；旧封存包保持原样。
