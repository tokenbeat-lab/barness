# ZAI Chat 原生来源的普通实施证据

工单：[12](../issues/12-zai-chat-native-source.md)。功能当前状态只在 [spec](../spec.md#当前状态)；本记录不把普通重放变成真实 candidate 资格。没有读取 Key、新增厂商调用、执行 race 或修改已关闭轮次原件。

## 真实故障与合同

`candidate-final/candidate-zai-chat/20261020T104434.930813000Z` 实际18 Calls /18 HTTP、17 PASS /1 FAIL。GLM-5.3 tool-call 原始 SSE 只有工具、空正文与 role，没有 reasoning_content / reasoning / reasoning_text；六帧完整、HTTP200、tool_calls终态、DONE。用量 reasoning0，资源全0。上次自身 PASS 的同一工具请求逐对象完全相同、端点相同，原响应有14字符 reasoning_content。当前进程 full-stream 返回70字符、GLM-5.2 tool-call 返回97字符；公共 Client 对当前四个原件保持精确请求与原生 Hash，不造缺失字段。171个原文件已固定 Hash 并再次确认未变。

R06并列要求工具往返和原生 reasoning 续接，R10要求精确来源，R18要求逐身份自身证明，没有要求原生必须与工具同回合返回。原实现把两项证据都固定到工具响应。根任务明确选择最小修正：只对 `zai-chat` 在任何请求前固定 `full-stream` 为 native-history 来源；其它五路线仍固定 tool-call。报告记录精确 NativeSourceCases，消费者独立推导预定集合，不能搜索另一个成功回合或 fallback。

native-history 发送原 full-stream 用户/助手前缀，再保留原工具用户/助手、Host结果与REGIONAL42末轮。源消息没有 native 时仍 FAIL；源为另一型号、账户或租户，domain请求不是原来源的确切前缀，actual wire没有相应native或可见内容，均拒绝。完整工具链、相同Tenant / Account / Provider / API / Model、downgrades0与独立原始SSE消费保留。每型号九场景、每路由18/9调用、24Calls/24HTTP、零重试及原字节/token上限不变；新增前缀按实际编码重新计量预留。

## 实际普通记录

全部原始命令、UTC、数字exit、完整源码前后Hash、manifest和审计在 [索引](/tmp/barness-regional/zai-chat-native-dev/index.json)，原件只读诊断在 [census](/tmp/barness-regional/zai-chat-native-dev/diagnosis-summary.json)。这些工件位于实现worktree之外，根任务最终索引应继续留存。

| 轮次 | exit | 实际事实 |
| --- | ---: | --- |
| red | 1 | 新研究 fixture 辅助读取误把content数组当字符串；构建好的公共CLI尚未发送受控请求，保留失败 |
| red-native | 1 | 公共CLI精确重现17 PASS /1 native=false FAIL，18受控HTTP；与真实故障相同 |
| green | 1 | 修正后producer18 PASS、四原件公共Client PASS；旧raw-thinking篡改定位仍指无推理工具帧，断言正确拒绝无效篡改 |
| green-boundaries | 0 | 父34 PASS、producer18 PASS；新增源缺失子进程含两个额外合成marker失败，后续收紧 |
| affected-final | 1 | 827 Go test PASS /3 FAIL；父781 native PASS /3 FAIL。源缺失合成脚本实际仍包含reasoning，新增反例及父/提案继承失败；其余五路线、21个实际Frozen PiDiff与普通合同通过 |
| source-boundaries-final / pipeline-final | 1 | 新wire反例改变字节长度，先被旧输入预留核对拒绝；随后等长替换未保留转义宽度，测试显式拒绝该无效反例。全部失败保持原件 |
| pipeline-complete | 0 | 153 Go test PASS；父152 native PASS、六个同commit/binary受控producer共81 PASS、受控promotion PASS；14审计均0。缺源负例子进程恰为16 PASS /2 native-history FAIL，由父用例正确认证拒绝 |
| absent-positive-final | 0 | 只补源缺失反例的显式正对照：另一工具回合实际含非空native；父2 PASS，正producer18 PASS，负producer16 PASS /2 native FAIL，4审计均0 |
| consumers | 0 | supportmatrix、release与releasegate三个包206测试PASS |
| vet / vet-live | 0 | 双vet通过 |

最终pipeline受测完整ai/go源码为 `23f3804d55a9f6bbd789f93c9ee0d10dd11e340400d1814c9d64d356a0939bb1`。最后只增强源缺失反例中“其它工具消息确实native-positive”的测试断言，定向普通对应完整ai/go源码 `d56fd38ddcb33992c74a52de14a615c34b60cb1e6ceeec1e5bb179e08430a92a`；运行/harness/consumer字节未改。每个回执的source-before/source-after均相等。没有把预期负例子进程的FAIL当成发布失败，也没有将早期整轮exit1写成整轮PASS。

有限trace仅向R06/R10/R18、RP07/RP11/RP18追加实际11个新case的对应义务；旧所有字段、case、P0、live及差分集合未删改。[严格只增差异](/tmp/barness-regional/zai-chat-native-dev/trace-additions.json) 独立复现旧normalized pin `c4daf796…`，新增literal为 `d79b29621b04b5fa46b7f907cecfa079b051dbef2ff1937d4f84b85a14f03df2`。Pi冻结、SDK/dependency、差分ledger和候选官方source/pin均未改。

本修正改变执行来源，先前成功和失败保留各自事实，不将成功原件刷新为当前资格；六条路线的新真实candidate、晋升后formal与最终门禁均由根任务在新的有限授权后执行。

根任务已逐字节留存全部12,687份原件于[永久目录](../../../.evidence/provider-sites/final-acceptance-20261020/zai-chat-native-dev/index.json)，无符号链接；[留存回执](../../../.evidence/provider-sites/final-acceptance-20261020/zai-chat-native-retention.json)固定原/tmp前缀与永久前缀映射、全部文件集合Hash及相等性，不重写原索引。[独立Standards / Spec复审](../../../.evidence/provider-sites/final-acceptance-20261020/round2-preparation/review/chat-source/index.json)均0 findings，范围含永久父报告路径修复。
