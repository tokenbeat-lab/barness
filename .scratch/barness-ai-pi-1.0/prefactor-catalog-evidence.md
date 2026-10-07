# 工单 01：目录预重构验证

## 实现前的回归清单与接缝

本工单保留冻结 pi-ai 0.87.1、目录版本 `2026-10-03.2` 与内容哈希
`sha256:13553fe428d710dab86563bdb865f44554b88e8637a83b206c922701a3c9aca4`。
不实施工单 04 的新类型、索引、重复键校验或新哈希输入。

测试接缝采用工单和规格已经指定的公共 `ai.Client` → 本地受控 Provider / 宿主替身，
以及现有目录包含验证，不测试私有构造函数或文件组织。

| 可能回归 | 行为证明 |
| --- | --- |
| 拆分时型号遗漏、顺序变化、JSON 形状改变 | `TestUsageAndCost/builtin-catalog-pinned` 的固定版本和哈希；`TestCatalogInclusion/builtin-models-are-pi's` 逐字段比对冻结数据 |
| Provider / API / ID 查询错配、白名单含义改变 | `TestClientRouting`、`TestPreflightRejections` 与现有 DeepSeek 路由、目录场景 |
| 价格、阶梯阈值、兼容字段或请求行为漂移 | `TestUsageAndCost`、`TestCatalogInclusion` 与冻结 pi 差分 |
| 构造后修改目录或 Policy 泄漏到 Client | `TestClientRouting/config-read-only-after-construction`，补充 `TestCatalogSnapshot` 从请求图片、采样参数、实际计价与元数据观察嵌套深复制 |

纯文件拆分应在重构前后均通过这些行为断言；不存在需要新增的运行时行为。

## 职责整理

原 `ai/catalog.go` 的声明、函数及其注释原文搬移，只有 import 按各文件依赖调整：

| 文件 | 职责 | 行数 |
| --- | --- | --- |
| `ai/catalog.go` | `Catalog`、`BuiltinCatalog` 装配及 Provider / API / ID 查询 | 69 |
| `ai/catalog_models.go` | `Modality`、`Model`、`ModelCompat` 及模型能力判断 | 84 |
| `ai/catalog_builtin.go` | 已有六条协议路线的内建型号数据与数据构造辅助函数 | 286 |
| `ai/catalog_snapshot.go` | 目录深复制、JSON 内容哈希与价格校验 | 45 |

查询仍按完整聊天键取首个匹配；构造和快照保持现有规则。`Model` 字段顺序、JSON 标签、
`Catalog` 的公开字段及 keyed struct literal 形式不变。未新增依赖或抽象。
`ai/README.md` 同步维护入口位置。此整理贯彻已有 ADR-0010 与 ADR-0018，不改变架构决策。

## 验证与可重复运行

重构前先运行补充 E2E 和既有目录、价格、路由场景。重构后编译全部 `ai` 包并重跑受影响场景。
两次均启用冻结 oracle，目录包含验证实际执行。新增测试只在公共 Client 接缝观察行为，
期望价格来自冻结 `text-basic.json`，请求期望独立写出，不以实现计算预期值。

| 验证 | 结果 |
| --- | --- |
| 全部 `ai` 包编译 | `go test ./ai/... -run '^$'` PASS |
| 重构前公共入口与目录验证 | 184 个场景 PASS（另一次更窄运行 83 个 PASS） |
| 重构后受影响公共入口与目录验证 | 180 个场景 PASS |
| 完整 `go test ./... -count=1`（差分与压力启用） | 2877 个证据场景全部 PASS，含 464 个差分场景 |
| 冻结 pi 差分 | 464 PASS、0 pending；沿用原账本，无放宽登记 |
| `go vet ./...`、`go vet -tags live ./...` | 均 PASS |
| 完整离线包的脱敏审计 | `audit.json`：`findings: []` |
| 发布目录快照 | 在证据目录重新生成，与已提交快照逐字节一致 |

工具链为 Go `go1.26.2`；SDK 为 `openai-go v3.66.0`、`anthropic-sdk-go v1.75.0`。
oracle 保持 pi-ai `0.87.1` / commit `898ab804050730e9dcefb4443875d5a932aa6a32`。
目录版本和哈希仍为本记录开头的固定值。

- [完整验证汇总](../../.evidence/barness-ai-pi-1.0/01-final/verification-summary.json)
- [完整离线 manifest](../../.evidence/barness-ai-pi-1.0/01-final/offline/20261007T141327.299485000Z/manifest.json)
- [脱敏审计](../../.evidence/barness-ai-pi-1.0/01-final/offline/20261007T141327.299485000Z/audit.json)
- [搬移与文件行数核对](../../.evidence/barness-ai-pi-1.0/01-final/relocation-check.json)

证据文件按仓库惯例保留在 git-ignored `.evidence/`，以下命令可重新产生。
本工单只改目录组织，不触及协议 adapter；验证范围为离线行为和冻结基线。

从仓库根目录重复执行（Node 依赖已按现有 lockfile 安装）：

```sh
mkdir -p .evidence/barness-ai-pi-1.0/01-replay
export BARNESS_AI_EVIDENCE_DIR="$PWD/.evidence/barness-ai-pi-1.0/01-replay"
BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test ./... -count=1
go vet ./...
go vet -tags live ./...
go run ./ai/release/cmd/releasegate -write-snapshot \
  -snapshot .evidence/barness-ai-pi-1.0/01-replay/catalog-snapshot.json
cmp ai/release/catalog-snapshot.json .evidence/barness-ai-pi-1.0/01-replay/catalog-snapshot.json
```

每个 E2E 包内的 `manifest.json` 记录版本、目录哈希、fixture 哈希和逐场景重放命令，
`audit.json` 给出运行结束时的脱敏审计结论，`PIDIFF-*/pidiff.json` 给出差分分类。

## 代码审查

按 `code-review` 技能以工作开始时的 `fdb075b5858966fbefd455d92386e2092f5a146c`
为固定比较点，两个独立子代理并行进行只读审查。

### Standards

零 findings：未发现文档标准违反，也未发现需要处理的 baseline smell。

职责拆分符合 `AGENTS.md` 原则 1、2、7、10：保留现有声明和行为，删除原位置实现，
未新增兼容层、依赖或推测性抽象。公开模型、查询身份及价格快照语义与 `GLOSSARY.md`、
ADR-0010、ADR-0018 一致。

新增测试通过公共 `Client` 观察嵌套配置快照，使用独立请求预期和冻结成本 fixture，
并产出可重放证据，符合仓库 E2E 原则。README 的维护入口也已同步。
DeepSeek 两种协议的数据重复保留了独立演进理由，符合原则 7，不建议抽象。

### Spec

零 findings。工单 01 的行为保持拆分已完成：模型元数据、内建数据、目录查询及快照职责
分到均低于 500 行的文件。现有 `Model`/`Catalog` 字段、keyed literal、查询行为、
内建内容、价格、版本、JSON 形状与哈希均不变。

补充公共 Client E2E 通过请求行为、计价及元数据观察嵌套配置修改。
证据包含完整离线测试、0 pending 的冻结基线差分、两种 vet、空审计发现及逐字节一致的发布快照。
未越界实施后续模型操作、校验、索引或 pi 1.0 迁移。

审查汇总：Standards 0 项，Spec 0 项；两轴均无待修复事项。
