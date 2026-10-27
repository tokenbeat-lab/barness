# Grok 正式 live 冻结计划 v7

Status: PROPOSED / NOT_AUTHORIZED / NOT_RUN

本计划只提出新来源的一次17路线正式验证。沿用维护者已选择的“收集全部独立结果”：各路线串行只启动一次；已确定的执行/资格失败保留FAIL并继续未运行路线；来源/输入/凭据隔离异常、取消或无法可信确认的采集/进程状态全批停止。不重发权限或自动重跑。六条Kimi仍候选，不连带晋升；race、压力、发布均不在范围。

## 实际已完成与余项

v6一次17条已全部执行，14正式资格通过、2资格器拒绝、1原生失败。238必交实际236PASS+1固定UNSUPPORTED+1FAIL，新270逻辑调用/267HTTP，全部账本和失败原件保持。此前正式241/240另计，历史正式合计511/507（候选另计）。[完整检查点](../../.evidence/barness-ai-grok/live-batches/gf6-20261026T233415Z-ceb8d8c1/completed-checkpoint-v1.json)及[独立总账](../../.evidence/barness-ai-grok/live-batches/gf6-20261026T233415Z-ceb8d8c1/independent-audit-v2/barness-grok-v6-final-audit-v2.json)均保留。

20已修复新ProcessGrant材料误入legacy origin的资格分派：新授权始终完整校验，历史材料仍按旧origin校验，无无效新授权回退或foreignmodel绕过。三个协议真实公共CLI先RED后GREEN；45新场景及53受影响主回归/6辅助通过；[v15双轴](review/review-v15.md)0/0。原生场景、模型、SDK、预算、重试及门槛未降低。Messages原thinking-on已有非空推理后发生响应体读取中断，控制请求的初始EOF已按原合同恢复；确切vendor/代理/网络原因无法确定，旧FAIL不改。

软件构建pin e1f6986677c905b48d80a6b01fc0c0f125427c82，2064默认来源。四个native与四个CLI实际构建PASS、模板及17无Key/无grant入口均PASS/0请求/0claim；准备脚本两次元数据错误及成功原件分开保留。Generic/Regional538、xAI539、M3407原生source改变；GLM342原生map未变，但全局Stage来源/输入已经变化，旧14条不冒称当前来源资格。[真实构建](../../.evidence/barness-ai-grok/build/20261027T003444-authority-freeze-z_mdumhu/formal-build-receipt.json)；[无Key准备](../../.evidence/barness-ai-grok/build/20261027T003444-authority-freeze-z_mdumhu/qualification-preparation-receipt.json)。

## 一次匹配完整普通前置

[有限清单v12](verification/cases-v12.json)1568主必交/129入口，全部旧1522保留+45新MM20+受影响既有MM10公共工具wrapper；GLM32和MM18六辅助仍须独立有效。119实际额外读取不变，单列原生合同文档新hash另固定。

为避免两次仅授权输入不同的完整普通，本轮先取得具体付费范围授权，生成新36份授权/grant/config/index并加入129项基础输入，再冻结同一提交/165输入，执行唯一一轮完整四包普通。必须实际exit0、1568/129/38完整PASS且全部原件/source/input匹配，才启动任何厂商请求。现在新来源完整普通明确NOT_RUN，不拿定向GREEN或v6旧普通替代。若预检失败，保留失败并零厂商调用；源码/输入修复回SOP3，不自动改旧批准pin。

## 固定路线和上限

| 顺序/路线 | 密钥来源 | 场景 | calls / HTTP上限 |
| --- | --- | --- | --- |
| 1 xai-responses | .env:GROK_KEY | 18 | 24 / 24 |
| 2 openai-responses | .env:OPENAI_KEY | 9 | 32 / 32 |
| 3 deepseek-responses | .env:DEEPSEEK_KEY | 9 | 32 / 32 |
| 4 openai-chat | .env:OPENAI_KEY | 11 | 32 / 32 |
| 5 deepseek-chat | .env:DEEPSEEK_KEY | 12 | 32 / 32 |
| 6 zhipu-chat | .env:ZHIPU_KEY | 24 | 24 / 32 |
| 7 zai-chat | .env:ZAI_KEY | 20 | 24 / 24 |
| 8 minimax-chat | .env:MINIMAX_KEY | 14 | 20 / 24 |
| 9 minimax-cn-chat | .env:MINIMAX_CN_KEY | 10 | 24 / 24 |
| 10 anthropic-messages | .env:ANTHROPIC_KEY | 12 | 32 / 32 |
| 11 zhipu-messages | .env:ZHIPU_KEY | 24 | 24 / 32 |
| 12 zai-messages | .env:ZAI_KEY | 20 | 24 / 24 |
| 13 minimax-messages | .env:MINIMAX_KEY | 14 | 20 / 24 |
| 14 minimax-cn-messages | .env:MINIMAX_CN_KEY | 10 | 24 / 24 |
| 15 google-gemini | .env:GEMINI_KEY | 7 | 32 / 32 |
| 16 minimax-responses | .env:MINIMAX_KEY | 14 | 20 / 24 |
| 17 minimax-cn-responses | .env:MINIMAX_CN_KEY | 10 | 24 / 24 |

238完整必交要求237PASS+1固定声明UNSUPPORTED/零请求（OpenAI Chat reasoning-history）。其他全部必交，包括每条xai-origin-history，仍须PASS；不将旧原生FAIL改成UNSUPPORTED或借控制/回放取得真实资格。

新批最多444逻辑调用/472HTTP尝试，历史511/507另计（合计最坏955/979）。Grok为5USD响应估算累计停止阈值，不是账单硬封顶；其他路线沿已批准的原生token/call/attempt预算，无USD硬上限，未知用量保留未知。每进程30分钟，17串行原生时限上界8.5小时；普通、采集与资格消费另计。有限原生请求重试/重跑按机器计划原合同，不新增进程重跑或备用grant。

只读取对应root .env变量，仅所选服务密钥进入该路线进程，不记录密钥或其hash。每路线显式代理http://127.0.0.1:7890；账户/地区别名、MiniMax国际standard API/M3资格沿用已确认事实，不另索取密钥。原生通过后运行自身正式资格器（地区-formal、国际M3-formal-publish）再运行已编译releasegate的schema2只读消费；publish仅原生投影，不表示外部发布。

## 执行与授权边界

采集器工作目录始终仓库根；原生argv已固定为/usr/bin/env -C /Users/cyber/RestoX/barness/ai/live 加绝对native路径和原始精准flags。新grant路径必须绝对指向同一固定文件，producer SHA独立于env launcher SHA。Scope直接派生真实原生模板，不能复制过期二进制hash。[机器计划](verification/formal-plan-v7.json)固定模型/SDK/endpoint、完整场景与outcome、所有限额及构建。所有模板ID为空、expiry为零，当前未授权。

旧v4的SOP3/4例外不延续；v6一次权限已经消耗。新一次必须有独立原始回复、新ID/有效期与新grant/claim，不能改ID/expiry绕过同一批准的scope guard，不能复用closed ledger。新计划不含备用或运输失败后的第二个进程。全部17本源正式资格有效后，才准备另行由人选择race，再依序选择压力；发布保持false直到全部门槛有效。
