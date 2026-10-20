# Ticket 08 — Host examples and complete rollback

2026-10-18，公共 Client / 外部 loopback Provider / 独立软件回滚 CLI 的合成机制验收。
实现提交 `43970b5`、`8dc7c9b`；已合入最新集成，最终验证源码为 `055a55d`。
本票没有真实凭据读取或厂商收费请求，不发布正式支持；实际 M3 证据和正式价格仍由
[工单10](../../issues/10-formal-catalog-support.md) 汇合，M3.1 资格及证据独立保留。

最终结果：79 个 MiniMax Host cases 普通/race 全部 PASS；88 个受影响旧 Host cases
普通/race 全部 PASS；MiniMax 完整软件回滚与 GLM 原有回滚各 1 case 普通/race PASS。
各最终包的 `audit.json` findings 均为空；归档后重新校验了全部 2335 个 manifest
artifact 摘要。三个受影响 example 包 `go vet` PASS，所有本票代码/指南文件 ≤500 行。

```sh
go test -v ./ai/e2e -run '^TestMiniMaxHost' -count=1
go test -race -v ./ai/e2e -run '^TestMiniMaxHost' -count=1
go test -v ./ai/e2e -run '^(TestMiniMaxFullRollback|TestGLMFullRollback)$' -count=1
go test -race -v ./ai/e2e -run '^(TestMiniMaxFullRollback|TestGLMFullRollback)$' -count=1
go test -v ./ai/e2e -run '^(TestLocalAssemblyExample|TestHostIntegrationExample|TestHostIntegrationStreams|TestMixedHosts|TestClassifierHostMigration)$' -count=1
go test -race -v ./ai/e2e -run '^(TestLocalAssemblyExample|TestHostIntegrationExample|TestHostIntegrationStreams|TestMixedHosts|TestClassifierHostMigration)$' -count=1
go vet ./ai/examples/localassembly ./ai/examples/hostintegration ./ai/examples/minimaxhost
python3 .scratch/barness-ai-minimax/evidence/ticket-08/rollback.py .evidence/minimax-rollback-review
(cd .scratch/barness-ai-minimax/evidence/ticket-08 && shasum -a 256 -c SHA256SUMS)
```

最后一个 CLI 参数必须是新的仓库相对目录。普通/race 全量测试由工单09执行；此处只做
本票及受影响 Host 验证。归档内 manifest/assertions 提供逐 case 的复跑命令。UTC 运行
时间对应北京时间 2026-10-18。日志只替换绝对 checkout/home 路径；原日志 Hash 及完整
归档内容的摘要记录在 [provenance.json](provenance.json)，没有改写 PASS/FAIL 或审计。

| 失败先行证据 | 初始失败 | 修复后验证 |
| --- | --- | --- |
| baseline-original-audit-failure.log / audit.json | 原基线全部测试断言通过，最终审计发现 commands.json 两处 home 路径 | 保留原审计；修复 writer 的日志与 argv 脱敏 |
| red-rollback-receipt.log / tar.gz | 公共 GLM 回滚 receipt 两处绝对 argv 泄漏 | green-rollback-receipt；最终 rollback 普通/race |
| red-bindings.log | Host purpose/三 Binding 接口缺失，公共测试无法编译 | green-bindings；最终79个 Host case |
| red-full-rollback.log / tar.gz | 完整软件回滚 CLI 缺失 | green-full-rollback；最终 rollback 普通/race |
| red-local-catalog.log | 实际 localassembly 无显式 Catalog，候选无法装配 | green-local-catalog；最终普通/race 六身份 |

`host-final` / `host-race-final` 覆盖两租户、三协议、两用途、四公共入口、三类 hook
顺序与完整流结果、工具往返、独立允许列表、用途拒绝、配置撤销/恢复、已消费与在途
结果不变、其他租户/用途/路线保持可用、云端 session 归属、可信原生持久化和本地显式
KeySource/Catalog 快照。每个 denied/disabled 场景断言零凭据/Attempts/HTTP。合成 Key
和签名只属于测试；请求、事件、Result、Observer、读凭据探针与错误均在独立 case 中。

`rollback-final` / `rollback-race-final` 的 MiniMax case 独立恢复固定提交
`ecfd4b4d6ec54db30453f4b5112989e9fdbd0782`，tree
`b9798a4e96dfb08cbfcce45c182915589b42eed1`。7039 个完整路径、原始 blob、可执行/符号
链接 mode 在运行前后都核对；不是仅删型号。旧 runtime、目录/价格、矩阵、差分 ledger、
发布合同和依赖均来自此完整树。未改写旧模块源码的公共探针证明六个 MiniMax 身份不再
存在且调用零凭据/HTTP/Attempts；原有本地路线的公共 E2E 与独立审计继续 PASS。

回滚报告保留完整 tree 清单、固定模块 pins、旧发布文件索引、编译/执行 receipt、六身份
拒绝及旧路线 case 摘要。caller 和用户主 checkout 的 HEAD、分支、状态、已修改文件及
未跟踪字节/权限前后一致；E2E 另种植自己拥有的未跟踪文件验证保留。临时恢复树已删除。
原 GLM 回滚也重跑，但保持其自身固定基线和历史合同，不借本票续期真实支持证据。

Host 示例见 [English](../../../../ai/examples/minimaxhost/README.md) /
[简体中文](../../../../ai/examples/minimaxhost/README.zh-CN.md)。ADR-0037 仅将旧的
“不需要二元标志”理由标明被 ADR-0040 取代，保留独立采样默认标签决策。SDK Messages
`?beta=true` 是传输事实；无 beta header/body 能力或 wire shim。指南的候选支持、USD
价格单位、默认思考、兼容控制与可信原生状态边界均按当前实现如实记录，最终状态待09/10。
