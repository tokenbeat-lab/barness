# 两轴审查修复

Type: task
Status: resolved
Blocked by: 04, 07

## Contract

对应 [spec](../spec.md) R07/R12/R13/R16/R19 与 [design](../design.md)、ADR-0054。当前状态只在 [spec](../spec.md#唯一当前状态) 维护，本工单只保留修复事实与原件指针。

## Scope and verification

修复 shorthand assistant 原生 ID 授权、候选报告最终落盘失败、xAI 会话头、重复协议控制键，以及 xAI opaque 工具历史降级碰撞。公开接缝为已确认的 Client/Catalog/Transport/Hooks 与原生子进程。逐项 red → 最小修复 → green；仅运行定向普通 E2E，无厂商请求、race 或压力。最后集成审查与完整冻结回执由主工作树负责。

## Failure inventory

- 缺失/null type 的 assistant ID 绕过原生引用基线；合法 shorthand 与原基线 ID 不应误拒绝。
- 候选报告发布后的目录 sync 或最终 audit 失败，不能留下 PASS/Complete/AuditPassed；受控子进程保留真实文件系统失败与精确来源/二进制/grant。
- header 名大小写、空值与多值绕过 X-Grok-Conv-Id 检查；其他 Provider 自定义头合同保持。
- raw reasoning 等协议对象的重复键先折叠后校验；业务参数/schema 以及原生扩展数字不应被全树扫描。
- 不同 xAI opaque 工具调用跨 Provider 降级后必须仍一一配对，不能保留账户 item 状态；旧来源同样回归。

## Evidence

所有回执根为 `.evidence/barness-ai-grok/e2e/<attempt>/receipt.json`，每次均由 `tools/verification/run.py` 留存独立 command、日志、来源库存、唯一主 bundle 与 audit。初次源码迭代的 receipt.source.commit 为 `a3c3a269d15c35563d2ec3c364595a7f179e0849`，完整实际工作树按 receipt.source.files 绑定，不能把旧 commit 冒充最终冻结。

| 范围 | 实际状态 / 数字退出 | attempt | receipt SHA256 |
| --- | --- | --- | --- |
| shorthand-red | FAIL / 1 | `20261026T025841.904749Z-uczas_9_` | `49702cd732b7f7ee21edefb1f854b67290b5c9c0cc111ae534eb96344c0322eb` |
| shorthand-green | PASS / 0 | `20261026T025857.979768Z-5vq41pnh` | `a2ea27353150076d06c068993fdebb9058d847bb301fb581924f7bb0520d2be0` |
| header-red | FAIL / 1 | `20261026T025932.651762Z-j7opy614` | `714d17c5859ac9bbfcbf47d72cb604cfad7cb5b10a3aaf7a9f13bc51a7e554f6` |
| header-green | PASS / 0 | `20261026T025951.423351Z-pidmt8bj` | `04516946006a93608a7086d74588c9886b1adb573816d3de1819ce0f261494b2` |
| duplicates-red | FAIL / 1 | `20261026T030120.530142Z-6np_hjtc` | `300cda14a5474bacfcfa53b803eab8a8b7e193963846713b747941acccf9b081` |
| duplicates-failed-green | FAIL / 1 | `20261026T030202.402345Z-m8mgs7pw` | `403fd3700a38b51dbe3b0b0629b69443db53ab38753b4d82f7d103ff60e81120` |
| duplicates-green | PASS / 0 | `20261026T030258.228117Z-ogvrzz1y` | `6f2acf473918cfd7bf404dc0de7c7f05588ddc1c77d875e87c5f031d06d9735d` |
| history-red | FAIL / 1 | `20261026T030405.608160Z-_obtu2h_` | `595e9f1a1639579d2c231b1ae1b63124d47eed696f4b95f3a7321f54112c47c2` |
| history-green | PASS / 0 | `20261026T030424.919120Z-2x9yp_uq` | `d03d9fd2d49829cb087332620695fb338b3310f9ed4fa6ccbe80b9b47390ad13` |
| late-finalization-red | FAIL / 1 | `20261026T030437.888611Z-j21rm6pz` | `a0ca761d69353b2fbda36d5956f7da9a5c4e9455b865a807d8f9d822aaec95e2` |
| finalization-green | PASS / 0 | `20261026T030552.054632Z-1b2w8bt3` | `593e58a5d96583ca1276775813af8cbf33715d3c602669b49aca0093cf34d342` |

重复键首次 green attempt 的六个失败来自无效 Catalog 夹具：既有配置合同已禁止 model/profile 的 tools/input 采样覆盖。它们不是新缺陷或通过证据；测试缩至真实可达入口，保留该失败 attempt。合法业务 schema/参数字符串与原生 `1e400` 的回归均在最终同来源范围重复覆盖。

子进程受控模式始终零 vendorHTTPAttempts。final-sync 真实关闭目录 fd 后调用 Sync；final-audit 在最终 audit.json 写入路径创建目录，真实触发 EISDIR。修复后保留原 RUNNING 报告、完整24调用预算、源/二进制/grant/claim，canonical/pending 都为非PASS，最终 audit 完整，原生正式消费者 exit1。无需扩大 live 环境白名单，也不产生真实或正式支持资格。

## Implementation

- shorthand discriminator 在原生 ID 指纹处规范为 message，信任仍来自原始 Host envelope；合法 shorthand 和已担保 ID 可保留。
- xAI 专属 header guard 在大小写归一化后拒绝任何非空值集合的 X-Grok-Conv-Id，包括空字符串；空/nil集合无传输值，其他 Provider 自定义头保持。
- 最终边界在 map 解码之前扫描协议对象键唯一性，保留授权优先错误；不遍历业务 schema、参数/结果字符串或未知 native 扩展。
- shared opaque history 只增加 xAI 来源识别，复用完整来源 ID 的碰撞分配，无跨 Provider decoder 耦合。
- 候选报告先暂存并审计再晋升，任何后续 I/O/audit 失败都尽力覆盖为非PASS，覆盖失败撤下可读报告；保留原报告与失败目标后重新审计。

## Final targeted source

提交源码和测试后，合并最新集成文档，再执行全部 TestXAI 与既有 GLM/Kimi opaque 来源回归。完整集成冻结普通回执及后续人工/付费阶段仍由主工作树负责。本工单不改 spec 当前状态。



## Answer

五项修复已提交于 `51c95922`，合并最新集成文档后的受测提交为 `e3ad302d5bffc234a7ad3563235bffdd95d4d819`。

冻结定向普通命令（exit0）：

```sh
python3 tools/verification/run.py --feature barness-ai-grok --stage e2e -- go test -json ./ai/e2e -run '^(TestXAI.*|TestGLMResponsesOriginIDCollision|TestKimiResponsesCrossTargetIDs|TestKimiResponsesSequentialProvenanceIDs)$' -count=1 -timeout=5m
```

[不可变索引](../evidence/review-fixes-e2e-v1.json)绑定 receipt `20261026T030721.568691Z-pzit2yuv`（SHA256 `f23768eeb8242c7a4ff71d374720c155ea03735843c05ef91e23ee7ac57dbf3c`）：64顶层/605实际测试、564原生case全部PASS，唯一主 bundle `bundles/20261026T030725.014638000Z`，manifest SHA256 `f2ef5e24a68e5fb3332fe3b88a9cf2304fb68a962e197fe7d5575f1082ae1c85`，audit SHA256 `8aa973100a15bf817985395637131e838cf84e427be4792583def923f020f5fa`。1961项完整执行来源与空选定输入集合在文档收尾后独立核对相等；相等性原件路径见索引。

本票关闭只表示审查修复与定向普通验证完成。独立复核、集成完整冻结普通回执仍待主工作树，live/race/压力未运行，release.pass=false。回滚可撤回修复提交，不改旧历史、候选授权或失败证据。
