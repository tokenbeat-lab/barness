# 历史分类图片发布验收

当前流程与本次迁移状态见[唯一当前状态](../../.scratch/barness-ai-test-process-review/status.md)；本文其余内容均为历史验收，仅适用于所记原受测源码。

2026-10-17最新整版验收已PASS：当前Go内容`9a8ea45`，实际报告提交`fc8f02b`，
公开releasegate退出0、10项门禁全部通过。11条独立真实支持和4项效果均被接受，
normal/race各9098 PASS、0 FAIL，各861 Pi及33 UC14，17个输入审计零发现；#14已resolved。
本次新增312 Calls/Attempts，包含已保留原失败的一次DeepSeek完整路线确认，低于550上限。
原配置失败、空推理失败、harness时限失败和先前TypeSafe/压力历史均保持原样。
当前结果、按Provider授权/有限预算和完整字节核验入口见
`.scratch/jev-1.14-multimodal-box/14-remaining-real/verification.md`。
下文的旧NOT_RUN与阻断段落保留其当时事实，不代表本次最终状态。

发布命令仍是 `go run ./ai/release/cmd/releasegate`。当前合同在原十条路线之外，
独立要求 `typesafe-classifier-jev-1.14.0`；`typesafe-classifier` 仍固定 1.13，
`openai-decisions` 仍为 Luna。必交能力不能用 UNSUPPORTED 绕过。
每条路线需要自己的当前目录 Hash、型号、SDK、完整报告、预算、捕获和审计。
Jev 1.14 正式包还需独立视觉效果包；门禁分别要求 Jev 与 Decisions 的当前正式目录固定视觉任务，
经已有严格 verifier 复核完整数据、源码、捕获重放、评分、预算和审计。
源码新增测试也会使该正式包过期；保留原包，在获准预算内重新采集，不能重写旧包。

当前入口已由 [ADR-0051](../adr/0051-scoped-verification-receipts.md) 取代；使用[发布指南](README.md#发布门禁)中的 `-record` 只读判定，普通、live、race、压力分别留独立回执。旧全库差分、固定追溯与自动普通/check/race 调度已移除。以下压力和资格说明保留原票历史事实。

原 `.scratch/barness-ai-thinking-sampling/issues/07-pressure-stream-integrity.md`
验收尚未 resolved 时，`pressure-stream-integrity` 快照检查直接阻断总体发布。
通过专项或完整重跑不能撤销其历史因果证据要求。本票不修改或关闭原工单。
该检查的来源文件和内容 Hash 随报告保留；原验收解决后再执行当前整版门禁。

迁移与可执行例子：

- `go run ./ai/examples/classifierimages` 覆盖两入口、两协议和 Host 消费。
- `Images` 是唯一来源，空引用列表保持旧顺序共享语义；显式列表声明全部使用。
- Jev 图片需有限策略、Binding 图片授权、固定型号白名单及
  `TypeSafeOptions{ExperimentalImages: true}`。静态 PNG/JPEG 与局部位置的边界见
  `docs/barness-ai/jev-instruction-images.md`、`jev-criteria-images.md`、`jev-box.md`。
  Decisions 拒绝局部位置和 Box；不提升作用域、不跨 Provider fallback。
- `ImageInputs` 是最终输入的 MIME/decoded SHA-256。Score 图例引用指向本次最终索引，
  pointer 相对该答案 Legend；不能当作入口索引或图片授权。Box 需独立存在性判断。
- 禁用 Binding、移除型号白名单或撤销图片权限影响新调用；已 pin 调用依原版本快照结束。
  不清除已报告 usage，不补文本或别的型号。实现回滚按提交整体恢复门禁与证据合同；
  不添加公共 Assets、内部兼容 shim 或双写。

本票的当前验收结果、命令和完整性回执见
`.scratch/jev-1.14-multimodal-box/14-verification.md`；没有当前真实证据时仍保留
NOT_RUN/整体阻断，受控协议验证与业务效果分别报告。

2026-10-17另行授权的续修已按原#07要求取得两轮独立自然因果RED、安全对照和完整
normal/race各9,098PASS、0FAIL；Decisions未执行清单和unary截止竞态亦已修复。
原#07只在验收后转换为resolved，启动时旧快照与原发布pass=false保持原样，不追认PASS。
真实支持及效果仍NOT_RUN，票14保持needs-info；当前续修报告、状态转换与独立封存入口见
`.scratch/barness-ai-release-verification/verification.md`。

2026-10-17后续获准的TypeSafe轮次已取得当前Jev1.13/1.14各自真实支持与独立效果包，
首轮51 Calls/Attempts后发现UC13迁移测试错误；用户追加授权修复并独立重采，当前源码
第二轮51/51、零重试，两轮累计102次。支持矩阵只更新这两行，其余9条支持和2项效果仍NOT_RUN，
不能据此宣称整版发布PASS。实际协议、效果分数、授权边界和本轮完整门禁见
`.scratch/jev-1.14-multimodal-box/14-typesafe-real/verification.md`；原历史报告和快照保持原样。
