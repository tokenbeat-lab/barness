---
status: accepted
date: 2026-10-02
---

# barness-ai 发布门禁：纯函数判定、按设计负载的策略压力场景、写出后脱敏审计与只由证据判定的追溯

工单 24 交付发布门禁（spec I10 后两条、Testing Decisions §6）。spec 规定了门禁项（`go test ./...`、`-race`、`go vet`、全部 P0、无待处理差分、六组合 live PASS 或明确 UNSUPPORTED）、证据要求（离线/差分/live 分开报告，研究条目 → 场景 → 证据可追溯，“已映射”不得填为 PASS）和“正式示例随实现提供已跑通压力场景的有限配置及数值依据”，没有规定门禁的形态、压力场景的负载如何定义、脱敏审计检查什么、追溯如何计算。本 ADR 记录实现时的做法，待维护者确认。

## 决策一：门禁是证据上的纯函数，由一个命令收集输入

`ai/internal/testkit/release.Evaluate(Inputs) Report` 只根据输入判定：命令结果、离线证据包（manifest 中每个用例的状态与每个差分用例 `pidiff.json` 的待处理数）、差分账本摘要、支持矩阵、追溯表、审计结果与快照检查。命令 `go run ./ai/release/cmd/releasegate` 负责执行 `go vet`（含/不含 `live` tag）、离线 E2E（差分与压力开关打开）与 `-race`，把证据写在发布证据包下，然后判定并写出报告。

live 门禁对每个组合要求：矩阵中有完整通过、当前无失败能力、最后一次完整通过记录的 SDK 版本与 go.mod 当前版本一致（否则视为 NOT_RUN，需要重跑），且该组合的 live 证据包经 `-live` 交给门禁审计。adapter 改动无法从矩阵判断，仍需维护者按 spec §6 重跑受影响组合；目录哈希未进入矩阵行，同样依赖重跑。差分门禁另核对每条差分记录含 spec §6 要求的字段（case_id、pi_commit、sdk_versions、model_catalog_hash、请求与帧哈希、发现列表）。

判定逻辑的错误会让不该发布的版本通过，所以按项目测试原则在隔离中先列出失败方式再实现（现为 21 种）（`gate_test.go`），并做了变异检查。live 门禁只读支持矩阵，不在门禁里调用真实 API：真实冒烟一进程一组合、只持有该组合的 key（ADR-0016），不能由一个进程统一执行。

### Considered Options

- 用 shell 脚本串联命令与 jq 判定：判定规则无法测试，“未执行即失败”之类的边界易遗漏。
- 把门禁写成一个 E2E 测试：它需要读取同一次 `go test` 尚未写完的证据包，且 `-race` 与 vet 结果无从获得。
- 纯函数判定 + 收集输入的命令（采用）。

## 决策二：策略压力场景按示例声明的设计负载运行，默认不运行

`E08-policy-pressure-<策略>-design-load` 以示例注释声明的负载运行：进程并发占满（按租户上限分摊到多个租户），每个调用发送设计大小的请求（长历史加图片），经 Responses 流式返回设计大小的单轮输出（每 token 一帧，带厂商实际发送的 logprobs/obfuscation 字段）或每四个调用一个设计大小的工具调用。断言每轮完整到达、并发达到上限、每项字节与队列限额用量不超过 75%、进程可达堆增长不超过策略假定的预算（本地 1 GiB、云端 2 GiB），并记录每项限额的 headroom、堆与吞吐（`policy-pressure.json`）。`-stalled-consumer` 以同一设计输出、不读事件的 Stream 证明以 `resource_limit`（event_queue）结束，记录按 100 token/s 估算的积压时长。

选 Responses 因为其终态帧重复整段响应，同等输出下最接近 MaxFrameBytes 与 MaxOutputBytes。堆用 `runtime/metrics` 的 `/gc/heap/live:bytes`（GC 标记的可达堆），不用 HeapInuse（含未回收垃圾、随 GOGC 变化）；它包含测试 Provider 替身保存的请求副本，报告单列该数，偏保守。替身回复不保存整段响应的证据副本（每轮数十 MB 会主导被测堆），证据记录生成参数。

