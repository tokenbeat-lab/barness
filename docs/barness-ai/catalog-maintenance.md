# 本地模型目录维护与加载

`catalogctl` 管理已审核的目录来源、冻结候选和本地生效版本。`catalog` 包在 Host 装配时加载选定的本地 JSON，返回有效 `ai.Catalog` 与独立报告。运行时不访问 models.dev，不刷新现有 Client，也不自动选择其他目录。

目录接纳、当前代码接受、厂商实测、模型健康和 Binding 授权是不同事实。`check` 的结论只有“代码兼容通过”；`apply` 不发布线上目录、不更新支持矩阵，也不表示完整发布门禁通过。

本文是 Catalog CLI 的操作 SOP。每次目录变更必须完成候选/check/apply、版本化
快照归集和 changelog；每次同步内置目录还必须给出核心运行代码变化及分发判定。
缺少这些交付项时，不能将“apply 成功”写成“同步交付完成”。

## 统一产出目录与版本化快照

统一根目录为仓库根的 `artifacts/catalog/`。使用真实 `Catalog.Version` 和唯一
UTC 批次组织每次交付；不能以型号名、日期或软件版本代替目录内容版本。

```text
artifacts/catalog/
  work/store/                          # 显式选择的本地维护库
  work/<run-id>/                       # 输入、工具、命令输出和失败操作记录
  <Catalog.Version>/<run-id>/
    <Catalog.Version>.json             # 原受测运行 JSON，可独立分发
    <Catalog.Version>.json.sha256      # 完整文件摘要
    review/                            # CLI 导出的完整审阅包，原字节
    CHANGELOG.md                       # 本次变更说明快照
    delivery.md                        # 人可读代码变化及分发判定
    verification/                     # 验证摘要、检查记录与原件索引
    delivery.json                     # 最后写入的完成索引与全部文件摘要
```

运行 JSON 内仍使用原 `{hash,catalog}` 格式，文件名必须为
`<Catalog.Version>.json`，其中版本等于文件内 `catalog.version`。不增加自定义字段，
不重新序列化或修改原受测字节。目录 Hash 与完整 JSON 文件 SHA-256 分别记录。
版本应是生成器产生的安全单一路径段；不满足时停下处理，不静默改名或编造版本。

所有交付 JSON 均位于该版本/批次的不可变快照目录内。`review/` 的 `state.json`、
`snapshot.json` 等固定文件名属于 CLI 原契约，保持原文和库存，不单独重命名。
同一目录版本可能有新的追溯 revision 或检查；使用新批次，不覆盖既有快照或索引。

原始测试胶囊继续按测试 SOP 保留在 `.evidence/<feature>/...`，历史工作库及原输入
也保持原路径。统一交付目录归集可分发文件、完整审阅来源及验证索引；索引必须
记录外部原件的原路径、SHA-256、真实状态和退出，并明确回放是否依赖外部原件。
不能移动原件破坏引用，不能把“已有摘要索引”说成“完整证据已包含在交付包内”。

`ai/catalog_builtin.json`、`catalog/data/providers/` 和 `catalog/data/baseline/` 是
源码采用位置，仍按下文同步；交付归集不改变这些编译路径。现有 `local-catalog.json`
只保留为历史原件，不再作为新交付的无版本文件名。目录索引见
[统一交付入口](../../artifacts/catalog/README.md)。

## Host 选择与装配

Host 先准备已有的凭据解析、Binding 和资源策略 `cfg`，再明确选择来源：

```go
import (
    "github.com/tokenbeat-lab/barness/ai"
    "github.com/tokenbeat-lab/barness/catalog"
)

// cfg is the Host's existing ai.Config, including its resolvers and policy.
// snapshotPath is the explicitly selected versioned JSON from a delivery.
effective, report, err := catalog.LoadFile(snapshotPath)
if err != nil {
    return err // Host decides how to handle this file-level failure.
}
cfg.Catalog = &effective
client, err := ai.NewClient(cfg)
// Keep report separately for trusted Host diagnostics.
```

