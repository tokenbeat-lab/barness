# 两轴 code-review：工单 07

基点：4f90386fa4c879dddfa3013cce778c44de873acc。
两个独立审查代理读取 git diff <base>...HEAD，不编辑代码。
完整变更复核于 25aebde，追加指针接收者编码修复后再复核；最终记录见下文。

## Standards

初次审查无发现。第二轮发现一项 P2：对任意 struct 的反射尺寸估算没有遵循标准
JSON 字段选择，匿名字段冲突可能忽略大字段，自定义 TextMarshaler 可能输出小值；
同时自定义 questions marshaler 的源 map 长度不等于编码后的问题数量。
这会误拒有效的小请求，违反 AGENTS.md 原则 3 的边界校验与稳定行为。

通过公共 E2E 先复现四类误拒（encoding-semantics-red.log），再将未知 struct、自定义
JSON/Text 编码交给标准编码器和最终精确预算；普通 JSON 值仍检查已知尺寸。
字符串 map key 的特殊标准库语义另由 string-key-red.log 先复现并修复。
第三轮另发现一项 P2：可取地址的切片/数组元素可能采用指针接收者 JSON/Text 编码，
只检查值方法集会误拒小编码。pointer-encoding-red.log 先复现三类场景，修复后
encoding-green.log 共 14 项通过。最终复核报告在本文件后续记录。

## Spec

最终代理报告：No remaining actionable Spec findings in 4f90386...25aebde. No scope creep identified.
The implementation and existing/new E2E coverage satisfy issue 07’s lifecycle behavior: cancellation and
deadlines, status/connection retries, pinned snapshots, callback execution and error identity, response
byte budgets, resource release and tenant isolation, observation attribution, and PhaseResponse matching.

此前两项 P2 均解决：

1. 超限回调值在拒绝前已经编码和独立转换。已将已知状态/问题尺寸检查移至序列化前，
   最终 MaxRequestBytes 检查移至独立问题解码前。allocation-red/green.log 先复现后修复。
2. 数字和空字符串数组未计非零最低尺寸，可造成 54–73 MiB 库内分配才拒绝。
   已计标量、容器标点和字符串 key 的编码下界。encoding-red/green.log 先复现后修复。

最终复核还确认自定义编码、匿名字段冲突和字符串 key 的标准库语义已得到回归覆盖。
宿主自定义编码自身的执行仍由宿主负责，结果在转换/发送前检查最终字节预算；ADR-0002
记录边界。代理复核时，全量测试和最终证据尚在运行；其结果由 report.json 独立记录。
