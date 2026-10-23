# 国际站候选续验范围04（用户已授权）

本会话用户直接指令：**拉长间隔时间，继续完成剩余验证**。本范围只落实国际站三条尚未完成路线；原第4项429与停止范围保持历史原件。此前预算上下文见[提案03](candidate-live-review-03.md)，本次新授权字节与来源绑定见 `.evidence/barness-ai-kimi/authorizations/03/approval.json`。不再使用旧grant。

原生版本已经两轴审查，执行前还必须通过新冻结完整普通验收并核对全部来源/输入集合。实际构建提交 `6a08f61578a8dc8999ee7a833c59b46d9ef036ec`；native source SHA `2f1513067bedf1431f1e702b6b24e222543f22bf992e01e7bd64fba3399dd1b4`；二进制 SHA `00b595e39dd242ead429123c45b49e217648ec4b56d3e002cc28f9767e1c4ff6`。完整工具/依赖/构建argv/数字退出见 `.evidence/barness-ai-kimi/preparation/04/build.json`；新native grant各自绑定来源、二进制、账户和价格原件。

| 路线 | Provider | 官方目标（grant endpoint） | 新grant SHA256 |
| --- | --- | --- | --- |
| kimi-intl-chat | moonshot | https://api.moonshot.ai/v1 | 2334e52dad87549754d8f50c766c4875f2393eddeba30b13a0e261cd54d5121d |
| kimi-intl-responses | moonshot | https://api.moonshot.ai/v1 | 1cba04350f67586c52796558ecbff6b8b802dfd61c289ae480c96f502077528f |
| kimi-intl-messages | moonshot | https://api.moonshot.ai/anthropic | 0097d1277d90be58f1e35b4b8e7b2cc672e8b3babfa049b6602d5c7e48790e63 |

全部使用Kimi K3、KIMI/GLOBAL账户别名，只转发所选 `KIMI_KEY` 对应凭据。每次公共调用前等待 **35000ms**，含首项；每条路线从头完整执行 full-stream、full-complete、simple-stream、simple-complete、native-history、tool-call、tool-result、cancel、recovery九项，不能拼接旧Chat前三项。

每路线≤9调用/9HTTP Attempt、零重试、估算预留≤US$1.50，三路线合计≤27调用、≤US$4.50。每次输出上限4096Token，每路线输出预留36864、输入预留131072；调用窗口120秒在等待后开始，进程1800秒及grant到期仍限时。到期 **2026-10-24T01:15:22+08:00**，不晚于此前披露期限。

三路线串行；首个实际候选执行或自身资格失败即停止整个本次范围，取消/失败grant永久消耗。只对每条实际发现bundle运行 `regionalverify -qualify` 并保留真实数字退出；候选不送入正式消费者。独立检查实际Observer四事件时序、首项及后续≥35秒间隔、九项非空断言、资源归零、账本与audit。实际费用未知，预算预留不能当账单。

本次不含中国站额外重跑、正式验证、晋升、旧路线、race、压力或失败后的自动新一轮。中国站旧来源候选保留，不能作为新native来源的资格。完整范围JSON：`.evidence/barness-ai-kimi/preparation/04/candidate-live-proposal.json`。
