# 05：验证回调授权、重试冻结、多租户隔离与全量差分

Status: resolved
Type: task
Depends on: 04
Updated: 2026-10-11

## What to build

用既有公开生命周期 seam 证明新采样配置在回调、授权、重试、取消和并发中不漂移，并关闭全部目标差分 pending。必要修正仅限已暴露的生命周期集成问题，不扩展 Observer 内容或重建调用框架。

## Ownership and context

负责对应 E2E/fixture、采样需求的差分映射与审计 evidence；如需改生命周期生产路径，先说明失败与原因，沿用现有 Hooks/授权/重试职责。协作时不要回退他人修改。

依据：[规范 R07–R10](../spec.md)、[验收 S26–S30、B03](../acceptance.md)、既有 TestTrustedCallbacks/TestProtocolCallbacks 与隔离/重试场景。用户确认的边界是公共 Client，无上层 Agent 扩张。

## Failure scenarios first

- OnPayload 看不到分档结果、替换 payload 后又被采样覆盖，或回调改变 reasoning 触发 reselection。
- 最终 model/托管资源变更绕过授权；错误、Observer 或 artifacts 泄露配置值/凭据。
- 重试再次执行回调、重选档或读取修改后的配置，多个 Attempt 请求不同。
- 同型号跨租户并发共享可变 map；调用结束后污染后续调用。
- 取消或失败增加发送次数、产生多个终态、资源不归零。
- 新 E2E 通过但旧差分、分类子集或目录对比未跑；扩展路线被伪装成 pi PASS。

## Acceptance criteria

- [x] 先写 S26–S30 失败场景，再修正被证实的集成问题；使用屏障而非 sleep 保证可重复时序。
- [x] 四个 WithHooks 入口的原地/替换回调验证完成；授权与限额在最终载荷上生效。
- [x] 有界重试同 payload bytes、同 profile，回调一次；Attempt 独立，取消后无后续发送。
- [x] 并发与连续调用无目录污染，租户归属正确，资源归零；输入的合法后续修改与返回值修改被隔离。
- [x] B03：全部既有差分子集及新增采样 case 实际运行，零 pending；所有 extension 有规范/ADR 原因及有限 case 范围。
- [x] DeepSeek Responses、Images/Interactions 等原有扩展仅走自己的验证与显式豁免，未删旧 P0 或原生对比字段。
- [x] 专项与全量离线/race 证据逐文件哈希与审计零发现；Observer 不增加采样正文。

## Verification and handoff

按[验收执行说明](../acceptance.md#6-可重复-artifacts-与结果判定)运行目标 E2E、全量开启 pi 差分、race 与两种 vet，交付逐需求证据索引、请求/帧/结果和账本发现。专项 race 通过不替代最终 gate 中开启压力/差分的 race。

完整 release 与 current live/evaluation 合同由[工单 06](06-document-and-release.md)执行；本工单只能声明离线/差分/race 的实际结果。缺证据不填 PASS，不把 research probe 当运行结果。

## Comments

2026-10-11：方案阶段创建，目标差分与隔离测试尚未执行。

2026-10-11 实施：全部生命周期/授权/重试/并发以及全量差分 PASS；完整整版 race/压力/live 结果仍由工单 06 独立约束。 命令、实际 artifacts、失败历史和未完成项见[专项证据](../evidence.md)与[来源核验](../release-verification/README.md)。
