# 旧目标的 xAI 来源历史验证

这份合同只准备验证软件，不授予厂商调用权限。当前正式范围为 Grok 加16条已发布旧路线，共17条；六条 Kimi 保留候选及受控兼容验证，不晋升。新增场景名为 `xai-origin-history`：同一 assistant turn 中的两个长公共前缀 opaque pair，以及后续 turn 对第一个完整 pair 的复用，分别关联三个固定可见结果标记。输入由公共 Host 消息和 `TrustNativeState` 构造，全部标为合成数据，不复制实际 xAI 密文或资源 ID，也不需要 Grok 请求。

`ai/internal/testkit/grokhistory` 检查实际目标 wire 的协议字段、三个唯一关联及结果顺序、可见历史、两个 cross-model downgrade 和不可变原输入。由于 Message JSON 省略可信状态，另以 `history-provenance.json` 保存两个 assistant 的公开 Envelope() 在调用前后必要字段，消费者在4096字节边界内对固定合成来源及实际租户独立校验。MiniMax/Kimi Responses 的关联值可以安全编码完整来源身份；这种关联值不等于向目标重放外来 native `id` 或密文。公共 Client 覆盖全部 22 条兼容路径；原生消费者仍要求完整能力集合，不能用单场景增量包替代。

| 原生族 | 完整场景 | 有限预算 |
| --- | --- | --- |
| Generic 六条 | 原 base/extras 后追加默认型号场景；9 / 9 / 11 / 12 / 12 / 7 | 保留32 logical calls、每场景最多一次环境重试、4096 output/call；新单进程上限1800秒，call180秒、scene240秒 |
| Zhipu Chat / Messages | 每个 glm 型号各增加一场景，22→24 | planned24/max24 calls、max32 attempts、零自动重试、4096 output/call、call120秒、process1800秒、finalization20秒 |
| ZAI Chat / Messages | 每型号各增加一场景，18→20 | 保留24 calls/24 attempts、98304 requested output、120000 reserved input、call120秒、process1800秒 |
| MiniMax CN 三条 | 9→10 | 同上区域合同 |
| Kimi 六条候选路径 | 9→10 | 明确10 calls/10 attempts、40960 requested output、131072 reserved input；保留原 binary/USD/pacing 控制 |
| 国际 MiniMax M3 standard 三条 | 13→14 | 保留20 calls/24 attempts、40000 requested output、120000 reserved input、call120秒、process1800秒；原有限 rerun 规则仍受这些总额约束 |

Generic 场景计数按 OpenAI Responses、DeepSeek Responses、OpenAI Chat、DeepSeek Chat、Anthropic Messages、Google Gemini 排列。ZAI Responses 不在本次 allocator 受影响集合中，保留原18场景。Kimi 的131072输入预留由完整十场景受控请求的实际编码字节加1024（最低2048）逐项求和验证，并保留 `fixture-reservation-proof`；实际执行仍逐请求预留，超过边界即拒绝。

## 新单进程授权

Generic 六条、Zhipu 两条和国际 MiniMax 三条使用 `livegrant` 的固定 scope。scope 绑定路线、mode/stage、官方 endpoint、Provider/API、完整型号和场景、目录版本/hash、SDK、源码/hash、冻结 executable/hash，以及该原生族已有的调用、token、重试与时间边界。账户/地区别名、唯一 approval ID 和最长24小时的有效期由另行批准的具体范围提供。模板中的空人工字段不构成授权。

没有原生 USD 硬上限的路线不新增虚构账单上限；缺失 usage 继续保持未知。受控和 replay 数据不能取得正式资格。GLM 的 `VerifyGLMControlledCapture` 只检查完整受控原件并返回错误/成功，不产生 support report；`NewGLMReport` 仍要求实际 live、正式目录及本次授权。

第一次实际 HTTP 前，producer 用 `O_EXCL` 在 common Git 目录下持久化 approval ID 和规范化完整 scope 两个 claim。复制路径、改 ID 后重用同一 scope、同进程重复领取、崩溃或部分写入均不能恢复预算。每个新 paid process 都需要新批准。原始授权字节保存在 `process-grant.json`，claim 在 `process-authorization-receipt.json`；独立消费者重新检查授权、当前机器事实、manifest 的 build source 和权威 claim 原件。

国际 M3 的当前实际入口拒绝 `BARNESS_MINIMAX_FORMAL_RENEWAL` 和 `BARNESS_MINIMAX_LEDGER`，使用新 claim 所属的独立有限账本。历史 stage/renewal 原件和固定历史 pins 保留，不恢复、不改写，也不新建恢复代次。显式受控授权可走同一新单次账本进行离线验证，仍不能正式发布。

| 族 | 唯一所选 credential | 新授权及非秘密配置 |
| --- | --- | --- |
| Generic | `BARNESS_AI_LIVE_KEY_<ROUTE>` | `BARNESS_AI_LIVE_PERMISSIONS_FILE`、`ACCOUNT_ALIAS`、`REGION_ALIAS`；显式 `BARNESS_AI_LIVE_PROXY` |
| GLM | `BARNESS_GLM_KEY` | `BARNESS_GLM_GRANT`、`BARNESS_GLM_ACCOUNT_ALIAS`、`BARNESS_GLM_REGION_ALIAS`；专用 `BARNESS_GLM_PROXY` |
| 国际 M3 | `BARNESS_MINIMAX_KEY` | `BARNESS_MINIMAX_GRANT` 及原账户/地区、standard/entitlement 配置；显式 `BARNESS_MINIMAX_PROXY` |

不继承系统代理或其他厂商 credential。授权文本不得包含 credential。外层 `tools/verification/live_environment.py` 只转发已选择路线的字段；正式执行时还需以 `run.py --input` 显式绑定授权及批准范围等不可变输入。

## 冻结 build 与只读模板

`processgrant -source ROUTE -stage formal` 在构建前输出当前 `sourceSha256`、机器 scope 和完整 `buildFlag`。它不需要 binary、credential 或网络，也不领取授权。六条 Kimi 继续用独立 `regionalverify` 候选模板；正式模板对未晋升型号的拒绝保持有效。

三个源码哈希公式分别是：

- Generic：`RegionalSourceHash(RegionalExecutionSource())`。
- GLM：`RegionalSourceHash(GLMExecutionSource())`。
- 国际 M3：`Hash(json.Marshal(ExecutionSource()))`，使用紧凑规范 JSON。

这三个来源库存都包含固定合成历史的 Go 源码和授权实现。用只读事实中的精确 flag 构建所选 producer，argv 形如：

```text
go test -c -tags live -ldflags=-X=github.com/tokenbeat-lab/barness/ai/live.legacyBuildSource=SOURCE_SHA256 -o FROZEN_PRODUCER ./ai/live
```

构建环境必须遵循 SOP，显式 `GOENV=off`、空 `GOFLAGS`，保留 argv、toolchain、来源 map 与 binary hash。之后 `processgrant -template ROUTE -stage formal -binary FROZEN_PRODUCER` 生成不完整模板，填入批准范围并冻结。producer 要求嵌入的 `legacyBuildSource` 等于当前 native source hash，并验证当前 executable hash；旧 binary 不能借用新磁盘源码。manifest 保存 `process_build_source`。不包含这个 flag 的旧构建不会取得新授权。

`processgrant -check` 是只读校验。它的 `-claim` 仅允许显式 controlled admission 测试；真实授权必须由所选 producer 在执行边界领取。所有这类模板和构建操作都不替代用户对实际厂商调用的批准。
