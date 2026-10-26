# Host 装配、生命周期、回滚与普通回归

Type: task
Status: resolved
Blocked by: 02, 03

## Contract

[Spec](../spec.md)；[设计](../design.md)；[验证 SOP](../../../docs/agents/verification.md)。

R03/R04、R16–R20、R22/R23；E13–E16/E18。完善公共 Host/runtime E2E；独立 live 候选入口与 Host 装配说明归兄弟工单07。集成阶段继续两轴审查、最终来源冻结与完整受影响普通包回执，不能用本工单定向回执替代。

## Ownership

xAI Host/运行/回滚 E2E及审查发现的公共验收断言；按需协调共享接线。07拥有候选入口/配置及 Host 装配文档，二者各自依赖02/03后在05汇合。

## Verification

通过 spec 已确认的公共 Client / Catalog 接缝，先保留失败场景，再实现并保留独立 green attempt；遵循 SOP，不能隐式运行 live/race/压力。当前事实统一见 [spec 当前状态](../spec.md#唯一当前状态)。

## Answer

已在 `codex/grok-host-verification` 提交 `c924c31181a374563fb500c1aeb033a48e3ccdf0`；合并了集成分支至 `39e0459c`。本工单新增7个公共接缝 E2E 文件，无生产代码、依赖或正式目录改动；单文件均少于500行。已有实现通过本工单行为断言，不虚构需要实现修复的 Red。

- E14：401/403/429/408/409/500/503 的明确 Code/Phase/HTTPStatus、默认0重试与显式有限重试、重试次数耗尽与延迟上限；失败正文不进入错误/Observer，未验证 xAI request-ID 保持空。headers→payload→response各一次，response先于领域事件；重试固定payload字节/Key/Binding/凭据版本，后续调用读取轮换版本且缓存键不因轮换变化。回调错误产生唯一终态并释放资源；按 R19 保留同步 Host panic 传播语义。307不跟随重定向。
- E13/E16：同一Client上，两租户使用完全相同BindingID/AccountScopeID/CredentialRef标签，与旧OpenAI Responses三次真实HTTP调用受控重叠；各自 bearer、缓存键、Result、每个事件归属、Observer和用量保持独立。
- E15：请求/frame/tool JSON/事件数/队列字节限额、密文及重复终态计入总输出；受控header/read deadline和发送前/headers前/可见内容后取消，均只结算一次、无中流重放。限额/取消后记录 body、生产者、内外Admission、waiter及queue归零，同一Client的下一次显式Complete恢复；Complete在极小队列限额下成功且峰值事件队列为0。
- E18：飞行中禁用Binding新版本，新调用在凭据前拒绝且不fallback；已固定调用继续按旧Binding/价格快照完成，caller目录变更不改Client快照。保留的历史/ciphertext/成本字节不变，显式恢复Binding后按原生历史续接。
- 审查补强 E02/E06/E07/E08/E09：显式auto/none成功；实际返回的工具+密文Result直接及经JSON持久化/TrustNativeState恢复后进入下一轮公共调用；跨API重命名与逆序结果逐一关联；`(a|b,c)`/`(a,b|c)`返回tuple产生不同本地ID并精确回放；用户text/png/text/jpeg精确顺序与payload；显式high effort仍带store:false/include密文声明。

### 验证范围与原件

已确认接缝为公共 Client/Catalog，以及 Config.Transport、Hooks、Binding/CredentialResolver。测试仅使用受控HTTP/SSE与既有Host/BodyTracker/Probe；普通范围为新增 `TestXAI(Runtime|Host|Acceptance)` 入口，包含旧OpenAI Responses并发回归。无厂商请求、race、压力或全包运行。

每轮实际命令均经 `python3 tools/verification/run.py --feature barness-ai-grok --stage e2e -- go test -json ./ai/e2e -run '<精确表达式>' -count=1 -timeout=5m` 生成新attempt：

- 分片PASS（均退出0）：重试/轮换 `20261026T023631.565703Z-avwq9ycw`；资源/取消/并发隔离 `20261026T023713.843733Z-itz5ye9a`；第一批补强验收 `20261026T023923.271211Z-nv2yh7mo`；禁用/回调/重定向/补强验收 `20261026T024112.624889Z-kqhnq8bj`。
- 失败attempt保留，不能当作实现缺陷的Red：`20261026T023440.352050Z-29r603dr`缺Observer容量；`20261026T023532.464980Z-svy22p0a`观察文件名含路径分隔符；`20261026T023558.182669Z-eqqoennk`请求字节前置拒绝实际为capability而测试错预期request；`20261026T023820.972645Z-ofnz9a9o`测试把值类型NativeState与nil比较导致编译失败；`20261026T023847.651133Z-iq3pdgse`测试误预期Stream Hook panic被恢复，实际按R19保留传播，测试进程退出。修正测试使其按公共Complete调用由Host恢复并验证释放，未改生产panic策略。
- 最终干净源码定向回执：[receipt.json](../../../.evidence/barness-ai-grok/e2e/20261026T024154.002320Z-vtrgi7we/receipt.json)，实际argv：`go test -json ./ai/e2e -run '^TestXAI(Runtime|Host|Acceptance)' -count=1 -timeout=5m`；退出0，48个原生case实际PASS，唯一 `./ai/e2e` bundle与audit通过。受测SHA为 `c924c31181a374563fb500c1aeb033a48e3ccdf0`；receipt SHA-256为 `60ba24f57f29909e7740673a7f66d3a19f0c855f6eb08cf8d392b692671d6b07`。
- 原生bundle `bundles/20261026T024157.032769000Z`；manifest SHA-256 `a29fffc7f8632648201433ab3c34fd05b83bc7dbbd374ad3b29201b6b4f010e7`；audit SHA-256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。工作树保留于 `/tmp/barness-grok-host-verification`，原件不覆盖。
- 本工单仅结束实现与定向验证。集成负责人仍须对所有payload/native/usage/取消/释放/隔离做两轴审查，修复后冻结并取得完整受影响普通包回执。live为NOT_RUN且仍需要本次明确授权；其后人工race与压力为NOT_RUN，release.pass=false。测试新增未改变生产路线语义；真实资格仍按01–03及07的总体影响处理，不宣称live不适用。回滚为撤回本工单测试提交；运行回滚仍是禁用xAI Binding并保留历史/价格/原件。

## Comments

- 2026-10-26：由 implement-spec 从已就绪规格建立依赖工单；实现不授予后续人工或付费验证权限。
- 2026-10-26：按实际依赖拆开04与07：04只等待02/03，05同时等待04/07；本工单实现完成不等于全局冻结审查或发布完成。
