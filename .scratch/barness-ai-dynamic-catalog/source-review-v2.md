# 分类加载状态：追加来源审阅

日期：2026-11-01。本记录补充 [原目录迁移审阅](source-review.md)，只解释新的分类加载诊断。原审阅表/原审阅文档及原测量、pin、授权、支持矩阵均保留。

## 实际问题与修复边界

公共加载 RED：`.evidence/barness-ai-dynamic-catalog/e2e/20261101T104121.913580Z-69uz6gkp`，实际 exit1。有效固定 Jev 可加载，非法图片限额已是 invalid_definition；但结构自洽的未来 TypeSafe 图片/Box 型号及 Decisions Box 定义错误地被归为 invalid_definition。

`classifierModelIssue` 把现有 ai 校验结果作为消息与“超出已实现协议子集”的结构化信息返回。只标记现有图片固定子集、Box 协议及图片位置子集分支；原 guard 条件、顺序、错误消息及成功结果不变。旧 `classifierModelProblem` 返回同一消息供 `Hash`/`NewClient` 使用。新增 `ValidateDefinitions` 消费此结构，加载器按既有 UnsupportedDefinitionError 归为 unsupported_definition；没有第二张能力表或根据诊断字符串猜测原因。

GREEN：`20261101T104158.347108Z-d1uy67fc` PASS/exit0。四个公共模型场景同时断言原 Hash/NewClient 的接受/拒绝和完整错误文本不变。没有增加可调用型号、格式或协议；只修正独立加载报告给出的处理方向。

## 精确来源追加

现有 xAI Images 来源审查因两文件实际字节变化继续拒绝，RED 保留于 `20261101T104252.729326Z-cfvc6ntu`，原生原因是 classifier_capabilities.go 漂移。以下两项只追加到既有确切审阅表，范围仍是 `xai × xai-images`。原 helper 的旧 C30 来源条目保留，未复写为新字节。

| 文件 | 原 724e9881 来源 SHA-256（空为当时不存在） | 当前实际字节 SHA-256 |
| --- | --- | --- |
| `ai/catalog_definitions.go` | `` | `fc3b769d8b9f0c2f87a9110d40ed5ed3cdfb0bea3cf4a7f92c51934c4549e76c` |
| `ai/classifier_capabilities.go` | `45545aee17f1521a0dbf6118902dc867357088fe73c4327c3d497b95e5976487` | `c06fc1d45ede568c5f973c232b160720f3013ee9c9723897059769cccb50aa94` |

实际补丁与追加索引：`.evidence/barness-ai-dynamic-catalog/loader-taxonomy-v1/taxonomy.patch`、`review.json`。此路线实际图像目录不含分类条目，且现有 Hash/NewClient 分类校验输出本身保持；所有认证、请求编码、图像解码、Usage 与计价代码未改。原始捕获及全部原生场景仍经现有消费合同校验。未来 helper/guard 编辑、改变原来源或其他路线仍拒绝；这不生成新厂商资格或执行权限。
