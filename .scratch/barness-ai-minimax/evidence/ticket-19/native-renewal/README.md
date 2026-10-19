# 票19 M3 正式源码续期边界（无 Key）

源码提交 `f599c6cee2a8c7a666cfbcccb5ad44ccd94b575e` 增加唯一续期 `current-source-eleven-renewal-1`。规范 common-root grant 绑定新的人类216/252批准、当前源码与审查/二进制/计划 pins、原正式账本39预留/36HTTP 的完整SHA和尾hash。新账本固定60/72、每身份20/24，原13场景、4096输出和40000/120000预留限制不变；独占身份 claim 阻止更换进程/session/目录或删除账本重新消费。原账本、旧 launcher/origin 和所有正式历史报告均未改动。

本记录只证明离线机制。没有读取 `.env`、真实 Key，没有付费请求，也没有建立真实 APPROVED grant。私有 Git 夹具中的 APPROVED 文件是可重放的合成数据，不能授权主仓库或证明真实厂商正式 PASS。

最终 native 辅助边界 ordinary/race 均 exit0；22种授权/来源/预算/重启组合，加独立缺批准与三种非 live/导入grant拒绝；原跨工作树阶段隔离也通过。MiniMax 七组公共受控流程覆盖现有 M3/M3.1 三协议、资源/重试、历史前驱、捕获完整性和重放：ordinary67PASS、race67PASS，audit0；双 vet0。Go 根辅助进程的 SKIP 是只供子进程执行的入口，不是正式场景缺项。

最终 native ordinary/race、harness race、双vet测量 HEAD 为 `751ad94` 加 WIP。`raw/final-source-measured-v3.json` 的872份 compiler Go/mod/sum 字节与最终提交逐一相等；不能写成已提交 HEAD 的重新测试。harness ordinary 使用 v2，v3 仅补充报告预算说明与隔离夹具；完整 harness race 使用 v3。`raw/committed-byte-equality.json` 保留边界。

保留所有实际历史分支：01新环境变量被旧代码忽略；03最初严格拒绝的占位实现尚不支持合法续期；05公开新 CLI flag 尚缺；07非 live 配置仍记录13条失败 call entries（实际预算/HTTP均0，不能称13次厂商请求）；08未先验证的 commit option 在私有夹具写出 sentinel；09最小无效包先报 seal 缺失，尚无新续期语义拒绝（不是非法包被接受为正式）；23留存入口尚未创建。04原 duplicate 夹具误用已登记可重试场景 full-stream，纠正为不可重复的 preflight-error 后通过；14发现无续期选择时仍查询 Git 路径误伤不带 PATH 的旧受控子进程，修复后15/18完整回归通过。原数字退出码和日志均不覆盖。中间早期 WIP 未逐阶段固定完整 compiler map；不把后来的 final map 冒充所有 RED 的源码冻结。原生产基点751与最终全部变更文件/diff在包内，最终测量 v1/v2/v3 map 原样保留。

额外私有公开辅助进程固定3身份基线39预留/36模拟Attempt，再做一次已登记同session rerun，最终新夹具账本40/37；原夹具账本保持39/36。它只调用公开 Ledger API，没有 Client/HTTP，真实 Provider 调用0。整棵私有账本/grant/Git/source 和测试二进制被保留以供复核，不能拿它作为 actual formal 报告。

`index.json` 固定636个普通原始成员和源码 pins；51,392,890成员字节保存在 common-root `.evidence/minimax-retained/ticket-19-native-renewal/raw.tar.gz`（28,258,717压缩字节）。不复制大 Git blob。公开消费者复用票13原有有界 reader，固定32MiB压缩/192MiB解压、32768项、32层、64MiB单成员、1MiB观察元数据、8MiB清单；不解包、不执行捕获的命令。`verification.json` 是一次实际完整独立遍历结果。

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-19/native-renewal/verify.py --common-root /Users/cyber/RestoX/barness

go test ./ai/internal/testkit/minimaxlive -count=1
go test -race ./ai/internal/testkit/minimaxlive -count=1 -timeout=5m
go test -race -tags live ./ai/live -run '^TestMiniMax(FormalRenewalControlledRefusal|Process|EvidenceContract|ResearchPredecessor|FailedResearchPredecessor|InternationalHistoricalAccounting|ReplayProcess)$' -count=1 -timeout=5m
```

原生只读 facts 入口：`go run ./ai/live/cmd/minimaxverify -renewal-source`。输出恰好八项：renewalId/commonRootSha256/sourceCommit/executionSourceSha256/originalLedgerSha256/originalLedgerTailSha256/originalCalls/originalAttempts；尾hash为最后原生 Receipt.Hash，不是原始末行的SHA。它不创建授权。真实新轮须父层另行冻结、复审并取得新批准，后续资格由新真实原件与完整发布门禁独立决定。
