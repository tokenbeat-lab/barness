# 新地区服务的独立资格验证

`zai` 和 `minimax-cn` 的六条路由使用新命名空间与一次性授权；旧 GLM / MiniMax 授权、预算余额、捕获和目录不能转用。默认普通测试不会访问厂商。候选、晋升和当前真实资格分别判定；实际进度见[地区规格当前状态](../../.scratch/barness-provider-sites/spec.md#当前状态)。受控流水线通过不代表真实资格完成。

## 无 Key 检查

按[测试流程 SOP](../../docs/agents/verification.md)使用清除 live/重负载/凭据环境的普通回执入口。下列 E2E 在受控边界验证地区合同；不在继承环境中直接运行 live-tag 测试包。

```sh
python3 tools/verification/run.py --feature '<功能 slug>' --stage e2e -- \
  go test -json ./ai/e2e -run '^TestRegionalProviderRelease' -count=1
# 以下两个命令只生成来源/模板；不执行验证或请求厂商。
env GOENV=off GOFLAGS='' go run ./ai/live/cmd/processgrant -source regional
env GOENV=off GOFLAGS='' go run ./ai/live/cmd/processgrant -template zai-chat -stage candidate
```

最后一个命令输出当前端点、source SHA、目录版本/Hash、SDK、型号和固定上限；ID、账户/地区别名和有效期仍需填写，因此模板不能授权请求。`releasegate -inspect-bundle <原生目录> -mode controlled -stage candidate` 只验证受控材料；live 模式验证真实原件。当前适用性按 ADR-0055 的本路线实际型号和产品差异核对，原 stage/mode/时间保持。

受控 E2E 直接运行真正的 live-tag 二进制，通过公共 Client 和 SDK 发到 loopback。它保留官方最终 URL、SDK 认证形态、所有请求/流响应、事件、Result、Observer、工具、原生历史、资源与原始授权回执；`mode=controlled` 进入 grant、回执、manifest 和报告，不能认证 live。原件直接保存在父证据目录旁，工作树删除不会销毁或更新它们。篡改实验只修改明确标记的受控副本。

## 将来执行自身 live 的合同

必须先按[测试流程 SOP](../../docs/agents/verification.md)获得本次六条路线之一的明确付费验证授权。账户、地区与用途别名由 Host 维护者确认；只允许 standard API，不能从 key 外形猜测服务或转用 coding / M-plan。每条路线单独进程、独立 key 和新 grant，不能从旧授权推断许可。

1. 固定完整运行源码，使用 `-source` 输出的 `sourceSha256` 构建：

   ```sh
   go test -c -tags live -ldflags '-X=github.com/tokenbeat-lab/barness/ai/live.regionalBuildSource=<sourceSha256>' -o /absolute/path/regional-live.test ./ai/live
   ```

2. 用 `processgrant -template <combo> -stage candidate` 生成机器事实。获准后填写新的合法 `id`、真实自身账户/地区的非敏感别名，以及未来 24 小时以内的 `expiresAt`。保留 `mode=live`、`stage=candidate`、精确 Provider / endpoint / source / catalog / SDK / models 与固定预算。源码或目录变化必须重新生成并获准，不能刷新旧捕获。
3. 在清理了其他路线、其他凭据和端点覆盖的子进程中设置 `BARNESS_AI_LIVE=1`、`BARNESS_REGIONAL_ROUTE=<combo>`、`BARNESS_REGIONAL_SERVICE=<实际 Provider>`、`BARNESS_REGIONAL_GRANT=<授权文件绝对路径>`、`BARNESS_REGIONAL_KEY=<由受信 Host 注入的该服务 key>`、`BARNESS_AI_EVIDENCE_DIR=<原件永久目录>`，运行二进制：

   ```sh
   /absolute/path/regional-live.test -test.run='^TestRegionalCandidate$' -test.timeout=1800s
   ```

   默认不使用环境代理。如确需已获准的代理，只能通过 `BARNESS_REGIONAL_PROXY` 显式指定 loopback HTTP 代理；它不改变记录的官方目标。真实模式禁止 `BARNESS_REGIONAL_CONTROLLED`。禁止使用 shell trace 或将 key 放入命令行参数。

4. 运行 `releasegate -inspect-bundle <原生目录> -mode live -stage candidate`，保留进程退出码、argv、时间、source/build/binary SHA 和所有原件。首次发送前在可信 Git common directory 的 `barness-regional-grants/` 账本中持锁独占保留授权的规范 scope 和 ID，并完成文件及目录同步。所有 linked worktree 共用此授权域，复制/移动 grant 或只改 ID 均不能重置预算；失败、取消或成功后都不能重用。旧 `.used` 原件保留，不迁移、不覆盖。

账本属于可信 Host 的授权存储，权限仅限运行 Host；不能复制仓库建立新授权域来复用批准，也不能删除、编辑或从旧备份回滚账本。它不防范有权修改批准或 Git common directory 的管理员，也不是跨独立仓库/主机的分布式审批服务。有效期按同一 instant 的 UTC 表示计算 scope，等价 RFC3339 offset 不能增加预算。新运行需要独立批准的新 ID 和新 scope（包括获准有效期），不能只改标签续用旧预算。中断或部分写入永久占用该批准；消费者要求两份完整保留原件与原生回执字节一致，缺失、部分或歧义保留均拒绝，Host 必须保留失败记录并重新申请批准。

每进程最多 24 个逻辑调用和 24 次 HTTP Attempt，零重试；单调用最多 120 秒，进程 1800 秒。单次输出预留取 `min(4096, 自身目录 MaxTokens)`，累计上限 98304。输入按实际编码 wire 字节加 1024、至少 2048 预留，累计不超过 120000；超额在发送前拒绝。每型号固定验证 full/simple 的 stream/complete、Host 工具调用和结果、原生历史、内容后取消与恢复，共九项。ZAI 每路由两型号 18 次，MiniMaxCN 每路由一型号 9 次。未报用量、未知缓存写价、截断、错误终态或缺任一原件均不能认证通过。

## 晋升与发布

候选资格通过后仍需独立审阅型号事实并晋升正式目录。按 ADR-0055，完整真实候选测量可以离线取得当前正式适用性，原 stage/mode/时间保持；仅当需要新增厂商事实时，使用 `processgrant -template <combo> -stage formal` 取得新的授权范围，获准后设置 `BARNESS_REGIONAL_STAGE=formal` 并运行 `^TestRegionalFormal$`。目录尚未晋升、与自身事实不一致或产品差异未审仍拒绝。

`regional-report.json`、原始授权/回执、完整执行源码、目录与来源快照、manifest、audit 和所有型号用例必须同属一个原生 bundle。消费器重新核对实际端点/模型/账户/Provider/API/SDK、预留与用量总计、每项单次 Attempt、事件终态、Host 工具配对、原生 reasoning 和资源归零，并拒绝隐藏、缺失、额外或改动文件。原始响应通过公共 Client 的现有 SDK/协议解析器独立重放；完成调用的正文、工具、推理/签名、原生状态、终态、用量数值/报告完整度和估算必须与保存的 Result/Attempt/Observer 相符，预算总计取独立解析用量。缺少、部分与明确零值用量保持不同。取消允许未完成帧、缺少终态/用量和已捕获但尚未发出的 read-ahead；保存的内容/原生字段必须为实际观察前缀，用量必须由该完整帧前缀支持。离线重放不发网络请求，也不产生真实服务资格。只有 `-formal` 通过的 `live-report.json` 可进入支持矩阵，按该自身 row 合并，不认证兄弟站点或另一个型号。

发布按功能影响选择需要自身证明的地区路线，依 [ADR-0051](../../docs/adr/0051-scoped-verification-receipts.md) 写入五项回执；不运行冻结 pi 差分。`REGIONAL-R01..R20` 和 `REGIONAL-RP01..RP20` 各自固定完整有限 case ID 集，不能用一个匹配 PASS 替代全部验收。涉及真实资格的条目还依赖 `LIVE`。历史 GLM 捕获仍保留原件与事实；相关运行行为改变后按路线核验，消费者或无关型号变化不自动使其他 Provider 过期；不刷新原 pin 或时间。

保存新只读原生消费结果后，再按 SOP 第 4 节使用 consume_live.py 取得 schema 2 外层消费者回执，供功能 `-record` 索引引用。普通 E2E、真实调用、正式资格与只读发布判定分别保留事实。

## Kimi 原九项与单项补采

两站六路线继续各自独立。`processgrant -template <kimi-route> -stage candidate -purpose xai-history-supplement -baseline-manifest-sha256 <原九项 manifest SHA256>` 只输出不完整机器模板，不授予付费权限。新 schema 2 仅允许 `kimi-k3/xai-origin-history`，绑定当前源码、冻结二进制、自身账户/地区与原九项 manifest；标准 schema 1 和单项 schema 2 不能交叉启动。新执行仍需人类明确授权，预算每路线 USD 0.10、六路合计 USD 0.60，1 调用/1 HTTP/零重试，输入预留 4096、输出上限 4096，调用 120s。真实模式发送前等候 60s，精确入口 `^TestKimiHistorySupplement$`、`-test.count=1`、`-test.timeout=240s`，其余沿用上述冻结构建与隔离环境。失败仍消耗授权；单项报告 `complete=false`，不生成完整十项支持报告。

目录中唯一 `kimi-composition.json` 含 `schema:1`、`kind:kimi-history-composition`、`combo`、`model:kimi-k3` 及 `baseline`/`supplement` 两个引用。每个引用固定仓库内原生 `bundle`、`manifestSHA256`、`reportSHA256`、`sealSHA256` 和 `producer:{path,sha256}`。producer 指向当次真实工具进程回执，必须保留原 stdout/stderr/数字退出和所运行二进制；旧九项按其原 31 分钟外层 argv 与 1800s 内部预算核验，不能改写历史。拒绝嵌套、重复覆盖、跨路线/型号/账户/地区/SDK、错误基包 pin 和不完整原件。

`releasegate -inspect-bundle <组合目录> -mode controlled -stage candidate` 只读验证受控组合；真实两来源用 `-mode live -stage candidate` 得到候选完整性，`-stage formal` 还要求本型号正式目录。`consume_live.py --bundle <组合目录>` 及发布 record 走同一原生校验，并绑定两个完整原包、进程证明和可信账本。结果分别保留两次测量和十项覆盖；受控不升级真实，目录未激活不升级正式。

普通 E2E 的旧九项来自独立 `live,kimi_baseline_fixture` 标签的 `^TestKimiBaselineControlled$` 实际九次 loopback 进程；该入口不进入普通 live 二进制，且拒绝 live 模式。受控单项允许零等待以缩短测试，真实模板和 live 校验仍固定 60s。本次实施不发放付费授权、不启动厂商 live/race/压力，也不恢复暂停轮次；真实六项仍为 NOT_RUN。

Kimi 晋升入口只消费六个固定九加一组合目录，各 combo 恰好一次。`regionalpromote -kimi -controlled -proposal <新审阅文件> <六个组合目录>` 输出 schema 2、active=false 的专用审阅证明；每路线包含原两次 qualification.measurements、十项 coverage 和完整传递 inputs，六路共十二份独立测量。原单包十项不再作为此晋升入口输入，但其自身完整资格消费合同保留。消费者不要求六路或同路线两次测量同 source/build/binary/commit；逐份检查相关产品/SDK/型号、账户/地区/租户、原时刻授权、预算、实际进程和权威账本。候选目录版本/价格事实由各测量原生合同核验，不能通过取消来源相等性跳过适用性。

将来真实晋升使用同六个组合的 `regionalpromote -kimi -receipt <新独占原件路径> ...`，仅接受 mode=live；controlled 无法写入目录。沿用专用 Kimi inactive 目标哈希、独占 intent 和原子替换。此说明不授权本轮激活或付费补采，账户及费率仍待确认；普通消费者修正只重建未批准二进制/范围，不使其他 Provider 或旧54项事实失效。

晋升审阅/intent输出也必须位于两个原生包、其producer父目录和可信Kimi授权账本目录之外；不得借组合根目录之外的路径新增文件破坏原库存。独占写入仍拒绝覆盖已有文件。