`catalog.Load(raw)` 接受同一格式的内存字节。选择内置数据时使用 `ai.BuiltinCatalog()`；既有 `Config.Catalog` 注入方式和模型查询接口保持。

- 文件无效 JSON、顶层/集合结构错误、版本非法或原始摘要不符时，整份拒绝，不返回可用部分，不回退。
- 有效文件按三集合的原顺序逐模型严格加载。未知字段、非法定义、明确未支持项只使该条失败；同一完整四维身份的全部重复项均失败。
- 零条成功仍返回完整报告和空有效集合。将其作为非 nil 的 `Config.Catalog` 注入，避免 Host 自行把空集合当成未选来源。
- Client 固定装配时的目录和价格。修改文件、报告或查询副本不会影响现有 Client、在途调用或其重试；调用错误不反写加载状态。

报告字段为 `SourceVersion`、`SourceHash`、`EffectiveHash`、`Loaded`、`Failed`、`Models`。每条状态包含完整或部分 `Identity`、`Collection`、从零开始的 `Index`、`loaded/failed`、原因、诊断和建议。失败 `Raw` 仅供可信 Host 诊断，不应直接写日志或对外返回。

模型失败原因区分 `unknown_field`、`unsupported_definition`、`invalid_definition`、`duplicate_identity`；文件失败通过 `*catalog.FileError` 的独立 `Reason` 表达。升级只是一种排查建议，不推断最低程序版本。

运行文件仍是 `{ "hash": "sha256:...", "catalog": { ... } }`。外层 hash 在模型解析前校验完整原始 catalog 的紧凑字节，只去除字符串外空白，不改属性顺序、数字或转义；未知模型字段同样参与校验。有效 `Catalog.Hash()` 继续标识当前接受的 typed 集合，可能与来源摘要不同。来源 Version 保持，调用元数据使用有效 Version/Hash。

## 维护入口与候选闭环

在仓库根编译 CLI，查看某入口的精确参数：

```sh
ARTIFACT_ROOT="$PWD/artifacts/catalog"
RUN_ID='<实际UTC时间>-<唯一批次>'
STORE="$ARTIFACT_ROOT/work/store"
WORK="$ARTIFACT_ROOT/work/$RUN_ID"
mkdir -p "$WORK/bin"
CATALOGCTL="$WORK/bin/catalogctl"
go build -o "$CATALOGCTL" ./cmd/catalogctl
"$CATALOGCTL" help
"$CATALOGCTL" add --help
```

这些路径是维护者显式传入的 SOP 约定，CLI 默认参数没有被修改。工作库首次迁移
须复制完整原库、逐文件核对路径集合及摘要后再选择；原库保留，不能用新种子
冒充原历史。CLI 本身的查询/help/dry-run 不写库，以上目录由操作者明确创建。
命令 stdout/stderr、实际 argv、退出和工具身份归入本批次 `WORK`，失败同样保留。

输出默认是 JSON；`--json` 可显式声明。命令失败返回稳定 `category` 和非零退出。查询、help 和 `--dry-run` 不创建维护库或证据目录。变更缺少 `--confirm` 时返回 `confirmation_required`，不会隐式读取 stdin 或自动接受变更。

以下用已有型号的显式下架演示完整闭环。示例中的候选 ID 和检查引用必须使用前一条命令真实返回的值：

