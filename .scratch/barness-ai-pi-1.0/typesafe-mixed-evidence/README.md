# 工单 06：TypeSafe 混合问题与 pi 1.0.0 差分

2026-10-08。Classify 在一份请求中提交 choice、score 与 bool，直接接收原生 JSON 状态和说明，
先登记用量后整体发布类型化答案。所有型号响应、状态和价格均为合成数据。
真实内建目录、官方费率和 live 支持声明的交付条件属于工单 08。

公共测试接缝由 issue 06 明确为 Client/HookedClient、受控 HTTP Provider 和冻结 pi oracle。
测试沿用宿主替身与既有证据设施，未新增生产测试接口或隔离单元测试。
[机器报告](report.json) 固定实际 bundle、manifest 哈希、场景状态、差分与审查结果。
完整大负载 bundle 留在本地 .evidence/barness-ai/；本目录 manifest 固定可提交证据的 SHA-256。

最终启用差分与压力的全量 race 运行 exit 0：3,641 个证据场景全部 PASS，P07 205 个，
冻结 pi 差分 623 个，待处理差异为 0，压力场景 11 个，脱敏审计 0 发现。
[最终运行记录](release-full-race.log)、[P07 用例索引](p07-cases.json) 和
[TypeSafe 12 份差分 verdict](p07-differential.json) 可独立核验；索引包含回放命令与 fixture 哈希。
两轴审查最终均无遗留发现，Standards 初次两项 P2 已先复现再修复。

[发布数据核验](release-evaluation.json) 的 P0 离线、差分、审计、追溯与快照门禁均通过，
工单 06 的三项需求追溯均 PASS。该调用使用 `-bundle` 评估已有证据，命令门禁按工具契约
标记 NOT_RUN，且未传入真实 live bundle，因此整体 release verdict 为 false、exit 1；
完整 race 和 vet 命令另有实际运行记录，这里不作整体验证发布或 live 支持声明。

## 验证行为

- 一次混合请求、2–10 个有序评分等级、bool ↔ noul、可省略的 bool criteria。
- string/object/array 状态与说明、choice null、独立副本；原始大整数、小数和指数不经过 float64。
- score 保留期望、置信度、可选完整分布和原生 JSON 图例；图例按最终等级与分布校验。
- 缺答/多答/错型、非法 choice 键、缺失/null/越界/非有限概率与置信度、错误键集、重复字段、
  分布和及期望不一致、错图例。任何错误整体清空答案，明确用量和授权归属保留。
- 回调合法替换问题/等级顺序，按最终冻结的问题集合校验；改型号、超限问题数/字节、
  错字段组合和无法编码的结果在 HTTP 前拒绝；原请求和保留的 callback map 不影响已冻结数据。
- 版本化 ResponseModel 与授权 ModelID 分离；不在本地计 token，厂商 422 保持请求失败分类。
- pi 1.0.0 新增 classify 和分类目录入口。相同的对象状态/字符串说明子集和回复进入两侧，
  pi modelPatch 固定 jev-1.13.0，maxRetries 显式等于 Binding，所有重试请求均对照独立 fixture。
- 投影只移除 score.probabilities/legend 与双方 usage.cost。ResponseModel、Code/Phase、宿主归属
  留在离线证据。pi 的其余结果字段及完整请求保留比较，未知目录字段阻断测试。
- 更宽 JSON 输入、评分详情、严格校验、版本 ID、输入计价与六个 transport header 差异
  逐路径/用例登记。三个实际差分用例证明严格校验比 pi 更强，不产生部分答案。
- P07 离线/差分/扩展场景进入 P0 与需求追溯。内置目录快照字节相同，无新增 live 支持声明。

## TDD 与决定

[mixed-red.log](mixed-red.log) 先复现旧 choice-only 解码器不能接收 score/bool；随后扩展封闭问题、
答案及边界转换，mixed fixture 通过。
[oracle-red.log](oracle-red.log) 先复现冻结 runner 不支持 TypeSafe，新增 classify 后实际运行两侧，
[oracle-first.log](oracle-first.log) 保留最初未登记的精确 transport 和严格校验差异，最终零待处理。
[legend-red.log](legend-red.log) 先复现超大 JSON 指数被 rational 解析器拒绝，且脱敏审计默认 float64
误判合法 JSON；按系数/指数精确比较，审计改用 json.Number 后通过。
[review-fields-red.log](review-fields-red.log) 先复现结构体解码大小写别名覆盖，以及显式 bool criteria:null
被当作省略；精确字段名边界校验与 presence 解码后通过，答案仍整体失败、用量保留。
分类目录比较也写出差分 verdict；[catalog-record-red.log](catalog-record-red.log) 先显示两个有意的
目录边界差异未登记，补齐仅限该目录用例的 baseUrl/input 决定后通过。

采用的厂商形状来自 https://docs.typesafe.ai/api 和 https://api.typesafe.ai/openapi.json，
抓取日期 2026-10-08；[OpenAPI 快照](typesafe-openapi.json) 与文件哈希在 manifest 中。
官方 legend 值允许对象和数组，因此采用 map[int]json.RawMessage，
区别于设计草案中的字符串图例，理由、影响与回退条件见 ADR-0021。

## 复现

从仓库根目录运行；冻结 npm 安装沿用既有 lockfile，无新增依赖：

    go test ./ai/e2e -run '^TestClassifier' -count=1
    BARNESS_AI_PIDIFF=1 go test -race ./ai/e2e -run '^TestClassifier' -count=1
    BARNESS_AI_PIDIFF=1 go test ./ai/e2e -run '^TestPiDifferential$/^typesafe-system-one$' -count=1
    node --test ai/internal/testkit/pioracle/node/runner.test.mjs
    go vet ./...
    go vet -tags live ./ai/...
    BARNESS_AI_PIDIFF=1 BARNESS_AI_PRESSURE=1 go test -race ./... -count=1
    go run ./ai/release/cmd/releasegate -write-snapshot -snapshot /tmp/barness06-catalog.json
    cmp ai/release/catalog-snapshot.json /tmp/barness06-catalog.json
    go run ./ai/internal/testkit/audit/cmd/auditbundle .scratch/barness-ai-pi-1.0/typesafe-mixed-evidence

两轴 code-review 基点为 0a3ba58bdaccc32fb1925203134c35c397183161，具体报告见 [review.md](review.md)。
