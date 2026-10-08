# 工单 05 两轴审查

基点：10abbd7c90733ed19bc2df627353cb00fd98cf3d。两名只读审查者分别按仓库标准和工单/spec 审查本轮暂存变更，并复核修复。

## Standards

初次报告一项 P2：ai/typesafe_response.go 使用 encoding/json 将外层响应和 usage 直接转换为 map，重复 answers、usage 或 input_tokens 会静默采用最后值。公开 Client.Classify 复现可成功返回并错误登记覆盖后的用量。违反 AGENTS.md 的边界校验原则和 ADR-0021 的严格响应契约。

处理：先写四个公开入口失败场景（外层 answers、usage、token、model 重复），记录 envelope-red.log。外层字段在 map 转换前检查，重复字段不进入权威数据；usage 独立严格解码。独立有效用量保留，歧义用量标记 unreported。定向 race 通过。

复核：该 P2 已解决；复现返回 protocol/response、空 Answers，歧义用量不再作为 complete。无残留问题，也无其他模块边界、并发、资源生命周期、隔离或既有代码异味问题。

## Spec

初次报告同一 P2：外层重复 answers 会隐藏无效答案，违背每个最终问题恰有一个同类型答案和先登记用量再整体校验的要求。

处理及复核：重复外层字段在转换为 map 前被识别并排除，不能通过最后值覆盖绕过校验；四个公开入口 E2E 验证拒绝、独立有效用量保留与歧义用量处理。P2 已解决，无其他未解决 Spec findings。

最终结果：两轴均无未解决 finding。