```sh
"$CATALOGCTL" withdraw --store "$STORE" \
  --operation chat --provider openai --api openai-responses --model gpt-4 \
  --reason '维护者确认的下架原因' --dry-run

"$CATALOGCTL" withdraw --store "$STORE" \
  --operation chat --provider openai --api openai-responses --model gpt-4 \
  --reason '维护者确认的下架原因' --confirm

"$CATALOGCTL" check --store "$STORE" --id '<candidate-id>' --repo . \
  --evidence-feature catalog-review --timeout 5m --confirm

CATALOG_VERSION='<冻结候选 snapshot 中的真实 catalog.version>'
DELIVERY="$ARTIFACT_ROOT/$CATALOG_VERSION/$RUN_ID"
SNAPSHOT="$DELIVERY/$CATALOG_VERSION.json"
mkdir -p "$DELIVERY"

"$CATALOGCTL" apply --store "$STORE" --id '<candidate-id>' --repo . \
  --check '.evidence/catalog-review/check-inputs/<attempt>/check.json' \
  --out "$SNAPSHOT" --dry-run
"$CATALOGCTL" apply --store "$STORE" --id '<candidate-id>' --repo . \
  --check '.evidence/catalog-review/check-inputs/<attempt>/check.json' \
  --out "$SNAPSHOT" --confirm
```

上述版本必须从已冻结候选取得；CLI 的 `candidate.snapshot` 是 Base64 原字节，
读取其中版本时先解码，不把重新编码的结果当交付 JSON。`DELIVERY` 必须是新批次，
已有输出不得覆盖。所有后续路径都复用同一组变量。

`--repo` 指向包含固定公共合约和验证工具的仓库。`--evidence-feature` 显式选择 `.evidence/<feature>` 的独立检查范围。候选和检查都保留原件；失败不推进生效目录。`check --dry-run`、`apply --dry-run` 只读预览，仍需选择同一候选/基准。

查询候选时可显式选择某份检查原件：

```sh
"$CATALOGCTL" candidate --store "$STORE" --id '<candidate-id>' --repo . \
  --check '.evidence/catalog-review/check-inputs/<attempt>/check.json'
```

不提供 `--check` 只查询冻结候选，状态为 `awaiting_check`；不会猜测“最新检查”。提供引用后按同一套完整性/当前适用性规则返回 `applicable`、定义/代码缺口/执行失败等状态。自报 PASS、错候选、旧基准或已变化证据不能变成可应用。

候选在检查前固定完整目标 State、来源字节、基准、拟生效 Provider 版本和最终运行快照。检查通过真实受控 E2E 为最终目录的每条身份核对查询、请求/响应、Usage、计价及适用能力；缺合约、缺必需子场景、零匹配、非零退出或只提供自报 PASS 均不能应用。定义失败、`code_gap` 和执行失败分别报告。

`apply` 重新核对候选、基准、实际受测源码/输入的完整文件集合和哈希，以及原生场景、审计和实际退出。基准、代码、规则、输入或证据变化后应形成适用候选并重验。成功仅安装原受测字节，不重新取源、映射或改变受测版本；安装或目录同步失败时执行有界回滚，恢复旧指针和原输出字节/原不存在状态。若回滚自身不能完成或确认持久化，返回 `storage_failed` 并明确要求修复，不能视作成功或无变化。原历史、已有候选和失败提交留下的不可变对象保留。

## 来源、确认和价格

主要输入是 models.dev `catalog.json?type=all` 的 v1 兼容投影。`add` 使用显式 serving Provider/model 关联；canonical 资料只作明确关联的通用补充，不提供其他服务方或操作的路由与价格。

使用已留存文件时，指定 `--source <file>`、真实的 `--source-url` 和已知的原始 `--source-time <RFC3339>`。不提供文件时维护命令可从所选来源 URL 获取；这不是运行时联网。CLI 固定实际使用的原字节、URL、时间及摘要。不要把历史研究摘要冒充已保留的响应，也不要补造抓取时间或上游部署版本。

新增时明确提供 `--operation`、`--provider`、`--api`、`--model`、`--source-provider`，必要时用 `--source-model` 指定 offering 键。`--definition` 是人工 JSON 定义/覆盖文件；`--decisions` 保存 JSON pointer 对应的字段归属、原因及依据。先查看 dry-run 的 `missing/loss/conflict/confirmation` 等诊断，再显式确认。blocking 项不能仅用 `--confirm` 消除。

当官方文档/公告用于补足准确 offering 缺口时，显式选择 `--source-kind official`、真实 `--source-url`、`--source-time` 和原始捕获文件。`add` 还要求完整的人工运行定义及四维身份，不把官方捕获伪装成 v1 自动映射：

