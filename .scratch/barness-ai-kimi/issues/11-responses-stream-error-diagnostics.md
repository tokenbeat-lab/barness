# 11 — 核对Responses嵌套SSE错误的安全分类

Type: bug
Status: needs-triage

## Context

- [Spec R11](../spec.md)
- [结果03原始流与归一化结果](../evidence/candidate-live-result-03.md)
- [测试SOP](../../../docs/agents/verification.md)

## Observed

自身真实Responses捕获的顶层event:error含嵌套error.type/code=rate_limit_exceeded，并由response.failed确认；当前Result为StopReasonError，安全文本Kimi stream error (unknown)。调用仍正确失败、无重试、usage/cost unreported、资源归零。厂商自由文本只在脱敏证据中保留，不应回显为SDK公共诊断。

## Next step

先用原始已脱敏SSE在公共Client四入口建立确定性回放，核对固定错误分类是否遗漏该嵌套schema或类型，再决定是否修改解析/归一化。不得先改code再补测试；不得把外部engine overload归咎于本地SDK，也不得因诊断修复把失败turn标成功。若修复改变来源，重新冻结普通并评估已完成Chat候选等资格。此票尚未运行回放、未定位或修复软件根因，不授予厂商重试权限。
