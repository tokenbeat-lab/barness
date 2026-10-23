# 候选付费执行结果03：35秒续验，Responses流内过载后停止

本次依用户“拉长间隔时间，继续完成剩余验证”执行[范围04](candidate-live-review-04.md)，[原始授权03](../../../.evidence/barness-ai-kimi/authorizations/03/approval.json)。新grant每次调用前等待35000ms，含首项；原生来源/二进制绑定且零重试，首失败即停止整个范围。没有后续自动新一轮。

| 路线 | 实际候选结果 | native / 自身-qualify退出 |
| --- | --- | --- |
| 国际站Chat | 9/9 PASS，含此前429的simple-complete、历史、工具、取消与恢复 | 0 / 0 |
| 国际站Responses | 前3项PASS，第4项simple-complete FAIL，后5项NOT_RUN | 1 / 1 |
| 国际站Messages | 9项NOT_RUN，未启动producer、未读其路线凭据 | 无 |

合计13调用/13HTTP Attempt、12项PASS/1项FAIL/14项NOT_RUN、零重试。13个HTTP状态均200，未再出现HTTP429。预算预留US$0.964002，不是实际账单，实际费用未知。两份grant永久消耗，四份可信ID/scope账本原件与副本独立逐字节/SHA核对一致；Messages旧新授权ID均不得当成自动重试权限。本范围已关闭。

## 新失败的直接证据

请求 `POST https://api.moonshot.ai/v1/responses`，Kimi K3、`stream:true`、`reasoning.effort:high`、`max_output_tokens:4096`，固定用户文本 `Reply with the exact short text regional ready.`。HTTP200、`text/event-stream`，9952ms后结束；流先created/in_progress，再给出：

```json
{"type":"error","error":{"type":"rate_limit_exceeded","code":"rate_limit_exceeded","message":"The engine is currently overloaded, please try again later","param":null},"sequence_number":2}
```

紧接 `response.failed`，status=failed，response.error.code同为rate_limit_exceeded，output=[]、usage=null。该失败的服务端明示原因是引擎过载；不能仅凭同类rate_limit名称把它等同于前次organization RPM=3，也不能将HTTP200当成功。没有证据进一步区分后端容量、区域或账户内部调度。

request ID `9d733f47-ce8c-11f1-b2aa-00163e5117db`；trace ID `e82ed917623edfd62fa48af5e890afe0`。完整脱敏请求/响应：[导出](../../../.evidence/barness-ai-kimi/diagnose/07/international-responses-overload.redacted.json)；原件 `.evidence/barness-ai-kimi/live/kimi-intl-responses/20261023T024838.748565Z-kvxhak0z/bundles/20261023T024838.772894000Z/REGIONAL-kimi-intl-responses-kimi-k3-simple-complete/exchanges.json`，SHA `8b9cc18bf9f43fcc8add2775f38406d8a5b0b7e10960afd517a5d8ee344f6d63`。授权头不交付秘密值。

Result正确保持StopReasonError，Usage/Cost unreported、资源bodies/calls/events/permits/waiters均0；失败断言为own legal turn succeeds、reported usage retained、finite reliable cost bounds available。当前安全诊断文本是 `Kimi stream error (unknown)`，嵌套错误分类另记[11](../issues/11-responses-stream-error-diagnostics.md)待公共回放核对，不推定已修复或凭此忽略服务端失败。

## 可重复核对入口

- [完整执行索引](../../../.evidence/barness-ai-kimi/paid-execution/03/execution/summary.json)：实际argv、数字退出、时间、路由回执、预算、授权、原件SHA。执行提交 `9c0e5205788b0fdcf73cda95d9191b386868b1a2`，起止 `2026-10-23T02:42:32.989094+00:00` → `2026-10-23T02:51:25.896665+00:00`。
- [独立时序与断言核对](../../../.evidence/barness-ai-kimi/paid-execution/03/postprocess/verification.json)：实际首项锚点和完整四事件时序；Chat最小35002.712ms、Responses最小35002.986ms，均≥35000。已执行13项逐项核对，NOT_RUN不伪造断言。
- [13次HTTP及4份账本副本库存](../../../.evidence/barness-ai-kimi/paid-execution/03/postprocess/http-and-ledger-inventory.json)：一项一次尝试、两audit零发现。候选只运行-qualify，没有送正式消费者。
- 普通验收实测 `9c0e5205788b0fdcf73cda95d9191b386868b1a2`，exit0/PASS，14071必需ID全部PASS、214重负载NOT_RUN、14943实际Go PASS节点，墙钟448.241秒。1359源码/148输入及85/6/5消费目录前后一致；主与五辅助audit通过。见[核对06](../../../.evidence/barness-ai-kimi/acceptance/06/verification.json)。
- native source `2f1513067bedf1431f1e702b6b24e222543f22bf992e01e7bd64fba3399dd1b4`，binary `00b595e39dd242ead429123c45b49e217648ec4b56d3e002cc28f9767e1c4ff6`；[双轴复核](pacing-code-review-02.md)Standards0/Spec0。准备清单工具的两次只读失败保留于`.evidence/barness-ai-kimi/preparation-failures/01/`与`02/`，未启动测试/厂商。

本次未完成剩余两路线。后续须由人选择新的Responses完整9项、Messages完整9项范围，并使用新grant；不拼接Responses前三项或复用失败授权。新来源的中国站重验、六路线formal、晋升、旧路线、race/压力均未执行，release.pass=false。旧失败/记录/pin不改写，实际唯一状态见[spec](../spec.md#唯一当前状态)。
