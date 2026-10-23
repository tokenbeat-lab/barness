# 缓存默认值与成本边界双轴审查

比较固定点 `2d55fcfa8957f430cb2b9950942429a58fc14a19`，冻结实现 `3a722ba3e89808dcbf081758dd33dfeeed813e67`；交付合并 `02aea23dc6bae2e172fa98e6215b4627454df196` 仅追加证据说明。两轴独立、只读执行。

## Standards — PASS

冻结范围：`git diff 2d55fcfa8957f430cb2b9950942429a58fc14a19...3a722ba3e89808dcbf081758dd33dfeeed813e67`；包含契约提交`a1bf7e1`、实现/公共E2E提交`178d8c3`及合并`3a722ba`。

**硬违规：0；需报告的 smell 判断：0。** 本轮采用已获用户确认、修订后的 ADR-0052，不将旧的读取 presence 或“仅TTL可bounded”限制当作当前规范。

- `kimi_usage.go`继续在协议边界执行精确数值、包含关系、总量及溢出校验。读取默认0仅作用于Chat/Responses；Messages与其他Provider保持原路径，符合AGENTS.md原则3、4及ADR-0052修订范围。
- `kimi_cost_bounds.go`最多评估三个独立值场景；未知分区仅用于估价，不写回`CacheWrite`/`CacheWrite1h`观测；请求TTL不参与实际档位证明。上下界逐项检查有限性，失败沿现有错误/资源释放路径结算；没有新增共享可变状态、重试或网络入口，符合原则3、5、9。
- 改动Go文件均不足500行；未新增依赖、推测性公共抽象或兼容层，符合原则1、2、7、10。合成与真实来源有明确标注，成本说明同步到ADR、Glossary、Host与原生验收文档，符合原则8。
- 只读核对了保留的公共Client/原生CLI进程红绿回执及`/tmp/barness-kimi-cost-bounds/tdd-attempts.json`九次尝试说明：真实SSE与恢复原runtime的干净红例先于最终实现，前序失败原件均存在，冻结来源后续388及10个实际用例分别PASS；未用它们宣称完整普通验收。真实SSE夹具与原失败capture的responseBody逐字相等，exchange和manifest哈希匹配origin记录；离线回放不追认旧真实资格，符合测试SOP §§2、3、4、8。

本次审查未运行测试、读取凭据、请求厂商或修改仓库。完整冻结普通验收及后续独立授权阶段仍由根任务执行；本报告仅给出本diff的Standards结论。

## Spec: PASS

Reviewed `git diff 2d55fcfa8957f430cb2b9950942429a58fc14a19...3a722ba3e89808dcbf081758dd33dfeeed813e67` against the revised R09/K10/K11, design §§8.2–8.3, issue08, ADR-0052 and the explicit contract pointer. Read-only; no tests or vendor requests executed.

No missing requirement, scope expansion or implementation defect found in this range.

- **Optional read:** Chat/Responses missing/null read has semantic zero for completeness and pricing, after Chat’s existing nested/legacy precedence. Explicit invalid selected counts still fail. Messages keeps its required read counter; other Provider paths are untouched.
- **Missing-write bounds:** With known P/O/R, the scenarios preserve P while testing W=0 and W=P−R at every allowed write tier. Actual Responses TTL narrows those tiers; request TTL is excluded. Comparing costs handles reversed/equal Host rates, and nonfinite arithmetic fails. Unknown public write/1h counts remain unobserved, UsageReporting stays partial, and Usage.Cost uses Minimum. Missing principal facts retain partial/unreported costs; reported total contradictions still fail. Existing Kimi tier rejection remains.
- **Snapshots and qualification:** Accepted usage/cost values still publish together to Result, Attempt and Observer. Controlled qualification now exercises default reads and missing-write bounds; its native consumer already independently replays cost states. Assertion wording now describes finite reliable bounds. No old live result is promoted by replay.
- **Real SSE provenance:** Independently read the primary retained capture. Fixture bytes exactly equal the original responseBody; manifest, exchange, response-body and fixture SHA-256 values all match the new origin record. Original bytes were not rewritten.

Added public E2E scenarios were inspected but not executed by this reviewer. Scoped/full ordinary results, new source qualification, paid authorization and release readiness remain separate evidence.

Standards：0项硬违规、0项smell；Spec：0项发现。完整冻结普通验收另行取得，审查不产生真实资格。
