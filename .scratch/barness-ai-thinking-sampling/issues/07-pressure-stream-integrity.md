# 07：定位当前发布压力流完整性失败

Status: resolved
Type: bug
Updated: 2026-10-17

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

### Verification and remaining acceptance

- Socket/reader GREEN: 11 public E2E cases; targeted race also 11 PASS.
- Original design command and post-fix ten-round loop: PASS with unchanged load;
  every request-correlated scripted byte arrived, handlers ended, resources zero.
- First full gate: normal/race each 7,618 PASS + 1 FAIL, solely
  OD57-release-evaluation-current (a retained live test fixture was no longer
  source-current). Both full original bundles and successful audits are retained.
- Fixed only the E2E fixture setup: generate current evidence with the existing
  public CLI; fixture positive control remains NOT_RUN, all tampering must FAIL
  and cannot be masked by source drift. No gate comparison was relaxed.
- Final full gate, source 1291d10: actual exit 0, all ten gates PASS. Normal and
  race each 7,620 PASS (including 861 frozen Pi differential cases), zero FAIL;
  commands took 357.680s and 1,214.584s. Both vet modes and all 14 input audits
  PASS. Ten own-route smoke inputs retain their original applicability checks;
  Jev and Luna were each newly collected through the explicit local proxy with
  24 Calls/Attempts, zero retries, COMPLETE and zero audit findings.
- Two-axis final review: Standards 0 and Spec 0 remaining code findings. The
  initial cancellation-cleanup and fixture-outcome findings were fixed before
  the successful gate.

[Archive and exact verdicts](../pressure-stream-integrity-evidence/manifest.json),
[full report](../pressure-stream-integrity-evidence/release-report.md),
[two-axis review](../pressure-stream-integrity-evidence/review.json), and
[verification](../pressure-stream-integrity-evidence/verification.json) preserve
22 complete bundles / 30,619 case records, including 13 historical failing
investigation cases and both first-gate failures. Outer SHA and inner manifests,
audits, inventories and verdicts are independently checked; top-level diagnostic
filesystem paths alone are aliased, original inner evidence bytes are unchanged.
The interrupted loop remains explicitly INCOMPLETE / audit NOT_RUN.

Reproduce the archived checks without network:

```sh
python3 .scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/verify.py
python3 .scratch/barness-ai-thinking-sampling/pressure-stream-integrity-evidence/replay-red.py
```

The second command compiles a temporary Go overlay with the removed unsafe retry,
requires both original byte/frame counterexamples to FAIL, and checks runtime
source remains untouched. It writes a new evidence directory and binary/source
SHA; it is explicitly regression RED, never production or release evidence.
Use the two original reproduction commands above for current GREEN.

**Status remains ready-for-agent; issue is not closed.** This delivery fixes the
separately confirmed unsafe socket resume and completes current verification.
It does not establish the first divergent bytes of the historical natural
call-04/call-11 or the unchanged-source call-15. A future natural failure must be
examined via the retained per-request byte/window/connection/write traces; only
if bytes match should decoder-frame capture and then disabled-sampling versus
empty-profile accumulation be investigated. Automatic gate PASS does not revoke
that remaining causal-evidence requirement or the historical delivery blocker.

## 2026-10-17 续修验收与关闭

用户另行授权按本工单原要求继续定位、修复并执行完整门禁；真实服务保持 NOT_RUN。
此前未关闭说明作为当时事实保留。本次已补足重复的自然因果反馈和当前完整验收，
因此将原工单自身转换为 resolved，不依赖整合票代为撤销历史要求。

两个顺序独立、未注入 errno/正文的普通 TCP 竞争缓冲反例，分别在
`.evidence/pressure-integrity/native-20261017T022612.488595Z` 和
`native-20261017T022622.224455Z` 关联 call-03/09/11 与 connection 2/4/30：
相同首差 body1,318,335 / SSE frame4098 offset262,214，JSON结束字节被重复c覆盖；
同连接 seq194、wire offset1,060,329、783,784-byte write 返回 ENOBUFS/0，继以同Hash续写。
它证明原解析前不安全续写的自然机制；不反向认证历史call-04/call-11原始字节。

当前安全边界同条件对照实际FAIL：22显式错误、10成功，32handler/trace，各一次Attempt，
零字节偏离、资源0、audit0；没有将资源耗尽当生产PASS。原不安全续写公开RED重放仍失败，
当前隔离70个回归、原压力设计和十轮均通过。默认受控 TCP 夹具请求64KiB发送缓冲，
限制自身内核队列增长；保持原32并发、脚本、预算、认证/授权、比较和真实错误传播。

当前Go内容提交cc78cfd，完整执行提交f29703b：normal/race各9,098 PASS、0FAIL，
实际命令成功，耗时493.945s / 1,849.728s。两种vet、P0、冻结Pi差分和聚合审计通过，
17个输入包审计0发现；两阶段各861个Pi及33个UC14用例通过。源码600文件Hash不变。
Standards/Spec独立复核无剩余发现，认为本工单因果、回归及完整受控验收均满足。

原始发布报告仍pass=false，保留启动时本工单未关闭的snapshot/Hash；关闭发生在本次
验收结束之后，绝不改写或追认旧报告PASS。真实11路线和4效果包仍NOT_RUN，
Calls/Attempts均0；整体发布及整合票14保持BLOCKED/needs-info。状态转换前后原文、
Hash、时间、所依据原报告见`.evidence/release-followup/pressure-tracker-transition.json`。
详细结果与可重复封存校验见
[续修验收](../../barness-ai-release-verification/verification.md)和
[两轴复核](../../barness-ai-release-verification/code-review.md)。
