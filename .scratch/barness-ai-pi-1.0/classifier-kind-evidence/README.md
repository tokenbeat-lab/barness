# 分类问题类型优化验证记录

2026-10-08，维护者要求“请按建议优化”。`ClassifierCapabilities.Kinds` 改为
`[]ClassifierQuestionKind`，提供 `ClassifierQuestionChoice`、`ClassifierQuestionScore`、
`ClassifierQuestionBool` 常量；目录校验与宿主 E2E 构造使用同一类型。
前轮审查的 Primitive Obsession 建议已落实；详细实际运行路径和哈希见 [机器报告](report.json)。

| 检查 | 结果 |
| --- | --- |
| `go vet ./...`；`go vet -tags live ./ai/...` | PASS |
| 定向公共目录 E2E | 60 PASS；审计 0 findings |
| 全量 Go 测试套件 | PASS；E2E 2817 PASS、619 NOT_RUN、0 FAIL；审计 0 findings |
| 定向 race：操作授权与目录 | 91 PASS；审计 0 findings |
| 内置目录重新生成后逐字节比较 | PASS |
| 三类宿主目录内容与原记录哈希核对 | PASS |

全量运行中的 619 NOT_RUN 为未启用的 pi 差分、压力及冻结目录比较。这些场景已在
[工单 04 完整验收](../operation-catalog-evidence/README.md)通过，本轮只改变 Go 类型，JSON
与目录内容不变；未重复运行这些 opt-in 场景或 live 探针。

## 公共契约

宿主改用专用切片与常量：

```go
Kinds: []ai.ClassifierQuestionKind{
    ai.ClassifierQuestionChoice,
    ai.ClassifierQuestionScore,
    ai.ClassifierQuestionBool,
},
```

已有 `[]string` 需要改写为此类型；动态字符串逐项转换后仍通过 NewClient 校验。
不增加兼容别名或 shim。未知字符串、空值、重复类别及不一致范围仍拒绝，类型不代替边界校验。
底层 string 保持 JSON 为 `choice`、`score`、`bool`；内置目录与宿主三类目录的内容/哈希均不变，
目录版本保持 `2026-10-08.2`。公共契约与 ADR-0020 记录本次经维护者授权的源码类型调整。

## 先红后绿

先把既有公共目录 E2E 改为专用类型与常量装配，并加入空类别失败场景、JSON 字符串黄金值与
目录 JSON 往返检查。公共声明尚不存在时编译失败，原始输出保留于 [编译失败记录](compile-red.log)。
随后实现类型、常量和类型化校验，定向 E2E 通过。没有新增隔离单元测试。

## Standards

0 硬性违规、0 可选建议。原 Primitive Obsession 建议消除；边界校验、字符串编码与复制保留，
未加兼容 shim；文档明确 Go 切片迁移，变更文件小于 500 行。审查使用本轮开始前
`0c10651361424272dad1a0d1643bcad382bef2c9` 到本轮暂存变更；最终已补齐审查时尚在整理的本记录。

## Spec

0 findings。已实现维护者要求的专用类型与三个常量，校验与调用方同步更新；JSON、哈希、
深复制与拒绝行为保持，文档记录明确授权，无无关 API 或运行时工作。

两轴结论：Standards 0 发现；Spec 0 发现。

## 重放

```sh
go vet ./...
go vet -tags live ./ai/...
go test ./ai/e2e -count=1 -run '^(TestOperationCatalogValidation|TestOperationCatalogDiscovery|TestOperationCatalogHash)$'
go test ./... -count=1
go test -race ./ai/e2e -count=1 -run '^(TestChatOperationAuthorization|TestOperation.*)$'
go run ./ai/release/cmd/releasegate -write-snapshot -snapshot /tmp/barness-kind-catalog.json
cmp ai/release/catalog-snapshot.json /tmp/barness-kind-catalog.json
go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/classifier-kind-evidence
```

E2E bundle 含公共目录快照、结果、断言、审计与精确 replay。机器报告固定本次 manifest 的 SHA-256；
本目录 manifest 固定报告、摘要、编译失败与审计记录的 SHA-256。前轮验证记录保留为历史证据。
