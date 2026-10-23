# 有界重试执行范围审查

产品Go/native源码及二进制保持原冻结版本；审查对象为范围05、授权04、fresh8grants和单进程执行器，比较基线c274ee7。源码身份和旧Chat资格不改写。

首轮Standards1项P2、Spec2项P2：结构化错误字段遮蔽及失败原件完整性缺失；第二轮Spec另指出终态协议/取消/超时被缓冲区过载遮蔽。全部先用公开CLI --assess保留红例，再修复；初次完整性实现错误假设manifest包含NOT_RUN，导致真实过载正例拒绝，该非通过attempt也保留。最终10项PASS只代表操作评估，不当成主E2E新case ID。完整原件位于`.evidence/barness-ai-kimi/preparation/05/controls/`；各审查版本原件在`preparation/05/reviews/`。

# Standards targeted follow-up 03

PASS — 本次增量无硬违规；无新增 Fowler smell 判断。前两份报告保留。

目标：`/Users/cyber/RestoX/barness/.evidence/barness-ai-kimi/paid-execution/04/execute_one.py`，SHA-256 `bb44925e84191b9112962b78106b5d031d76a610482b00ca5536c7ada5fcc573`。比较保留的 `preparation/05/controls/terminal-red-executor.py.txt` 与当前文件，仅审查本次终态、SSE 与时序增量。

- 第 65–73 行逐个已执行 case 要求完整四事件、非递减时间及从授权开始/前次结束计算的 grant 最小间隔，同时保留资源全部归零检查。符合 AGENTS.md 原则 5「主动考虑…取消、重试…和资源释放」。
- 第 76–88 行先要求实际 `call_finished` 为 error，错误码仅为 `upstream_error`/`rate_limited`，HTTP 200 对应 stream，429/503 对应 request；任何 SSE data JSON 解析失败或非对象立即拒绝。取消、超时、协议失败不会因另有 overload 文本获得重试资格。符合 AGENTS.md 原则 3「数据必须在边界处完成校验」及原则 4「任何外部输入均视为不可信」。
- 读取保留原件：`classify-red-terminal/result.json` 的四个误判确实存在；`classify-green-terminal/result.json` 十项全部通过、每项进程退出 0，记录的执行器哈希与当前字节相同。符合 verification.md §2「保留实际命令、退出和原件」及 §8「成功重跑不覆盖旧失败」。这仅确认既有 keyless 控制证据，不代表冻结普通验收或 live 已完成。

本审查未运行测试、执行器或厂商请求，未读凭据，未修改仓库。


## Spec

Spec targeted operational re-review: PASS (0 remaining findings)

Reviewed `execute_one.py` SHA256 `bb44925e84191b9112962b78106b5d031d76a610482b00ca5536c7ada5fcc573` and unchanged proposal SHA256 `d25ebbf173c930cdc43c85ac858e278a4d31f665a45c1d4f2f14bb90caddcb81`. Prior reports remain unchanged. Read-only review; no tests, credentials or vendor calls executed by this reviewer.

The remaining terminal-classification finding is resolved at executor lines65–99. Eligibility now requires the failed public `call_finished` observation to have stopReason `error`, code `upstream_error` or `rate_limited`, and the matching stream/request phase. Protocol, canceled and timeout terminals therefore veto continuation despite a later buffered overload. Malformed/nonobject SSE data also refuses. Complete four-event ordering and initial/inter-call timing are checked across all executed cases before that decision. This implements scope05's explicit nontransient-stop rule.

The earlier fixes remain present: every populated error type/code and every error object must be recognized; quota/unknown/conflicting records refuse. Complete originals are independently checked through required roots, exact sealed inventory and hashes, manifest executed-case membership/status/artifacts and grant/report/receipt binding. Native failure and candidate-consumer refusal remain failures.

The retained `controls/classify-green-terminal/result.json` records 10/10 passing keyless assessments using this exact executor hash, including the real overload positive control and all submitted quota, unknown, incomplete, corruption, protocol, cancel, timeout and malformed-before-overload counterexamples. This review read those originals without rerunning them.

The financial guard includes the pinned prior US$0.964002 reservation before new actual reservations plus each next US$1.50 grant; the global ceiling remains US$4.50. The reviewed scope remains international Responses/Messages only, fresh whole-route grants, 60-second native intervals, 120/240/480-second cooldowns, at most four rounds per route, fixed deadline and no SDK/active-stream replay. Product/native source and binary remain unchanged, preserving the separate current-source Chat9/9 evidence. Frozen ordinary acceptance and actual execution remain separate pending stages.


Standards0项/Spec0项；最终执行器SHA256 bb44925e84191b9112962b78106b5d031d76a610482b00ca5536c7ada5fcc573。新完整普通验收与真实付费结果仍独立判定。
