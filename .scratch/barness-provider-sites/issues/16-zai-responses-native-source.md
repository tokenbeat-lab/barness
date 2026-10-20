# 16 ZAI Responses 预先固定自身文本推理来源

Status: claimed
Type: task

## Context

[07最终验收](07-final-acceptance.md)、[12同类Chat来源修复](12-zai-chat-native-source.md)、[规格R06/R17/R18/R20](../spec.md)。本票修正验收harness及其原生来源合同，不把原真实失败改写为成功。

## Actual failure

固定恢复3在实际6b87a584来源完成M3三路39/39与ZAI Chat18/18；ZAI Responses完整18场景17 PASS/1 FAIL，`glm-5.3/native-history`断言`native=false`。全部走显式代理7890、零重试，失败进程实际18 Calls/18 HTTP；父UNKNOWN和US$0.65全额占用不改。[停止索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery3-stopped/index.json)固定1012原件，跨轮实际378/374、claimed564/612、整笔12.60USD；后19进程未启动。

失败断言检查预先固定的tool-call源消息。该实际源SSE没有reasoning item/事件，保存Thinking0且reasoning_tokens=0；同进程full-stream自身保存完整推理/原生状态。旧成功tool-call原件由当前公共Client正确保留；新旧最终tool-call请求逐字相等。没有parser回归或代理丢帧的证据。

## Required fix

沿用12已验证的设计，将ZAI Responses的原生推理源也预先固定为同轮同型号full-stream，保持独立tool-call/tool-result链与Host标记；不得事后挑成功消息、从兄弟协议借来源、伪造/拼接原生状态、降低推理/零downgrade断言或将真实FAIL推广。harness记录与`supportmatrix.RegionalNativeSourceCase`必须同步，严格消费者仍核对原件和精确续接前缀。

先公共CLI/Client受控失败覆盖“工具响应合法但无推理，先前文本流有自身完整推理”，再实现并跑受影响完整六路线、原生消费者及拒绝场景。生产协议/请求/SDK/模型目录不改。supportmatrix消费者进入MiniMax执行来源，因此三M3与六formal需新的自身资格；GLM245逐项相等才可保留。不能缩小任何来源选择器以绕过重验。

## Paid boundary

固定恢复3已按实际批准的失败即停条件停止。本票的离线实现、普通测试和审查不授予新付费调用。新的完整有限范围、全部历史/UNKNOWN/预留、预算分配、源码/二进制/工具及双轴审查必须成为具体计划，再取得明确批准；预算上限仍US$30。无新READY，race仍NOT_RUN，完整发布继续开放。

## Offline implementation progress

070b452已交付固定自身full-stream来源与严格消费者同步，生产协议/SDK/目录未改。公共Client/CLI先RED，六路线81受控场景及原生消费者最终通过，旧失败/中间失败和1012原件保留于[永久索引](../../../.evidence/provider-sites/final-acceptance-20261020/responses-native-source-fix/index.json)。17机制并入后07807a1公开API实算Regional414/MiniMax320，GLM245逐项仍等b384；不能把六受控路线或旧真实PASS当作新来源资格。控制器[259工件](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-controller-dev/index.json)与[138修复工件](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-review-fix-dev/index.json)保留完整公开RED/GREEN；最终50/50、204.167s、exit0。Standards两项、Spec一项经单一implementer集中修复，81e79b3最终[两轴复审](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/review/index.json)均0；GLM245/MiniMax320/Regional414未改。Status继续claimed；截至81e79b3离线复审结束，实际新付费授权4/grant4均无，该阶段READY/race NOT_RUN、release.pass=false。后续授权/运行、实际冻结及根检查状态见[准备当前状态（CURRENT_STATE）](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/index.json)与[计划](../../../.evidence/provider-sites/final-acceptance-20261020/recovery4-preparation/plan.md)；唯一当前状态仍见[规格](../spec.md)。

## Recovery4 actual stop and budget correction

恢复4实际08a3d10完成六地区formal及三M3自身完整PASS，前17进程成功；第18项OpenAI Decisions在finalbudget前发送估算失败，后六项按约未启动。原报告9Calls/9 recorderHTTP与USD guard0/0并存，不改写计数或退款。停止2032原件和closed4、跨轮native591/584、claimed980/1040、整笔28.80USD与UNKNOWN4全部保留。

[18分类预算隔离](18-classifier-budget-operation-isolation.md)离线最小修复进入三类严格来源，既有GLM245/MiniMax320/Regional414成功不能作为修后当前资格。新预算E2E加入后Regional416；没有改变任何来源选择器。八generic的保留须额外绑定原件与逐份final-wire/ledger差异证明。上述只记录本票历史和依赖，唯一当前状态见[规格](../spec.md#当前状态)；本票保持claimed，新付费机制/授权与最终门禁另行完成，READY/race未运行、release.pass=false。
