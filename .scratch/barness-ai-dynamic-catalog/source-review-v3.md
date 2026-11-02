# 图片声明与实际选项：追加来源审阅

日期：2026-11-01。接续 [v1](source-review.md)、[分类诊断 v2](source-review-v2.md)；旧文档/确切来源迁移及原测量、pin、支持矩阵均保留。

## 公共失败与实现

RED：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T104838.761827Z-tfen6v_0`，exit1。Google 声明 mask、透明背景、自定义尺寸、输入保真度、固定 size/quality 等当前选项不消费的能力仍被加载；OpenAI aspectRatio/imageSize，以及现有请求数量、参考数量、quality/fidelity guard 会拒绝的声明也被加载。三条有效内置图像记录和非法零输出上限的对照已正确。

在 ai 内新增 `imageDefinitionProblem`，只由本地 `ValidateDefinitions` 在旧 `ImageModel.problem` 确认结构/价格自洽后使用。Google 依据现有 `GoogleImagesOptions` 和最终 payload guard，只接受实际可表达的选项；OpenAI 复用 `OpenAIImagesOptions.validate()` 的数量和闭合 quality/fidelity 规则。参考数直接复用现有 `googleImagesReferenceLimit`、`openAIReferenceLimit`，没有新增厂商限额或第二张 Provider 表。Google MaxOutputImages 仍是响应集合的资源边界，不推造一个不存在的精确 N 请求选项。

旧 `Hash`/`NewClient` 不调用新增严格边界，接受/错误文本保持。GREEN：`20261101T104923.379641Z-r7ddjbyd` PASS/exit0，17 图片声明、4 分类诊断与初迁移合计 24 个实际 Go 测试；每项均验证旧 Hash/Client 合同。有效 118 条数据、模型类型和请求路径不扩展。自洽但不可表达的声明现在报告 unsupported_definition，非法值仍按原验证报 invalid_definition。

## 精确来源追加

以下两项只允许既有 `xai × xai-images` 原件在这些当前字节上消费。先前 C30/fc3 helper 的确切条目与分类 C06 条目保持，不改写为新字节。

| 文件 | 原 724e9881 来源 SHA-256（空为不存在） | 当前实际字节 SHA-256 |
| --- | --- | --- |
| `ai/catalog_definitions.go` | `` | `2ca081f3218cf52584679b01e9703136b7317cabfde2cbe74cbef69726a66bc6` |
| `ai/image_definitions.go` | `` | `3a2a301ea8398e1a58ecefb9cdd6540f0c0e3edce1961c4a9059aa9cac7a57cd` |

实际补丁与原/当前 classifier guard 条件逐项相等性见 `.evidence/barness-ai-dynamic-catalog/image-capabilities-v1/definitions.patch`、`review.json`。xAI 当前图像定义继续经原固定 guard 接受；新增声明投影没有 xAI 分支，不改其 wire、SDK、认证、解码、Usage 或价格。原生消费者仍核验完整来源、实际型号投影和原始捕获/授权/退出，未来编辑、错路线或改变原来源继续拒绝。本次仅受控/普通机制验证，不生成新真实厂商资格或权限。
