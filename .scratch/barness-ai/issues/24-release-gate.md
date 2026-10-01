# 24: 发布门禁：压力验证的资源策略示例与脱敏审计

**What to build:** 维护者得到可发布的 barness-ai：附带经压力场景验证的有限资源策略示例及数值依据、完整的公共契约/错误/装配说明、支持矩阵和差异登记，且所有门禁绿灯（spec I10 后两条、Testing Decisions §6）。

**Blocked by:** 12, 13, 23

**Status:** ready-for-agent

- [ ] 正式资源策略示例在可重复压力场景下跑通，记录各数值依据（部署策略，不构成 pi 兼容承诺）
- [ ] 对离线与 live 证据包、日志、观测、错误做脱敏审计，无 key、Authorization、非合成正文泄漏
- [ ] 交付公共契约、错误分类与装配说明、支持矩阵、模型/价格快照、fixture、差异登记（含 D1/D2 与 DeepSeek Responses 扩展）
- [ ] 用例报告可追溯：研究 T/C 条目 → E/P 场景 → 实际证据；追溯表中的“已映射”不被填为 PASS
- [ ] 门禁全部通过：`go test ./...`、`-race`、`go vet`、全部 P0、无待处理差分、六组合 live PASS（或明确 UNSUPPORTED 的能力）
- [ ] 报告分别列出离线、差分与 live 结果
