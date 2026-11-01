# 模型目录迁移：xAI Images 来源适用性审阅

日期：2026-11-01。此记录只解释既有 xAI Images 原件的当前源码适用性，不生成厂商测量、授权或支持矩阵事实。

## 原件与范围

原测量来源为 `724e988132ac5bf6ac393db156e07081eb0e2427`，完整来源清单固定在 `ai/e2e/testdata/xai-images/measured-source.json`；本功能起点为 `9a43ec0827b0ab794501682a7e66037a592a2f8d`。两提交的下列三个已有文件字节相同，新 helper 在原来源中确实不存在。审阅后的产品代码提交为 `0e9ef483cb132c681e7a09b22e230190b19b9511`；四个文件的后续任何编辑仍会拒绝。

| 文件 | 原来源 SHA-256（空表示当时不存在） | 本次实际字节 SHA-256 |
| --- | --- | --- |
| `ai/call_runtime.go` | `1ab0a490768dd66283591a85c8bd4551ab9c0cc777341b07660544341a6a3bb1` | `dcbc4a104780c1ddb4edead79aa1a7c3029bdae094eb9ce7fa3b93108a37ffb6` |
| `ai/catalog_definitions.go` | `` | `c30aeb0bbb90c7635ae17555d52313549b3be2ddfd6c1367e1584afdbf9a2df6` |
| `ai/catalog_xai.go` | `8aae1166490ce06100e7fc934ad1d22e200c21791c95d989a2c84d0cce517c88` | `9a11958fbf29bce4d0e54c4867ce5af22a87f295a795b413fe7cfc96b1574dfc` |
| `ai/catalog_validate.go` | `c02aef617fc87fba5b9159a1fdd6350915b26d4abce94035c96ef00232493f72` | `97ac8eec4f00117ec4060974a12ced11ae031bd0e5e83e6fb49503b899cc7b27` |

完整实际补丁和四项迁移索引：`.evidence/barness-ai-dynamic-catalog/source-review-v1/runtime.patch`、`review.json`。原审阅表的旧条目逐项保留；只追加 `xai × xai-images` 的确切迁移，不把它们扩大到其他路线。

## 逐文件结论

- `call_runtime.go` 只把已有操作/API 判断提取为 `supportsOperationAPI`。图像和分类的 Provider/API 布尔式逐项一致。聊天仍查询 `registry()` 的四个键；原 `Client.adapters` 在 Client 构造时从同一不可变 registry 创建。没有改变授权顺序、Binding、网络尝试、用量或价格选择。
- `catalog_definitions.go` 的路由 helper 是上述原式；新增 `ValidateDefinitions` 在本地加载/候选检查边界使用，未替换 `Hash`/`NewClient` 的既有配置合同。xAI/Kimi 固定型号约束仍由原函数决定。此文件继续属于运行时来源，不能按文件名永久豁免。
- `catalog_validate.go` 仅将 `choice/score/bool/box` 的同一四项表达式提取为 `classifierKindSupported`。`Catalog.Hash()` 的编码及接受语义不变；xAI Images 本身不经过分类分支。
- `catalog_xai.go` 仅删除 `builtinXAICatalog` 数据构造函数，`xaiModelProblem` 的完整函数体保持。保留文件的运行时归属，并限定准确前后哈希。

新 `catalog_builtin.json` 是生成的纯定义数据；与原 `catalog_builtin.go` 等构造路径一样，适用性使用 `VerifyRouteCatalog` 对该原生合同实际使用的全部型号进行能力/限额/价格投影比较。不得以新的全目录版本冒充原版本，不能因来源分类跳过型号比较。旧 MiniMax/Kimi/地区构造器的删除未加入永久文件名豁免；未来重新引入仍按原运行时归属拒绝。全 118 条定义与原快照的独立对照见 `TestMaintenanceCatalogMigration`，原发布快照/pin 保持原字节。

## 失败与验证

父阶段完整普通失败保留在 `.evidence/barness-ai-dynamic-catalog/e2e/20261101T094358.499706Z-718a6l1a`。本工作树先重现原资格 CLI 失败（`100651.165542Z-j0y2w0_v`），再建立精确迁移、后续 helper 编辑、改原来源、错路线、缺运行时文件及生成数据分类的失败场景（`100753.894057Z-ygxagdr1`）。首次修复后 `100826.680234Z-20m6l0gd` 仍以 `catalog_validate.go` 漂移失败，原件保留；补充同值表达式审阅后 `100924.392357Z-p38c1_xw` PASS/exit0。

完整匹配原始 Git 文档、缺失/改写原文档、缺运行时代码/JSON 的既有真实进程场景均在受控夹具上执行。原始捕获、授权、账本、退出、manifest/audit 和每个原生场景仍由既有消费者逐项核验。受控 fixture 合格只验证消费机制；本次未消费新真实测量，不产生新的 live 资格。最终扩展普通结果另记工单 05 和唯一当前状态。
