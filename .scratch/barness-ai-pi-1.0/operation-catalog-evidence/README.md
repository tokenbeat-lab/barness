# 工单 04 验证记录

2026-10-08。模型操作、三类目录与 Binding 授权已实现。最终判定及本次实际证据路径见
[机器报告](report.json)；三类宿主目录见 [捕获快照](host-catalog-snapshot.json)。

| 检查 | 结果 |
| --- | --- |
| `go vet ./...`；`go vet -tags live ./ai/...` | PASS |
| 全量离线 E2E（含六条聊天路线、冻结 pi 差分与压力） | 3435 PASS；新增操作目录/授权 90 PASS |
| 冻结 pi 1.0.0 差分 | 611 PASS、0 FAIL、0 pending |
| policy pressure | 7 PASS |
| race | 2816 PASS、619 NOT_RUN、0 FAIL；脱敏审计 0 findings |
| 目录快照重新生成与逐字节比较 | PASS |
| 全量 E2E 脱敏审计 | PASS，0 findings |

race 的 619 NOT_RUN 是未开启的差分、压力与冻结模型数据比较；它们已在普通全量运行中通过。
本次未运行 live 探针，不新增未验收的内置图像/分类型号，也不宣称新的厂商调用路线已支持。

## 范围与兼容

- `Operation` 只取 chat、image、classifier；Binding 零值解释为已有的聊天授权，未知值在 binding
  阶段拒绝，聊天跨操作误用在 capability 阶段拒绝。拒绝前不读凭据、不发请求、不创建 Attempt。
- `Model` 和 `Target` 保持原字段与聊天含义；新增两种独立型号描述，完整身份包含操作、Provider、
  API 与 ID。宿主可同 ID 跨操作存放；客户端构造拒绝重复完整键、矛盾能力与模态、非法分类范围、
  非有限/负价格、声明模态缺费率和聊天默认 samplingParams 保留字段。
- `Client.Catalog` 与三类 Lookup/列表返回完整独立副本，测试修改原配置、查询副本及其嵌套内容，
  再通过真实公共聊天入口核对请求、价格和三类目录哈希。同名跨租户绑定并发调用及轮换冲突均验证。
- Operation 在 CallStarted、事件归属和预检失败时即为 chat；服务归属与版本仍等待一致快照。
  新字段经 Observer 白名单审阅；秘密、正文、答案未进入观测。
- 内置目录升为 `2026-10-08.2`，聊天条目和价格不变；新增目录类型为空时省略 JSON 字段。
  三类宿主快照中的 image/classifier 是合成测试元数据，其 Provider/API 故意与 chat 相同以证明操作隔离。
- ADR-0020、ADR-0001/0018、词汇表与公共契约同步。反向 Classify/GenerateImages 拒绝由各自实际
  调用切片验证；此工单未引入空 adapter 或提前抽取仅一个调用方的共同运行时。

## 先红后绿

公共测试边界由 issue 04 指定为目录构造/发现和四个聊天入口；测试文件先列出所有失败场景。
Operation/Binding/归属的公共声明尚不存在时编译失败；补齐声明后，操作授权组的 20 个场景全部失败。
其证据审计另外产生 56 个 operation 未进 Observer 白名单的预期发现。加入授权行为、入口归属及
经审阅的枚举白名单后通过。目录组在类型已声明、验证尚未实现时为 2 PASS / 49 FAIL；深复制尚未
保留新增集合时仍失败，之后通过。发现 API 尚不存在时也先编译失败，再实现强类型发现与深复制。
原始失败 bundle 与 manifest 哈希保留在机器报告，不清理或隐藏失败记录。

## Standards

审查基点为本轮开始前 `f99ced9bbf0c14254acf0a50d158b86bf72419f8`，范围是工单 04 的暂存变更。
无硬性标准违规：操作授权早于凭据读取，租户隔离保持，查询副本包含独立嵌套数据，变更源码均小于
500 行，零值例外有 ADR 和公共契约说明，无内部兼容 shim。

一项非阻断判断：可能的 Primitive Obsession，`ClassifierCapabilities.Kinds []string` 使用封闭的
choice/score/bool 字符串，类型常量可以减少拼写错误。当前按已批准设计第 6.1 节保留此公开形状，
构造时拒绝错误值；不将可选建议扩成当前工单的公共类型变更。相似的强类型查询方法由独立模型契约
与避免过早抽象的项目原则支持。

## Spec

0 actionable findings。操作身份、类型发现、目录校验与复制、凭据前授权、提前归属、旧绑定规范化、
快照、E2E 和文档均覆盖；未发现范围扩张。后续 unary 运行时与厂商型号列入保持在后续工单。

两轴结论：Standards 0 硬性违规、1 可选建议；Spec 0 发现。

## 重复验证

在仓库根目录运行：

```sh
go vet ./...
go vet -tags live ./ai/...
go test ./ai/e2e -count=1 -run '^(TestChatOperationAuthorization|TestOperation.*)$'
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test ./... -count=1
go test -race ./... -count=1
go run ./ai/release/cmd/releasegate -write-snapshot -snapshot /tmp/barness-04-catalog-snapshot.json
cmp ai/release/catalog-snapshot.json /tmp/barness-04-catalog-snapshot.json
go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/operation-catalog-evidence
```

每个 E2E 产出 `.evidence/barness-ai/<time>/`，包含快照、请求/响应、断言、审计与精确 replay 命令。
本目录 manifest 固定摘要、报告和合成宿主快照的 SHA-256；机器报告固定完整运行与红阶段 bundle
的 manifest SHA-256。运行 bundle 留在本地，源码提交保存可审阅摘要和可重复命令。