一次运行约 30 秒、数百万帧，不适合每次 `go test ./...`；与差分相同，需 `BARNESS_AI_PRESSURE=1` 打开，未打开时用例为 NOT_RUN，门禁的 P0 检查因此失败。门禁的离线运行打开它，`-race` 运行不打开。

75% 的上限是本决策的取值：厂商帧信封（obfuscation 填充、logprobs、推理摘要）会变化，设计负载接近填满某项限额的策略会拒绝普通请求。

## 决策三：压力数据修正示例数值

- `MaxToolJSONBytes`：云端 512 KiB → 128 KiB，本地 1 MiB → 256 KiB。流式工具参数在每个 delta 上全量重解析（pi `parseStreamingJson` 的移植，pi 同样如此），单调用 64 KiB 0.4 s、128 KiB 1.6 s、256 KiB 6 s、512 KiB 23 s；限额实际约束的是 CPU。工单 32 跟踪增量解析，修复后按内存预算放宽。
- `CloudBatchPolicy.MaxOutputBytes`：32 MiB → 64 MiB。128K token 单轮约 29 MB，占原值 88%；流式响应体只读不留，加倍几乎不增内存（实测可达堆 +0.9–1.2 GiB，32 并发）。
- 云端注释中请求体的最坏内存由 `(并发 + 等待者) × MaxRequestBytes` 改为其两倍：实测一次尝试期间约持有请求体两份副本（4.9 MB 请求约 10 MB）。

## 决策四：脱敏审计检查已写出的证据包，并令运行失败

`ai/internal/testkit/audit` 逐文件检查证据包（规则见包文档）：已注册秘密；审计进程环境中名称像凭据的变量值、主目录与主机名（出现即说明记录了真实而非合成的数据）；OpenAI/DeepSeek/Anthropic/Google key 形状；JSON 中请求凭据头（Authorization、x-api-key、x-goog-api-key、api-key、cookie、proxy-authorization）的值只能是脱敏标记或测试别名；自由文本中的 bearer token；URL 的 key 参数；Observer 记录（`observations*.json`）只能含 Observer 元数据字段（白名单，新增字段须审阅后加入）；live 捕获的响应头值只能属于记录器白名单。结论只给文件、规则与位置，不复述值。

审计在每次离线 E2E 与 live 运行结束时执行（`evidence.Run.Audit`，写 `audit.json`，有发现即令进程失败；取代 live 原先只查 key 的检查），门禁再对离线、race 与 `-live` 给出的证据包复查，并要求每个包的 `audit.json` 存在且零发现：只有运行进程知道它注册的测试秘密，门禁的复查只能依赖形状与环境规则。找不到 race 证据包时记为审计错误，不少审一个包。响应的 Set-Cookie 不按凭据头检查：它由服务端签发，离线 fixture 有意设置合成值（cookie jar 加固用例），live 捕获本就不保留其值。

### Considered Options

- 只在来源处脱敏（现状）：证明不了写出的内容，变异检查显示关闭来源脱敏后无任何失败。
- 维护“合成输入”清单并要求请求正文只含其中内容：只是用套件验证套件自身；以环境值与凭据形状作为“非合成数据”的可检查替代（采用）。

## 决策五：追溯只由证据判定

`ai/release/traceability.json` 把研究条目（T01–T12、C01–C11、V§4 各协议、V§6、H1–H5、D1、D2、正式示例、S§6）映射到场景（E01–E11、P01–P06、D1、D2、H1、PRESSURE、PIDIFF，各为用例 ID 的正则），条目可再以正则收窄到具体用例；保留场景 `LIVE` 取支持矩阵。条目状态由匹配到的证据计算：有 FAIL 为 FAIL，有 NOT_RUN 为 NOT_RUN，没有 PASS（无证据或只有 UNSUPPORTED）为 NO_EVIDENCE，否则 PASS；离线、差分与 live 证据分别计数。映射文件只表达“哪些证据证明哪条需求”，从不携带状态。

## 决策六：目录快照入库并由门禁核对

