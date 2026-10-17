# #07：自然 socket 反例的有界重现

这是诊断装置，不是真实支持或发布 PASS。只使用受控 loopback，所有子进程均白名单环境。
`source-base.json` 固定 `cc78cfd` 的完整 600 个 Go 源文件；源码不匹配时拒绝运行。
原始 32 并发、24 条 64K-token 文本、8 条 256 KiB 工具参数、生产策略/授权/比较保持。
辅助进程先建立全部连接，再通过普通 TCP 写入形成有界竞争缓冲，三秒写 deadline、
有限 refill/drain；不注入 errno、帧或正文。不要与完整门禁或其他负载实验并行运行。

从仓库根目录执行；每次输出新的 timestamp 目录，不覆盖旧回合：

```sh
# 诊断 RED：只在临时 overlay 恢复历史不安全续写及原默认发送缓冲。
python3 .scratch/barness-ai-release-verification/native-pressure/reproduce.py 1 condition 50 256 1 1
# 同一条件下的当前安全边界：实际 ENOBUFS 仍应是显式失败，不能伪装 PASS。
python3 .scratch/barness-ai-release-verification/native-pressure/reproduce.py 1 condition 50 256 1 1 --fixed
```

收集脚本退出 0 不是测试通过；检查 receipt.json 的实际 `exitCode`、manifest、audit、
每请求 provider-outcomes/stream-integrity/native-call-terminals 和 probe.json。
只有观测到原生失败与关联首差才算因果 RED；未出现故障的一次成功不是修复证明。
运行主实验时仅保留失败 wire 的首个 payload，不复制整条响应，逐字节比较仍覆盖所有流。
`--fixed` 的预期是无字节偏离、无 AI Attempt 重放、资源归零；外部缓冲耗尽仍记真实 FAIL。

两个独立、顺序执行的自然 RED（runtimeSourceUnchanged=true、audit=0）：

- `.evidence/pressure-integrity/native-20261017T022612.488595Z`：call-03，connection 2。
- `.evidence/pressure-integrity/native-20261017T022622.224455Z`：call-09/11，connection 4/30。

共同首差：body 1,318,335、SSE frame 4098 offset 262,214；应结束 JSON 的字节变为重复 `c`。
关联原生 seq194、wire offset 1,060,329、783,784-byte write、ENOBUFS/0-reported、同 Hash 续写。
当前安全对照 `native-20261017T022910.882286Z`：32 handlers/traces、22 个显式 transport
失败、10 个成功、每调用一个 Attempt、零字节偏离、资源全部归零、audit=0。
这是故意资源耗尽诊断的实际 FAIL，不是生产门禁 PASS。

上述早期 prototype receipt 的 captureScope/proxy 字符串沿用了旧装置标签：实际源码保留
完整 streamed-byte 比较，未走外部代理；最终脚本已更正元数据。原始 receipt 不改写。
最早重叠实验、未复现回合、冷编译/设置超时及编译诊断仍保留，不计独立有效验收。
历史 call-04/call-11 的原始字节未被反向认证；证据证明可独立重现的同类历史边界机制。

既有公开回归保持：`TestPressureFixtureWriteBackpressure`、`TestPressureToolWriteIntegrity`、
`TestPressureTransportDiagnostics`；恢复旧续写会 RED，当前实现须 GREEN。
最终仍需原压力设计/十轮与完整 normal/race，不能用本装置代替整版验收。
