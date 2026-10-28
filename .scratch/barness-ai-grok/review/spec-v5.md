固定增量：`36eaca012d45062242a67e5f48d01ef07bfe5c5e...56301103ae1e35a9e0354e8c1b5574a7aa32fda1`。发现5项，均P2。

- **流事件证据可缺失全部内容块。** R11要求完整领域生命周期；设计第235行要求“start/block/end/terminal生命周期”，wire合同第16行要求拒绝缺失/篡改的triplet。`ai/internal/testkit/supportmatrix/xai_events.go:75–78` 只比较opened/closed数量，未要求覆盖Result所有块；仅保留匹配的start+done即可通过。delta也只检查字符串类型（第52–56行），改写/丢失增量而保留end仍通过，无法独立证明正文流正确。

- **Result内容未完整对照原始输出。** R08要求“可见推理与原生条目分开”并保留顺序。`xai_output.go:72–99` 只比较Thinking.Signature，不核对Thinking.Thinking；正文被合并成一个字符串，跨种类块顺序丢失；工具只核对RawArguments，没有核对用于显示/历史的Arguments。原生签名保持而可见推理被清空或替换、内容块被重排，仍可取得资格；流end只是再次比较同一Result。

- **工具往返的第二轮约束与结果内容未兑现。** R12要求“none却出现函数调用…protocol失败”；wire合同第9行要求完整工具配对。`xai_history.go:128–173` 只校验第一轮choice/count，未独立检查results请求的none与零返回调用；第232–240行仅要求结果文本非空。第二轮再次返回工具，或两个并行调用的既定`synthetic-host-result-a/b`被互换且ID不变，都不会被拒绝。

- **不支持的收费语义仍可按默认价取得资格。** R14限定模型token/Host函数；设计第308行要求priority/fast“成本标partial并protocol失败”。`xai_usage.go:89–130` 未检查终态service_tier及正cache_write_tokens/server_side_tool_usage/额外收费类别。保持现有数字和Result，加入priority或额外用量后，消费者仍计算全球默认价并接受resolved。

- **合法的空密文被扩大为正式资格拒绝条件。** R08要求合法原生JSON保留，既有公开合同明确接受encrypted_content:null。`xai_output.go:33–37` 却要求所有24次响应的每个reasoning item都有非空密文；因此非续接场景的合法null/缺失值会失败。应保留专门encrypted-continuation场景的非空密文证明，避免扩大到全部响应。

本次只读静态审查，未执行测试、构建或厂商请求。已接受的未发布Builtin接线顺序无问题；新正式live、人工race/压力及旧路线准备仍是分别待完成的事实，不作为本增量缺陷计数。