```sh
"$CATALOGCTL" add --store "$STORE" --source-kind official \
  --source "$WORK/inputs/official-capture.html" --source-url '<official-source-url>' \
  --source-time '<actual-RFC3339-time>' --definition "$WORK/inputs/full-model.json" \
  --operation chat --provider '<provider>' --api '<api>' --model '<wire-id>' \
  --reason '官方缺项补充及逐字段确认依据' --dry-run
```

确认后才用同一输入加 `--confirm` 冻结。`revise --source-kind official` 用于明确的人工补充，并保留原主源关联；原捕获、定义与决定文件的实际字节都进入候选追溯。已知类型能表达但当前代码不支持的候选保留为 `code_gap`，不进入生效目录；适配后重新检查，不标注最低程序版本。

- 身份始终是 `(Operation, Provider, API, ID)`。地区 Provider 分开；一个 offering 的第二条 API 也要单独明确确认，不能从 SDK 名展开。
- 缺省源 type 只生成聊天候选；decision 经确认进入 classifier。未支持类型、未映射模态、非整数限额及缺来源均不静默接纳。
- `revise` 只修改已有身份显式提供的字段或字段决定，不能改身份。人工值和人工覆盖不会被批量刷新撤销；改动它们需要新的理由和候选。
- token 价格按 USD/百万 token 确认输入、输出、缓存读/写。缺价不等于零价；明确零值与缺失分别处理。现有历史价格不自动变成来源管理字段，需先显式确认 serving 关联和归属。
- `cost-update` 只更新来源管理、语义/单位一致的价格，不更新协议或能力；人工字段保留并报告来源差异。图片 token/每张价、缓存 TTL、地区和服务等级条件不能混用。
- 阶梯必须确认严格大于边界和完整费率；barness 按整请求换价。v1 `size` 不能证明 gt/gte，`context_over_200k` 不能用来猜门槛或重复应用阶梯。
- missing、deprecated、账户列表变化和未来停服时间只形成审查提示，绝不自动删除型号。

批量改价以已审核的 serving 关联和字段归属为前提，例如：

```sh
"$CATALOGCTL" cost-update --store "$STORE" --provider openai \
  --source "$WORK/inputs/models-dev.json" --source-provider openai \
  --source-time '<actual-RFC3339-time>' --reason '固定来源价格复核' --dry-run
```

可用 `--operation`、`--api`、`--model` 缩小完整身份范围；预览后用同一输入显式确认。人工/历史字段、缺失费率、图片和无法对应的条件价保留待审，不靠全局确认抹去差异。

## 版本、历史与恢复

`versions` 查询当前 revision 和每个 Provider 内容版本；`history` 查询追加历史；`candidate --id` 查询冻结内容。各 Provider 独立升版，稳定排序组合得到 Catalog.Version。查询、预览、失败和无变化不推进内容版本；纯来源/决定元数据修订可产生新的维护 revision，运行内容版本不必变化。

`restore` 把已下架身份变为新候选，重新检查后才恢复。`rollback --provider <provider-id> --revision <historical-revision-id> --reason <reason>` 也形成新候选和新生效变更，不重写旧版本、旧 pin 或旧检查。已下架记录保留在维护 State，所选运行快照不会从内置旧条目补回；不可用型号请求明确失败。

## 默认数据与生成

默认来源是 `catalog/data/providers/*.json`，按实际 Provider 保存定义和人工决定；默认运行数据是生成的 `ai/catalog_builtin.json`。初迁移保留原 118 条记录（112 chat、3 image、3 classifier）的定义、顺序、价格及 null/省略语义。`ai/release/catalog-snapshot.json` 和旧 activation 原件保留原历史身份。

