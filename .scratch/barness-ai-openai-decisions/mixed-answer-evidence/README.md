# 工单 03：混合答案、分布与逐题拒答

宿主通过公共 `Client.Classify` 与 `HookedClient.Classify` 在一次文本 Logical Call 中提交 choice/score/bool，
获得完整命名答案。score 保留原值、完整分布及经验证的字符串图例；bool 保留概率。
混合和全拒答成功结束，拒答仅为 `RefusalAnswer{}`。任何损坏答案整体清空，唯一合法 usage 保留。

实施基线 `9fa44f9ffea5252fdc0bcce71ef13e31515b90cc`，实现提交 `aa9da73`，分支 main。
测试接缝由本票指定，失败清单先写在[这里](../mixed-answer-failures.md)，随后记录真实公共 E2E 红→绿。
起点的 GLOSSARY 和其他未跟踪规划保留，仅提交本票改动。
目录为合成 `classifier/openai/openai-decisions/host-decisions`，Provider 为受控 loopback HTTP；live 为 NOT_RUN。
本票没有交付图片、usage presence 专项、正式型号/价格或 Decisions 自身 Pi 差分与真实服务验证。

| 验收 | 可重放证据 |
| --- | --- |
| OD13 | choice/score/predicate wire、bool 无/空/单侧/双侧说明、ordered 重复 label；公共 2/10 档成功及 1/11 档 scope 拒绝，模型收窄能力在凭据前拒绝 |
| OD22/25 | Hook Keep/Replace 改名与三题型互换、重排候选/等级、最终 rubric 校验、迟改深层引用隔离、原对象未修改；非法 schema/型号能力/资源限额零 Attempt |
| OD27/28 | 乱序完整答案、混合/全拒答、refusal metadata；封闭公共 JSON 没有虚构原因 |
| OD29/30/31/32/33 | name 集合、discriminator、重复字段/别名、候选最大概率并列、数值 presence/finite/range、整数分布/labels、和及期望 1e-6 内外对照；失败原子保留 usage |
| OD34 | Jev 固定百分位舍入 replay 保持；同数值在 Decisions 拒绝，高精度不一致拒绝，原概率与 score 不重算 |
| OD39/41 | envelope/answer/entry benign metadata 可忽略但整体输出有界；矛盾答案/非法语义字段拒绝；2xx 不重放，Attempt/Observer/资源释放与审计 |

两个公共入口新增 292 个 E2E 全部 PASS。初次 284 case 中 250 FAIL、34 PASS，扩展生产路径后同 284 case 全部 PASS。
随后专项回归的 metadata 限额测试因装配违反 MaxToolJSONBytes≤MaxOutputBytes 而失败；保留这次 883 case 的
880 PASS、2 FAIL、1 NOT_RUN，修正测试配置后专项与专项 race 均为 882 PASS、1 既有 Pi opt-in NOT_RUN。
失败历史没有改写为 PASS。

普通全套开启冻结 Pi，退出 0：6,720 case，6,711 PASS、9 既有压力 opt-in NOT_RUN；815 项既有 Pi 差分全部 PASS。
完整 race 退出 0：5,894 PASS、826 默认 opt-in NOT_RUN（未开启差分及压力）；新增 292 case 全部 PASS，审计无发现。
Go 编译检查、普通 vet、live-tag vet 均退出 0。完整测试和最终包的审计收据见 `validation/` 与 `audit.json`。
两轴审查及验证器补充审查均 0 项实质发现，见 [REVIEW](REVIEW.md)。实现 Go 源码在测试和打包之间未变化，
最后仅补文档证据链接；`source-snapshot.json` 记录测试时源文件 SHA-256。

`mixed-answers.zip` 保存完整专项捕获、红绿历史、fixture、请求/Provider 脚本、结果/错误、usage/Attempt、
Observer、assertions、版本、源文件快照、日志和完整测试收据。
`manifest.json` 固定所有成员的 SHA-256；`verify.py` 固定 manifest 摘要，用同一份不可变字节快照验证和审计。
manifest≤8 MiB，ZIP≤32 MiB，成员≤30,000、每文件≤16 MiB、总展开≤128 MiB，路径和普通文件类型均检查。
这些边界按本包真实大小取有限预算，未扩大库运行时限额。全套既有 case 仅保存 manifest/audit/进程日志，
本票及其专项回归/race 的逐例捕获完整保存。

核验固定证据（不会请求厂商）：

```sh
python3 .scratch/barness-ai-openai-decisions/mixed-answer-evidence/verify.py --audit
(cd .scratch/barness-ai-openai-decisions/mixed-answer-evidence && shasum -a 256 -c SHA256SUMS)
```

从仓库根重放当前实现；每 case 的 assertions/manifest 还保存逐例准确命令：

```sh
go test ./ai/e2e -run '^Test(Decisions|JevRegression|Classifier|Unary)' -count=1
go test -race ./ai/e2e -run '^Test(Decisions|JevRegression|Classifier|Unary)' -count=1
BARNESS_AI_PIDIFF=1 go test ./... -count=1 -json
go test -race ./... -count=1 -json
go vet ./...
go vet -tags live ./...
```

固定 Pi 依赖沿用 `ai/internal/testkit/pioracle/node/package-lock.json`；首次重放环境可运行
`npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node`。重放生成新证据，不覆盖固定包或既有失败。
原 Jev 的 native refusal 拒绝、宽 JSON、图例和有界舍入保持；其历史证据仍按原固定验证器独立核验。
