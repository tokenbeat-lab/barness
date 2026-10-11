# 07：定位当前发布压力流完整性失败

Status: ready-for-agent
Type: bug
Updated: 2026-10-11

## Actual failure

当前受测源码 `3b74fb3` 的完整离线 gate 有一个实际 FAIL：
`TestPolicyPressure/cloud-interactive/design-load`。
call-04 得到不合法 Responses JSON；call-11 的累积工具参数超过 512 KiB。
设计输入仍为 32 并发、24 条 64K token 文本、8 条 256 KiB 工具参数。
全部资源最终释放；不能将该场景写成 PASS。

两条客户端请求按相同合成 requestId 对应的 fixture handler 都记录了
`writeBufferWaits`（2、3）。call-04 随后 broken pipe；call-11 handler 写完。
这只是与 ENOBUFS 等待的关联，尚未确认损坏来自 socket 续写、SSE 分帧或解析。
旧工单 18 的 unexpected EOF 已有不同直接因果证据；不能据此宣称本轮原因相同。

完整原包与审计随 release-evidence 归档，case 为
`E08-policy-pressure-cloud-interactive-design-load`；当前发布仍被阻断。

## Failure scenarios first

- 受控 Provider 的累积字节与 Client 实际读到的字节不同，需关联具体连接和写入边界。
- 字节相同而 SSE 帧不同，需保留最早发生偏离的序号与原始边界。
- 帧相同而工具累积超限，需验证分档采样关闭时与当前空 profile 的同输入行为。
- 到达顺序误当调用身份，或只保存成功复现覆盖第一次失败。
- 减并发、加大限额、重试整个场景或改成合成 PASS 掩盖问题。

## Reproduction and acceptance

```sh
BARNESS_AI_PRESSURE=1 go test ./ai/e2e \
  -run '^TestPolicyPressure$/^cloud-interactive$/^design-load$' -count=1
BARNESS_AI_PRESSURE=1 BARNESS_AI_PRESSURE_LOOPS=10 go test ./ai/e2e \
  -run '^TestPressureLoopback$' -count=1
```

每次使用新 evidence 目录。先形成重复且能定位首个偏离字节/帧的反馈，再写回归，
再修改确认有问题的边界。保持原负载、预算、生产认证/授权和比较规则；保留完整
失败历史、审计、资源释放和每个 handler 的结束记录。修复后重新跑完整 gate，
一次成功的专项复现不能替代本次失败发布。

2026-10-11 追加实际复现：同源码、独立受控 Provider 的十轮
`TestPressureLoopback` 共 10 PASS、0 FAIL，执行 71.455 秒，审计零发现。
本轮没有重现自然失败，不能认定问题已修复或改写失败 gate。
普通 gate 的单次 FAIL 与完整 race 的同负载 PASS 均保留。

第三轮当前源码 `dba212b` 的完整普通离线命令已 PASS（6,125 case，272.638 秒），
同一压力场景保持原设计负载并 PASS；完整 race 命令及审计也已 PASS（970.760 秒）。
此轮只调整 release CLI 的整体 race 时间窗口，没有改流读取、解析或受控 Provider；
所以本工单仍是未确认原因的缺陷，不能因重跑通过而关闭。