`ai/release/catalog-snapshot.json` 是内置目录及其哈希的入库快照（`-write-snapshot` 生成）；门禁在它与代码不一致时失败，使价格与模型变更在评审中以文件差异出现。fixture 不另建入库索引（它们本身入库），门禁在发布证据包中写出带哈希的 `fixtures.json`。

## Consequences

- 2026-10-02 六组合真实冒烟完成并合并后，门禁全部通过（`.evidence/barness-ai-release/gate4`）。
- 门禁一次约需离线约 2.5 分钟加 race 运行时间；压力场景只在门禁或显式打开时运行，日常 `go test ./...` 不变。
- Observer 增加字段时，须同时更新审计白名单，否则 E2E 失败——这是有意的审阅点。
- 压力场景测得的堆与吞吐依赖机器；数值依据记录的是一次运行，门禁每次重测并以 75% 与预算为判据，不以精确值为判据。

## 待维护者确认

1. 门禁形态（纯函数 + 命令）与九道门禁的组成。
2. 压力场景默认不运行、由门禁打开；75% headroom 上限；本地 1 GiB 的内存预算假定。
3. 三处示例数值修正，尤其 `MaxToolJSONBytes` 的收紧（在工单 32 之前）。
4. 脱敏审计规则，尤其以环境值作为“非合成数据”的判据、Set-Cookie 的例外。
5. 追溯表的条目与收窄正则是否足够细；目录快照入库。

## 维护者决定（2026-10-03）

以上五点全部采纳。`MaxToolJSONBytes` 的收紧是临时措施：工单 32 使流式工具参数解析成本近似线性后，按内存预算放宽示例值并更新压力场景的设计负载。

## 后续（2026-10-04，工单 32）

决策三中 `MaxToolJSONBytes` 的收紧已撤销：云端 128 KiB → 512 KiB，本地 256 KiB → 1 MiB（即原值）。两处平方成本都已消除：`PartialView` 不再在每个 delta 上解析工具参数，而是在被读取（Snapshot、序列化、终态）时按当时的原始文本解析并缓存到下一个 delta，读到的值与 pi 在该时刻持有的相同（差分无新增差异）；适配器累计原始文本时不再 `raw += delta` 整段复制。新增 `E08-policy-pressure-tool-arguments-scaling` 以 64 字节 delta 流式传入 128 KiB、512 KiB、2 MiB 的单个工具调用，要求最大与最小尺寸的每字节耗时之比不超过 3（平方增长时为 16），实测约 1.0（12 ms、48 ms、192 ms）。压力场景的设计负载相应改为本地 512 KiB、云端 256 KiB 工具调用（用量 50%）。工具调用现在数十毫秒即结束，可能在最后一个调用被准入前释放许可（一次运行测得峰值 31/32），因此设计负载的每个回复都等到全部请求到达后才开始流式返回，“并发占满”不再依赖时序。仍按 pi 成本的情形：消费者在每个 delta 后都读取视图时，每次读取都完整解析一次参数。

## 工单 08：操作身份与分类发布输入（2026-10-08）

当前 liveCombos 增至七条，新增 TypeSafe classifier。live 门禁和输出报告携带
operation；schema 或 operation/provider/API 与固定组合身份不一致时拒绝复用历史通过。
六条聊天历史只迁移 schema/operation 元数据，不据 TypeSafe 的通过更新其他行。
本工单只交付 TypeSafe 当前路线的 live 证据；旧六条尚未在 1.0 升级后重新跑过，
不能把本次专项验收称作整个版本发布通过。后续图像路线仍由其工单分别验收。

审计复核：operation 是入口常量；providerRequestId 是厂商 ID 元数据；
x-typesafe-request-id 属于现有 request-id 响应头白名单。无需扩大白名单到正文；
真实 observations 文件与错误响应一起经过逐文件审计，禁止 state/questions/answers/错误正文进入 Observer。

## 工单 14：Google 图像真实证据（2026-10-08）

追溯增加工单 14 的真实图像/目录/省略 delivery 条目，liveCombos 纳入第九组合 google-interactions-image。该路线的离线回放证明归一化，真实支持另需自己的已审计 live bundle；不能把回放或共享 adapter 当作 live PASS。此工单不宣称九组合全发布门禁通过（后续工单 17）。
