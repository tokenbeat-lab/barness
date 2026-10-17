---
status: accepted
date: 2026-10-18
---

# MiniMax 固定价格快照与有凭据的目录晋升

M3 中国区标准按量的三个协议身份采用独立固定快照。人民币每百万 token 的输入、输出、
缓存读价格在输入侧合计不超过512000时为2.10/8.40/0.42，超过时为4.20/16.80/0.84。
输入侧包含缓存读写。换算采用2026-10-16官方固定6.7220 CNY/USD，USD费率保留十位小数。
源文、日期、SHA256和换算依据固定在 ai/internal/testkit/minimax 的价格清单；运行时不访问汇率。

自动缓存写没有可靠费率，CacheWrite零值不声明免费。对应标准快照收到正写入用量时，Client
保留 Result/Attempt/Observer 的已知用量并返回 protocol/stream 失败；晋升和正式认证同时拒绝。
合成候选机制测试的明确非零写入费率仍可验证用量归一化。取得实际正写入及可靠价格证据后，
再做证据支持的最小费用语义变更。M3.1不复制M3售价，也不进入正式公共费用目录。

所有 Go 执行源码、测试、验收器和辅助工具先冻结审查。三个M3独立自身、实际、完整候选报告
通过当前源码/SDK/模型/固定计划/已知用量/脱敏/封存验证后，minimaxpromote 工具先独占写入
候选报告和seal的SHA256、候选catalog/source/plan/SDK/Git凭据，再原子替换嵌入目录JSON。
编译后的运行时只读取模型目录数据，不读取验收文件或环境。候选、controlled、research和replay
均不能制造正式PASS；正式验收必须在晋升后的实际BuiltinCatalog Hash下重新取得三份自身报告。

数据晋升后更新现有中文效果路由的目录pin并重新冻结执行输入；旧14身份和四组效果证据全部
重新采集，不能给旧capture重算hash。所有Go源码必须在Jev全源码封存前完成。失败/中断、预算
和候选历史独立保留；回滚使用既有Host目录和binding机制，并按当前公共回滚E2E验证。

[规格R06/R18/R25–R28](../../.scratch/barness-ai-minimax/spec.md)、
[工单10](../../.scratch/barness-ai-minimax/issues/10-formal-catalog-support.md)。
