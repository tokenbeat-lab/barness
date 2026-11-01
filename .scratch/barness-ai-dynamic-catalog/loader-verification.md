# 工单 01：加载器验证记录

本文件记录实施事实，不替代唯一 [当前状态](status.md)。本轮没有厂商请求，没有 race、压力或发布通过声明。

## 实施与公共边界

- `catalog.Load([]byte)` 和 `LoadFile(path)` 返回有效 `ai.Catalog`、独立 `Report` 及文件级错误。先按原始 catalog 紧凑字节校验完整来源，再逐条严格结构/类型解码，保留失败原文与位置。重复完整身份的全部成员失败；空白、缺失和 null 身份按位置保留，不冒充完整身份。
- ai 的 `Catalog.ValidateDefinitions()` 复用原有目录/价格/采样约束、Kimi/xAI 固定子集和实际操作/API 路由。新增严格边界包括完整身份/名称、正数 chat 限额、已声明输入、已知模态及分类问题类型。`Catalog.Hash()` 算法、公共类型和显式 Host Catalog 配置契约不变。
- 路由谓词从原 runtime 按等价条件提取。该操作改变 `ai/call_runtime.go` 的字节，历史来源适用性必须另外审阅；不得因受控通过而自动更新厂商资格。
- 公共受控调用覆盖 chat/image/classifier；重试在已收到首次请求后修改调用者目录/报告/原文件，仍保留 Client 的请求字节、价格和版本。新条目不扩大 Binding 白名单。认证失败不刷新或下架 Client 目录。

## 职责适用性

加载器只接纳 ai 当前明确理解的定义，不复制另一套协议能力表。已知 compat 若在既有 adapter 中没有语义，不能仅凭本加载报告推导其硬约束已获得调用证明；候选协议合约须按实际声明单独验收。价格字段的上游提供情况、人工归属、单位和条件由维护边界确认；普通 CostRates 的合法零值不作为来源已提供费率的证据。非 Nullable 的显式 JSON null 会被拒绝。

## 原件与结果

[版本化索引](loader-verification-20261101.json) 固定最终命令、受测提交、完整实际 PASS 测试集合和有限必需 CAT case ID；manifest/audit 与来源/输入精确集合见原回执。

| Attempt | 实际状态 | 数字退出 | 受测提交 | receipt SHA-256 |
| --- | --- | --- | --- | --- |
| [20261101T093905.069330Z-7olygojq](../../.evidence/barness-ai-dynamic-catalog-loader/e2e/20261101T093905.069330Z-7olygojq/receipt.json) | FAIL | 1 | 4eae987a | ba48261f7147fb38d488a393702928c54975759fdcfe7db5b35c7526b39a2c9d |
| [20261101T093935.024642Z-aii2zr83](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T093935.024642Z-aii2zr83/receipt.json) | PASS | 0 | 4eae987a | 51335dcf0d694bbcdc28556823550eb3ca6d116b45f9736d421324967476fb56 |
| [20261101T093952.919087Z-j_g39tsc](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T093952.919087Z-j_g39tsc/receipt.json) | FAIL | 1 | 4eae987a | 433d697ccefbe50b393826dd04f640db61a26299e577a5bada0019e80dc472c3 |
| [20261101T094029.580916Z-107ads46](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094029.580916Z-107ads46/receipt.json) | PASS | 0 | 4eae987a | 558fb6549fc722b6a15d825832635b5f3c46f7b74d028e4ee3a3b639122faa5d |
| [20261101T094105.335727Z-17g2r905](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094105.335727Z-17g2r905/receipt.json) | FAIL | 1 | 4eae987a | d1f1cd895551ca5d7eb4b74306ab5b751283a9b372bd454df9272ecc5f0e16b5 |
| [20261101T094121.490033Z-gq1dnbsp](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094121.490033Z-gq1dnbsp/receipt.json) | PASS | 0 | 4eae987a | 20fc51dc811f3011e095d780420ee5c12e1a53b1fb0cfe16c827aa9308ec5e66 |
| [20261101T094150.657283Z-l4_bdhio](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094150.657283Z-l4_bdhio/receipt.json) | FAIL | 1 | 4eae987a | e356e6a032b53b441015c05784a67e24f664764b0a08606954ecc5f17990eedc |
| [20261101T094206.196791Z-h1jrjvzg](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094206.196791Z-h1jrjvzg/receipt.json) | FAIL | 1 | 4eae987a | 746df5cbc223085c45bda9fbc5de6de18862f5edc7c391900e793578c8652ba6 |
| [20261101T094224.920699Z-yillhzs1](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094224.920699Z-yillhzs1/receipt.json) | PASS | 0 | 4eae987a | 055ca36c02ca4c28e8d2bc1574eed007a5aaaffad48112b1d2942985ce5dbf5f |
| [20261101T094338.593980Z-xrphtq6v](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094338.593980Z-xrphtq6v/receipt.json) | PASS | 0 | 4eae987a | 7a6315c0a497cabef5bcd6c07ac743d1b1e224c8d56ac2584fe3da690be81a37 |
| [20261101T094358.499706Z-718a6l1a](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T094358.499706Z-718a6l1a/receipt.json) | FAIL | 1 | 7a70400b | 13c3ac3de7748a8677a61e314c187f97cb3453664bcfae522ad306ede83f9b8d |
| [20261101T095604.355364Z-bu16ha8b](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T095604.355364Z-bu16ha8b/receipt.json) | FAIL | 1 | 0e404ffe | f027f8356468c0644d9923cc1778235e48897e846d8a17229f199002eda78b2b |
| [20261101T095622.715671Z-wfwvbt63](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T095622.715671Z-wfwvbt63/receipt.json) | PASS | 0 | 6ed7ed3f | feb7e9da45c961a23ff3319d1c0d3cd24720cb671d340bd56b41c57a797a9b04 |
| [20261101T095707.682235Z-1bz74bwa](../../.evidence/barness-ai-dynamic-catalog/e2e/20261101T095707.682235Z-1bz74bwa/receipt.json) | PASS | 0 | 6ed7ed3f | a5a2b4ea6fbf88bf991fe79bb1fe68fa5585fad84f24b5bad6d2d85c5b8951da |

初次缺少新公开 package 的构建失败、后续结构/身份/hash 场景的实际失败，以及一次测试自身 fixtureReply 类型错误均保留；不把构建失败冒充已完成原生场景。最后两份定向回执的差别是最终命令纠正了迁移/地区组合测试的精确函数名；以前一份实际发现集为准，不声称未匹配的测试曾运行。

完整 `go test -json ./ai/e2e ./ai ./catalog -count=1 -timeout=20m` 在 7a70400bd0248a415f722275d2381eff7d2d42c6 运行 643.2537969578989 秒后退出 1。原生诊断为 `route product input changed: ai/call_runtime.go`；既有 `TestXAIImagesProducerOriginalDocuments/matching-original-git` 随后对缺失 `inputs` 作无保护断言而 panic，未完成 TestMain 的主 manifest。该 FAIL 不被定向 PASS 覆盖，也不是无效 Git 原件的问题。集成收尾按现有有限 reviewed_sources/source-review 流程核对机械差异及回放范围，本工单不修改消费者或历史期望。

最终定向命令运行完整受影响的新加载边界与 Client/目录公共回归，包含默认迁移和地区组合；所有实际 case 与 audit 通过。先前全包失败仍须由集成阶段处理，不能把本记录写成全包通过或 release.pass=true。