2026-11-02 首次采用完整审阅基线，新增
`chat / openai / openai-responses / gpt-6.1-terra`，默认目录为 119 条
（113 chat、3 image、3 classifier）。OpenAI 内容版本为 2，其他 Provider 为 1；
原 118 条定义与相对顺序保持。该型号的来源、采用和验证记录见
[本次唯一当前状态](../../.scratch/catalog-gpt61-terra/status.md)。

本地 `apply` 只推进所选维护库和明确选择的外部快照。导出已批准 State 或快照：

```sh
mkdir -p "$WORK/export"
"$CATALOGCTL" export --store "$STORE" > "$WORK/export/$CATALOG_VERSION.json"
"$CATALOGCTL" export --store "$STORE" --state > "$WORK/export/state.json"

go run ./cmd/cataloggen --state "$WORK/export/state.json" \
  --out "$WORK/export/generated-$CATALOG_VERSION.json" --dry-run
go run ./cmd/cataloggen --state "$WORK/export/state.json" \
  --out "$WORK/export/generated-$CATALOG_VERSION.json" --confirm
go run ./cmd/cataloggen --store "$STORE" --out "$SNAPSHOT" --check
go generate ./ai
```

stdout 重定向由 shell 写目标文件；CLI 自身的查询不会创建目录。替换已有正式快照应使用上述原子生成/确认入口，不用 shell 覆盖作为 apply 的替代。`cataloggen --store <store>` 可直接选择同一已批准 revision；`--builtin` 明确选择已审核入库来源。文件写入要求明确来源和 `--confirm`；省略 `--out` 只输出完整运行字节。

`go generate ./ai` 现在只执行 `--check`，验证入库来源与生成数据相等，不覆盖文件。生成器本身不授予候选兼容、厂商支持或发布资格。

需要采用已批准 revision 作为下一版入库来源时，导出到一个尚不存在的新审阅目录：

```sh
"$CATALOGCTL" export --store "$STORE" --repo . \
  --review-dir "$DELIVERY/review" --dry-run
"$CATALOGCTL" export --store "$STORE" --repo . \
  --review-dir "$DELIVERY/review" --confirm
```

审阅包包含 `providers/*.json`、完整 `state.json`、原受测字节 `snapshot.json`、`revision.json`，以及去重后的 `inputs/<digest>.bin`。最后写入的 `review.json` 标识完整导出，绑定文件库存、来源/人工定义/决定原件和外部 revision/candidate/check 引用及各自摘要。缺少完成标记的失败目录不能当成完整审阅包；重试使用新目录，不覆盖旧失败或既有导出。历史 catalog 身份与其原始文件字节摘要分别保留。

采用时在同一代码审阅中完成三项更新：整组 `providers/*.json` 更新到 `catalog/data/providers/`，`snapshot.json` 原字节更新到 `ai/catalog_builtin.json`，完整审阅包目录更新到 `catalog/data/baseline/`。基线目录须保留 `review.json` 和库存中的每个文件，包括 `inputs/*.bin`；替换旧基线目录，不能叠加遗留库存。初迁移仅含历史来源时允许没有基线 index；已有非历史来源却缺失基线，或库存、摘要、来源 URL/时间/类型、Provider/State/运行快照不一致时，新维护库拒绝启动。

维护工具编译时嵌入此基线，运行 `ai` 仍只嵌入生成快照。缺省种子 revision 留存完整基线字节，后续维护和再次审阅导出不依赖旧维护库仍在原位置，也不会重新抓取来源。继承的 revision/candidate/check 引用保留原路径和摘要，并用 `originReview` 绑定原始 `review.json` 字节摘要；它们是历史引用，不声明旧外部原件当前可取或当前资格仍通过。需要复核原检查时仍须保留并独立核验其原件。

相同 Provider 组成保持同一 Catalog.Version/Hash；采用基线形成的维护 revision 可因追溯链不同而拥有不同 ID。仅替换生成 JSON 会使 `go generate ./ai` 的核对失败。导出不移动当前维护指针；未来部署/线上发布仍是独立工作。

## 每次交付必须更新 changelog

