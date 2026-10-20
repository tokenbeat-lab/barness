# 21 中国站 MiniMax Responses 的固定原生推理来源

Status: claimed
Type: task

## Context

阻塞[07最终验收](07-final-acceptance.md)。唯一功能当前状态见[规格](../spec.md#当前状态)，实际运行/冻结/授权状态见[恢复5准备索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-preparation/index.json)。本票仅离线诊断及必要的窄harness/原生消费者修复，不发行新授权/nonce/claim/grant，不读真实Key或调用厂商，不运行race。

## Actual stopped failure

恢复5实际8c83426在第八完整进程`formal-minimax-cn-responses`停止：前七完整PASS；本进程九场景8PASS/1FAIL，`MiniMax-M3/native-history`报告`native=false`，后11进程NOT_RUN。本轮新增native111Calls/108 recorderHTTP，累计native702/692、claimed1160/1232；整笔US$32.05（已批总上限US$36），UNKNOWN父记录5、退款0。全部数字与失败原件只读保留，不把guard前recorderHTTP改称已发送物理HTTP。

[实际停止索引](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-stopped/index.json)，SHA256 `5e8014e7405cfa86da21637b96e47cb294c4ff8e766e79440fe88ff56eef4f73`。原native目录为`.evidence/provider-sites/final-acceptance-20261020/recovery5/formal-minimax-cn-responses/20261020T225703.540330000Z`。失败不授权真实完整过程重跑或追加新付费。

## Offline diagnosis and boundary

先以两份原始SSE在公共Client/CLI回放完整工具/Host-result续接，两个最终请求body与URL逐字符合原件，复现同一`native=false`且零downgrade。再提出可证伪假设，独立比对raw/events/Result、上轮自身成功请求与本轮固定文本源；不得先假定parser有bug。

原工具响应若没有原生推理，不能合成、借同品牌/协议来源或降断言。任何固定文本来源方案必须在请求前按精确服务/型号确定，保留原user/assistant前缀和独立工具链，并由原生消费者严格验证；固定源缺native仍失败，不能事后择优。

来源变化对已通过M3/地区formal与八generic严格保留政策的影响必须显式列清，不缩小selector、不刷新旧报告。新增付费范围/预算、固定authority/controller与最终完整门禁另行准备和审批，不在本票离线结果内。

## Evidence

[永久诊断开发目录](../../../.evidence/provider-sites/final-acceptance-20261020/recovery5-cn-responses-diagnosis-dev/)保留原件副本与SHA、最小公共RED、先验假设、单变量probe、所有中间失败和最终结论；不改变实际停止原件。最终离线结论/受影响普通来源将在该目录索引关联，不能替代真实当前资格。


## Offline implementation

公共精确原SSE两调用回放先RED：最终request body/URL逐字符合原件，`native=false`且所有downgrade为0。新工具原SSE为0reasoning事件/0item，terminal reasoning_tokens字段缺省，公共归一值0；旧自身同请求83字/21tokens与新自身full-stream150字/31tokens均由同一codec正确保留。因此本票修复固定验收来源，未发现生产parser缺陷。

新指定来源的公共harness/consumer亦先RED：full-stream有推理/工具无推理时旧流程失败；反向full-stream空/工具有推理时旧流程与consumer错误接受。实现仅让harness复用固定`RegionalNativeSourceCase`，按精确中国Responses/M3选择full-stream，并保留原文本前缀与完整工具链。空固定源仍失败；无重试、成功搜索或伪造Thinking。新增七个有意义的公共证据case只追加到有限trace及其独立approved pin，所有旧case保持。

六路线完整81受控场景、原生consumer、既有篡改/取消与七新边界、公共release contract/source CLI组合实际exit0、27.640s；两个vet exit0。随后Git fixture中response/item/tool ID、opaque原生值与时间替换为明确controlled数据，精确原始响应只留ignored诊断原件；现有audit CLI零发现，受影响CN流程与trace合同再次通过。中间fixture假设/Nullable编译/空fixture装配失败原样保留，不冒充GREEN；全部实际命令与最终源见永久索引。

M3执行集合仍325项但regional_contract一项变；Regional421中合同、harness及有限trace approved pin三项变；GLM245逐项相等。三个M3和六formal必须新当前资格，八generic不能越过恢复5严格来源策略。guard两文件、SDK/生产协议、目录价格与所有selector均未改，不新增authority/controller。票据保持claimed，新的实际资格/付费准备仍只见spec及恢复6当前索引。
