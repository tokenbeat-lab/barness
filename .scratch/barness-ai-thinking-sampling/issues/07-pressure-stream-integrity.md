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

## Implementation investigation (2026-10-13)

- Public seam: existing authenticated/authorized ai.Client stream against the controlled
  loopback Provider, retaining the original policy, load, comparisons and deadlines.
- Failure inventory before changes: omitted/duplicated/corrupted body bytes; correct bytes
  but differing SSE frame boundaries; exact frames but incorrect tool accumulation;
  request identity confused with fixture queue order; loss of the first failed round;
  instrumentation retaining whole response copies or masking cancellation/resource release.
- First unchanged-source reproduction: design-load FAIL, call-15 invalid Responses JSON;
  same handler completed all 65 chunks with one writeBufferWait. Audit: zero findings.
  Evidence: `.evidence/pressure-integrity/baseline-design/20261013T072521.798237000Z`.
  This establishes recurrence, not the cause.

### Confirmed boundary defect

Two independent synthetic socket reproductions returned identical first differences:
fully enqueued/unreported ENOBUFS at byte 492 (frame 2, frame offset 16), and
partially enqueued/unreported ENOBUFS at byte 354 (frame 1, frame offset 120).
The former even returned public success with 2,846 bytes read from a 2,604-byte
script. Both correlate the same requestId, connectionId and socket write digest;
no AI request/Attempt retry occurred. Regression was red before the fix.

Apple's TCP path calls sbappendstream before tcp_output, whose ENOBUFS can
propagate through sosend; fp_writev suppresses only ERESTART/EINTR/EWOULDBLOCK
after partial progress, so ENOBUFS need not expose consumed bytes to Go:
- https://github.com/apple-oss-distributions/xnu/blob/main/bsd/netinet/tcp_usrreq.c
- https://github.com/apple-oss-distributions/xnu/blob/main/bsd/kern/uipc_socket.c
- https://github.com/apple-oss-distributions/xnu/blob/main/bsd/kern/sys_generic.c

Removed the unsafe ENOBUFS retry entirely. The controlled Provider now preserves
native socket errors and records failed write boundaries. Temporary exhaustion
is an explicit fixture/transport failure, not a recovery or fabricated PASS.
Product parsing, authentication/authorization, policies, budgets, load and
comparison rules are unchanged. Revert this commit to roll back the diagnostics
and boundary change together; no compatibility path or new dependency exists.

The original natural failure still lacks a captured divergent byte/frame.
Pre-fix traced 10 + 10 + 50 rounds passed without ENOBUFS; they do not prove a
fix or establish the cause of call-04/call-11. The first unchanged-source FAIL
and all later runs remain separate. One baseline loop run was interrupted by
an evidence-directory relocation during execution; its files are preserved as
incomplete and never counted as a valid test/audit run.