在最终来源冻结前更新 [CHANGELOG.md](CHANGELOG.md) 的“未发布”条目；没有实际软件
发布时不能填写已发布版本号。条目至少包含日期、四维身份增删/改价、目录版本与
两类摘要、Provider 版本变化、原有定义是否变化、核心代码/其他代码变化、分发判定、
实际验证和未执行项，并链接本次完整记录及统一交付目录。

把本次条目作为交付目录中的 `CHANGELOG.md` 快照，说明其源文件和摘要；内容应能
独立阅读，不依赖仓库内部相对链接。已有历史条目、测量与 pin 不重写。事后纯文档
补录另留原状态及来源/输入相等性，不能修改旧回执冒充新测量。

## 每次同步内置目录必须给出代码与分发判定

在 `delivery.md` 和 `delivery.json` 同时记录完整比较基准/目标提交、实际受测提交，
以及未提交文件是否影响判定。基准是本次变更前的明确程序/交付来源，不能只比较
最后一笔文档提交，也不能凭“只增加一个型号”推断代码未变化。

核心范围包括 barness-ai 运行库的目录解释、Provider 协议适配、请求/响应、推理、
Usage/计价、授权/资源和实际运行依赖。检查完整差异及必要的依赖关系，明确列出
变化文件；示例 CLI、测试、维护工具、生成数据和文档分别列出，不混作核心代码。
使用 `YES / NO / UNKNOWN` 表达核心及依赖变化，未知不能写成无需升级。

| 使用方式/实际证据 | 必须给出的分发结论 |
| --- | --- |
| 指定受测程序已有外部加载/注入能力，且该原 JSON 已在其上通过兼容检查 | 可只分发版本化 JSON；Host 显式加载后创建新 Client，现有 Client 不热更新 |
| 新定义依赖新增核心能力/字段，原程序未通过或缺少加载能力 | 先升级 SDK/重建程序，再使用 JSON；列明实际代码依赖和证据 |
| 程序使用 `ai.BuiltinCatalog()` 的编译内置数据 | 更新默认目录需要重新构建和分发程序，即使核心运行代码没有变化 |
| 未取得目标旧程序的实际检查依据 | 外部 JSON 兼容性未验证；不能承诺任意旧版本或虚构最低支持版本 |

外部 JSON 与内置默认是两种使用方式，必须分别给结论。若只有示例默认配置路径
变化，说明其影响，不能据此声称 SDK 核心变化。软件发布、内置重建、JSON 分发及
厂商真实资格分别记录；目录升版不自动要求或授予它们中的任何一项。

## 完成与归集核对

1. 确认版本化运行 JSON 与候选原受测 snapshot 完全同字节，写完整文件 SHA-256；
   同时记录 `Catalog.Version`、`Catalog.Hash`、维护 revision 和 Provider 版本。
2. 归集完整 `review/`，逐文件验证 `review.json` 库存与摘要，不叠加旧批次文件。
3. 归集 changelog 快照、代码变化/两种分发判定、验证摘要及原件索引。记录检查与
   普通阶段的真实退出、必需场景、来源及输入集合、未执行/不适用依据；原件外部
   引用写明位置和摘要，不把受控通过标成厂商实测。
4. `delivery.json` 最后写入，绑定本目录全部文件的相对路径和 SHA-256（不含自身），
   并记录上述版本、代码变化和分发结论。验证目录文件集合与索引完全相等。
5. 核对 changelog、维护源、内置数据、基线和交付快照的一致性，更新本功能唯一
   当前状态并保留不可覆盖的完成索引；没有完成标记的部分目录不得分发，重试用新批次。

这些是现有 CLI 之上的维护与交付步骤，不是新增自动打包参数；未执行的动作不能
由手写 `PASS` 替代。当前 Terra 的归集实例见[统一交付入口](../../artifacts/catalog/README.md)。

验证范围、未执行阶段和不可变证据索引以本功能的[唯一当前状态](../../.scratch/barness-ai-dynamic-catalog/status.md)为准；测试/验收遵循[测试 SOP](../agents/verification.md)。
