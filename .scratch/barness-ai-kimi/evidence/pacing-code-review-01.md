# Pacing初轮审查（历史，发现后已修复）

冻结目标`fafd72e1f2b0f74e481058b27e4baa912f7adcde`，比较点`1889322b136f9daa8e34ba94c1aa8a9832b964f7`。

## Standards

固定范围：`1889322b136f9daa8e34ba94c1aa8a9832b964f7...fafd72e1f2b0f74e481058b27e4baa912f7adcde`；提交`d27e19c`、`8600988`、`fafd72e`。只读审查，未运行测试、读取凭据或访问厂商。

**硬违规：1。**

- **[P2] 新授权间隔存在重复字段及大小写别名绕过。** `ai/internal/testkit/supportmatrix/regional_grant.go:84–97`先执行`json.Unmarshal(raw, &value)`，再仅检查`fields["callIntervalMs"]`。Go结构体解码不区分字段大小写，map又会覆盖重复键。因此`"callIntervalMs":35000,"CallIntervalMs":0`会接受实际间隔0；单独`"CallIntervalMs":null`也会被当成省略0，绕过“optional integer, never null”的新增边界。原生执行和消费者均直接解码grant，后续scope/receipt只能绑定已被覆盖的值，无法发现矛盾声明。违反**AGENTS.md原则3**“数据必须在边界处完成校验和独立转换”和**原则4**“任何外部输入均视为不可信”。应在typed投影前，对本次新增的`callIntervalMs`拒绝大小写别名、重复语义键及null，保留显式整数范围检查。其他grant字段既有宽松行为早于比较点，不计为本diff新增发现；无需借此重构整个授权模型。

**Smell判断：0。** 未见需要报告的Fowler启发式问题；原生grant、生命周期和独立消费者各自承担真实边界职责，没有引入通用调度或公共SDK扩展。

其余检查未见具体违规：等待位于call/token预算与120秒调用窗口之前；timer按返回停止，Kimi信号context通过defer释放；process/grant截止及一次性授权保留；间隔进入既有canonical scope散列，并由receipt及真实Observer时间核对。改动文件均小于500行。失败清单及公开原生进程红例/后续绿例原件分开保留，未将普通通过宣称为真实资格。

结论：**1项硬违规，0项smell判断；需修复输入歧义后定向复核。**


## Spec

Spec review: FAIL (2 findings)

Target `fafd72e1f2b0f74e481058b27e4baa912f7adcde`; fixed point `1889322b136f9daa8e34ba94c1aa8a9832b964f7`. Reviewed `git diff 1889322...fafd72e` and commits `d27e19c`, `8600988`, `fafd72e`. Read-only source review; no tests or vendor calls executed.

1. **[P2] Ambiguous interval aliases can disable the approved pause.** `ai/internal/testkit/supportmatrix/regional_grant.go:85–95` decodes the struct case-insensitively, then checks null using only the exact map key. `"callIntervalMs":35000,"CallIntervalMs":0` therefore selects zero, and `"CallIntervalMs":null` silently becomes omitted/default zero. R12 requires “callIntervalMs整数，范围0..60000、默认0”; design §10.4 specifically requires “非整数” rejection. Reject aliases/duplicates before decoding, or decode the interval only from an unambiguous exact field. The Standards reviewer independently identified this issue; it remains present in this immutable target.

2. **[P2] Impossible observation chronology can satisfy the pacing proof.** `ai/internal/testkit/supportmatrix/regional_pacing.go:26–34` checks only CallStarted/CallFinished times. It ignores AttemptStarted/AttemptFinished times, and the existing `verifyRegionalCase` also checks no timestamps. For example: a call starts at T, its AttemptFinished is T+100ms, and the next call starts at T+121ms. Falsifying the first CallFinished to T+50ms manufactures a 71ms pause for a 40ms grant despite the contradictory terminal occurring before its recorded attempt finished. R12 says “消费者用实际观察时间核对，不能仅自报等待”; §10.4 requires “时序/间隔篡改拒绝”. Validate nonzero, ordered timestamps for the complete four-event lifecycle before trusting its terminal as the next gap anchor, and add a resealed terminal-before-attempt refusal case.

The remaining reviewed amendment matches the requested scope: first/every-call wait precedes reservations and the 120-second context; grant/process cancellation and expiry bound waits; canceled initial waits retain consumed authorization with zero HTTP/budget; positive configuration participates in canonical scope, report and receipt; defaults remain zero; no retries, SDK payload changes, or source-hash grandfathering were introduced.


首轮：Standards1项，Spec2项（其中授权解析同一项）；均P2，后续复核见pacing-code-review-02.md。
