# 工单 04：Decisions 用量完整性与输入阶梯计费证据

公共 Classify / HookedClient.Classify 区分 unreported、partial 和 complete 全零报告，
Reasoning unset/0 独立。六项非负 int64 计数在 Decisions 边界验证；缓存合计、reasoning、total
矛盾或溢出使整份 usage 无效。唯一合法 usage 先于答案登记，坏答案仍保留 Result/Attempt 的输入估算。

起点 `17f2f070fa63678c310b75191dbb2ae012091468`，实现 `49f119c`，审查修复 `1e01c9a`，分支 main。
失败方式在编码前记录于[失败清单](../usage-failures.md)，沿用本票指定的公共接缝做 TDD。
其他既有工作区改动保留。本票不修改内置目录、Jev 价格、支持矩阵或其他协议 parser。

Input 保留 inclusive input_tokens；Output 如实报告，canonical cache 桶为零。
合成宿主目录的费率为 .1 USD/百万，严格超过 272000 后全请求输入用 .2，非零 output/cache 费率被排除。
费用示例：271999 → $0.0271999，272000 → $0.0272，272001 → $0.0544002。
同名聊天价格、ResponseModel 与构造后源目录修改均不会改变授权型号和 CatalogVersion/Hash 快照。
本票使用明确合成计数，未向厂商提交超长输入，也不推算图片 token。
全球标准处理估算未核对区域 premium 或真实账单；非零 cache 统计是合成协议证据，live 为 NOT_RUN。

| 验收 | 公共观察 |
| --- | --- |
| OD35 | usage 缺/null/空对象、六项逐个缺/null、details 缺/null、完整全零、Reasoning unset/0、ResponseModel 缺/alias |
| OD36 | usage/token/details 重复、大小写别名、非负整数类型/精度/int64 边界、两种 checked sum 溢出及 cache/reasoning/total 矛盾；无 last-wins 或无效前缀 |
| OD37 | inclusive Input、非零免费 Output/Reasoning、canonical cache=0、合法 reported total/已知项 checked sum |
| OD38 | 271999/272000/272001 × 无缓存/非零缓存/全缓存；全请求阶梯，四维身份与价格快照独立 |
| OD41 | 显式 retry=2 下 malformed JSON、坏答案、refusal、read abort、response callback error；一个 Attempt，终态、用量和资源释放 |
| R19 | 原始 presence fixture/capture、expected/actual Usage/Cost、Result/Attempt、metadata-only Observer、assertions、manifest 和脱敏审计 |

最终新增 307 个用量/生命周期 E2E 全部 PASS。
红色历史保留第一次 294 case 的 168 FAIL / 126 PASS；后来发现费用十进制 literal 与 IEEE 舍入之间
一个末位差异，测试只对数学费用允许 1e-15 USD 绝对误差，token/presence/free 桶仍精确检查。
生命周期补证据时 malformed 原始字节误命名为 .json 曾触发审计失败，已按实际内容使用 .txt；
未放宽审计规则。补充测试只验证既有共享 runtime 行为，没有新增 runtime。

两轴审查最终 Standards 0、Spec 0，详见 [REVIEW](REVIEW.md)。
完整测试、专项 race、两种 vet、编译、固定 Jev 655 case 的实际命令/退出码/统计记录在 validation.json。
固定 Jev 历史 release FAIL 保持，不用历史包核验冒称当前发布成功。
最终完整测试在 `1e01c9a` 退出 0：7027 E2E case，7018 PASS、9 既有压力 opt-in NOT_RUN，
815 个既有 Pi 差分全部 PASS。最终专项 race 开启 Pi，1258 PASS；两种 vet 和编译检查均退出 0。
审查补 OD41 后重新跑完整测试；源码快照与最终完整/race 的实现提交一致。

usage.zip 保存初次公共红色捕获、最终 307 case 和专项 race 的完整捕获、测试/转换源码、
全套 manifest/audit/进程日志、源码 SHA-256 快照和 validation 收据。原始 presence 在 response fixture
与脚本/HTTP capture 中保留；不新增 Observer 明细。外层 manifest 记录每个 ZIP 成员的 SHA-256，
SHA256SUMS 固定 ZIP、manifest、验证收据、README 和 REVIEW。全套逐例捕获留在运行产生的 .evidence，
提交包只保存全套汇总及本票/专项的完整 captures。

核验固定包（不请求厂商）：

```sh
(cd .scratch/barness-ai-openai-decisions/usage-evidence && shasum -a 256 -c SHA256SUMS)
unzip -t .scratch/barness-ai-openai-decisions/usage-evidence/usage.zip
python3 .scratch/barness-ai-openai-decisions/jev-baseline/verify.py --audit
```

重放当前源码（仓库根目录）；每 case 的 assertions/manifest 还保存逐例准确命令。
原 Pi 依赖可由 `npm ci --ignore-scripts --prefix ai/internal/testkit/pioracle/node` 恢复。
重放只生成新包，不改写固定历史：

```sh
go test ./ai/e2e -run '^TestDecisionsUsage' -count=1
BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^Test(Decisions|JevRegression|Classifier|Unary|OperationCatalog)' -count=1
go vet ./...
go vet -tags live ./...
BARNESS_AI_PIDIFF=1 go test ./... -count=1 -json
```
