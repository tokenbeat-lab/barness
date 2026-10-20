# 16 — GLM 验收 Host 的响应头预算对齐

Type: task
Status: resolved
Updated: 2026-10-18
Blocked by: None

## Context

本轮GLM Responses两次实际失败约90秒，没有HTTP响应；原始error明确为响应头等待到期。既有GLM R21/ADR0036允许单调用最多120秒并服从Host更短上限。当前GLM专属装配只把CallTimeout设为120秒，继承共享livePolicy的ResponseHeaderTimeout=90秒。[只读政策与来源诊断](../evidence/ticket-13/glm-deadline-policy.json)记录精确文件、限制与原错误。

该90秒较短Host配置在旧合同下有效，旧失败保持FAIL；不能推断上游在120秒会成功。本票是未来新合同的明确验收配置调整，不是改写旧捕获，也不放宽支持断言。

## Scope and failure-first acceptance

只在`ai/live/glm_assembly_test.go`的GLM专属Host装配中把ResponseHeaderTimeout对齐到已有CallTimeout的120秒。共享livePolicy、connect15秒、read-idle90秒、全部prompts/inventory/结果断言、输出4096上限、默认零重试和新增调用预算均保持。

先扩展既有公共`TestGLMCandidateAssemblyControlled`：受控TLS服务的响应头超过90秒、低于120秒时，调用仍应完成；保留可核验RED，再做最小装配修改至GREEN。现有更短Host取消和总调用时限的公开E2E继续有效。没有可瞬时证明该真实90→120边界的公共时钟接缝，不新增只镜像配置字段的私有单测。

该文件属于三份GLM217文件执行映射，三路必须在新同源完整实测；不重算旧GLM证明。M3三份263文件映射应逐字不变，明确验证后才保留原正式证据。拟复用generic10与text2相关runtime/SDK/model/catalog/生产者保持；严格protocol114及视觉2本来就在固定6项重验集合中。

独立工作树单独提交并做Standards/Spec复审，再冻结票15 PENDING合同。票15需要人类单独批准新增108calls/132Attempts；本票仅授权离线修复/核验，不读取厂商Key或新增付费。

## Answer

源码 `dcd30f7ce6bf65f5bd1101697c8e5c98155ceb98` 及留存 `4bfdaa7` 已通过两轴审查，并于 `7231fdad77bd1fb69bb1f21b934e55db77aab814` 合入集成分支。公共 Client 本地TLS的91秒响应头先在旧90秒限制下RED，再仅加GLM专属120秒赋值后GREEN；较短取消/超时路径普通/race及双vet通过。三份M3正式263项逐项不变，三份GLM217项只有专属装配改变。

[完整证据](../evidence/ticket-16/deadline-alignment/README.md)区分本地受控证明、明确WIP来源及旧实际失败。本票没有厂商请求或Key读取，不证明厂商120秒成功；旧Responses FAIL保持，未来三GLM实际同源验收由[15](15-source-scoped-revalidation.md)承接。
