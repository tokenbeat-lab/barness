# GPT-6.1 Terra 目录接纳与内置采用

按 `docs/barness-ai/catalog-maintenance.md` 的 CLI 流程新增
`(chat, openai, openai-responses, gpt-6.1-terra)`，输出根目录
`local-catalog.json`，供 Host 显式加载。保留原 118 条定义。

以 models.dev offering 为主要核对输入；若映射无法表达，使用已捕获的官方
型号文档和完整人工定义。必须确认四项 USD/百万 token 价格、严格大于的
阶梯边界、推理档位、默认档位和适用缓存期限。

依次执行预览、确认冻结、固定公共受控 E2E、应用预览和确认应用。
保留实际 argv、数字退出、来源时间/摘要、候选、检查、revision 和输出摘要。
目录接纳不等于厂商实测、支持矩阵更新、内置目录采用或线上发布。

## 2026-11-02：采用为内置目录

用户进一步要求将已批准的本地 revision 同步到默认内置目录，并由
GPT-6.1 Terra 子 agent 编排和执行对应验证。

按 SOP 导出已批准 revision 的完整新审阅包，在同一变更中采用整组
`providers/*.json`、原 `snapshot.json` 字节和完整基线目录，包括 `review.json`
及库存中的所有 source/definition/decisions/historical blobs。不得重取源、
重新映射定义、修改旧检查或混入遗留基线库存。

验证须证明新构建的公共 `BuiltinCatalog` 和 Client 真正接受此型号，以及
推理、输入、用量、标准价格和严格大于 272K 的整请求阶梯。旧 118 条历史定义、
原相对顺序和原价格合同保持。无旧本地 store 的新维护种子必须从完整入库基线
产生同一运行身份，并能再次导出完整来源。按影响执行普通 E2E 与必要构建/生成
核对；live/race/压力分别作有依据的适用性判定，不把旧授权扩大到新厂商请求。
