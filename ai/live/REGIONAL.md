# 新地区服务的独立资格验证

`zai` 和 `minimax-cn` 的六条路由使用新命名空间与一次性授权；旧 GLM / MiniMax 授权、预算余额、捕获和目录不能转用。默认普通测试不会访问厂商。当前九个新型号身份仍是 `regional.CandidateCatalog()` 中的候选，矩阵保持 `NOT_RUN`；受控流水线通过不代表真实资格完成。

## 无 Key 检查

```sh
go test ./ai/e2e -run '^TestRegionalProviderRelease'
go test -tags live ./ai/live -run '^TestRegionalCandidate$'
go run ./ai/live/cmd/regionalverify -source
go run ./ai/live/cmd/regionalverify -grant-template zai-chat
```

最后一个命令输出当前端点、source SHA、目录版本/Hash、SDK、型号和固定上限；ID、账户/地区别名和有效期仍需填写，因此模板不能授权请求。`-controlled <原生目录>` 只验证受控材料；`-qualify <目录>` 只接收自身真实候选或正式执行；`-formal <目录>` 还要求已晋升目录与一次新的正式执行。每次命令只允许一个消费模式。

受控 E2E 直接运行真正的 live-tag 二进制，通过公共 Client 和 SDK 发到 loopback。它保留官方最终 URL、SDK 认证形态、所有请求/流响应、事件、Result、Observer、工具、原生历史、资源与原始授权回执；`mode=controlled` 进入 grant、回执、manifest 和报告，不能认证 live。原件直接保存在父证据目录旁，工作树删除不会销毁或更新它们。篡改实验只修改明确标记的受控副本。

## 将来执行自身 live 的合同

必须先按 `docs/agents/verification.md` 获得本次六条路线之一的明确付费验证授权。账户、地区与用途别名由 Host 维护者确认；只允许 standard API，不能从 key 外形猜测服务或转用 coding / M-plan。每条路线单独进程、独立 key 和新 grant，不能从旧授权推断许可。

1. 固定完整运行源码，使用 `-source` 输出的 `sourceSha256` 构建：

   ```sh
   go test -c -tags live -ldflags '-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource=<sourceSha256>' -o /absolute/path/regional-live.test ./ai/live
   ```

2. 用 `-grant-template <combo>` 生成机器事实。获准后填写新的合法 `id`、真实自身账户/地区的非敏感别名，以及未来 24 小时以内的 `expiresAt`。保留 `mode=live`、`stage=candidate`、精确 Provider / endpoint / source / catalog / SDK / models 与固定预算。源码或目录变化必须重新生成并获准，不能刷新旧捕获。
3. 在清理了其他路线、其他凭据和端点覆盖的子进程中设置 `BARNESS_AI_LIVE=1`、`BARNESS_REGIONAL_ROUTE=<combo>`、`BARNESS_REGIONAL_SERVICE=<实际 Provider>`、`BARNESS_REGIONAL_GRANT=<授权文件绝对路径>`、`BARNESS_REGIONAL_KEY=<由受信 Host 注入的该服务 key>`、`BARNESS_AI_EVIDENCE_DIR=<原件永久目录>`，运行二进制：

   ```sh
   /absolute/path/regional-live.test -test.run='^TestRegionalCandidate$' -test.timeout=1800s
   ```

   默认不使用环境代理。如确需已获准的代理，只能通过 `BARNESS_REGIONAL_PROXY` 显式指定 loopback HTTP 代理；它不改变记录的官方目标。真实模式禁止 `BARNESS_REGIONAL_CONTROLLED`。禁止使用 shell trace 或将 key 放入命令行参数。

4. 运行 `regionalverify -qualify <原生目录>`，保留进程退出码、argv、时间、source/build/binary SHA 和所有原件。首次发送前在可信 Git common directory 的 `barness-regional-grants/` 账本中持锁独占保留授权的规范 scope 和 ID，并完成文件及目录同步。所有 linked worktree 共用此授权域，复制/移动 grant 或只改 ID 均不能重置预算；失败、取消或成功后都不能重用。旧 `.used` 原件保留，不迁移、不覆盖。

账本属于可信 Host 的授权存储，权限仅限运行 Host；不能复制仓库建立新授权域来复用批准，也不能删除、编辑或从旧备份回滚账本。它不防范有权修改批准或 Git common directory 的管理员，也不是跨独立仓库/主机的分布式审批服务。有效期按同一 instant 的 UTC 表示计算 scope，等价 RFC3339 offset 不能增加预算。新运行需要独立批准的新 ID 和新 scope（包括获准有效期），不能只改标签续用旧预算。中断或部分写入永久占用该批准；消费者要求两份完整保留原件与原生回执字节一致，缺失、部分或歧义保留均拒绝，Host 必须保留失败记录并重新申请批准。

每进程最多 24 个逻辑调用和 24 次 HTTP Attempt，零重试；单调用最多 120 秒，进程 1800 秒。单次输出预留取 `min(4096, 自身目录 MaxTokens)`，累计上限 98304。输入按实际编码 wire 字节加 1024、至少 2048 预留，累计不超过 120000；超额在发送前拒绝。每型号固定验证 full/simple 的 stream/complete、Host 工具调用和结果、原生历史、内容后取消与恢复，共九项。ZAI 每路由两型号 18 次，MiniMaxCN 每路由一型号 9 次。未报用量、未知缓存写价、截断、错误终态或缺任一原件均不能认证通过。

## 晋升与发布

候选资格通过后仍需独立审阅型号事实并晋升正式目录。正式阶段使用 `-grant-template <combo> -stage formal` 生成新授权，再设置 `BARNESS_REGIONAL_STAGE=formal` 并运行 `^TestRegionalFormal$`。目录尚未晋升、与自身事实不一致或原源码已变化会拒绝，候选捕获不能替代正式运行。

`regional-report.json`、原始授权/回执、完整执行源码、目录与来源快照、manifest、audit 和所有型号用例必须同属一个原生 bundle。消费器重新核对实际端点/模型/账户/Provider/API/SDK、预留与用量总计、每项单次 Attempt、事件终态、Host 工具配对、原生 reasoning 和资源归零，并拒绝隐藏、缺失、额外或改动文件。原始响应通过公共 Client 的现有 SDK/协议解析器独立重放；完成调用的正文、工具、推理/签名、原生状态、终态、用量数值/报告完整度和估算必须与保存的 Result/Attempt/Observer 相符，预算总计取独立解析用量。缺少、部分与明确零值用量保持不同。取消允许未完成帧、缺少终态/用量和已捕获但尚未发出的 read-ahead；保存的内容/原生字段必须为实际观察前缀，用量必须由该完整帧前缀支持。离线重放不发网络请求，也不产生真实服务资格。只有 `-formal` 通过的 `live-report.json` 可进入支持矩阵，按该自身 row 合并，不认证兄弟站点或另一个型号。

发布计划追加六条路线到旧 17 条；原有八项冻结差分保留，并追加 ZAI Chat 与中国 MiniMax Messages 的真实冻结差分，其余四条按精确 Provider/API 注册扩展。`REGIONAL-R01..R20` 和 `REGIONAL-RP01..RP20` 各自固定完整有限 case ID 集，不能用一个匹配 PASS 替代全部验收。涉及真实资格的条目还依赖 `LIVE`。历史 GLM 捕获仍保留原件与事实；共享运行源码改变后，原生消费者应拒绝其作为当前支持，不刷新 pin 或时间掩盖过期。
