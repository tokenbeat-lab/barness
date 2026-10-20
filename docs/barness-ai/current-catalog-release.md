# 当前目录的分类图片发布验收

本次门禁拆分后的实施与实际验证状态见[当前状态](../../.scratch/barness-ai-race-optimization/status.json)；本文既有 PASS 仍限定于各记录的原受测源码。

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

从仓库根目录先运行普通阶段，差分和压力保持开启：

```sh
go run ./ai/release/cmd/releasegate -stage check -out .evidence/NEW_RELEASE_CHECK \
  -evaluation-jev CURRENT_JEV_113_TEXT_EFFECTS \
  -evaluation-luna CURRENT_LUNA_TEXT_EFFECTS \
  -visual-evaluation-jev CURRENT_JEV_FORMAL_VISUAL_EFFECTS \
  -visual-evaluation-decisions CURRENT_DECISIONS_FORMAL_VISUAL_EFFECTS \
  -live CURRENT_ROUTE_PACKAGE

# 普通完成（READY_FOR_RACE，退出2）后，由人工明确选择：
go run ./ai/release/cmd/releasegate -stage race \
  -from .evidence/NEW_RELEASE_CHECK -out .evidence/NEW_RELEASE_RACE
```

这是ADR-0048定义的新执行方式；上方2026-10-17的历史验收结果保持原样。新入口默认不运行race；第二阶段重新验证原来源、输入和实际命令回执，十项门禁全部通过才认证发布。新工具源码可能使历史真实证明过期，前置会如实拒绝，不能将原PASS当作本次源码的资格。需要续期时另行确定范围和预算。

`-live` 对每个独立路线包重复；生产者和各自有限预算继续见
`ai/live/testdata/jev-formal/README.md`、`ai/examples/chineseeval/README.md`
及原各路线指南。命令本身不读取凭据、不代为调用收费服务。
缺少获准凭据或当前包为 NOT_RUN，历史 Hash、候选、受控 HTTP 或另一个型号的
PASS 均不能代替。当前报告经严格核验后只合并其自己的型号/路线，其他历史行保留；
整个发布仍需新的整版PASS，父规格或ADR交付状态不能依据局部支持记录提前更新。

`ai/release/traceability.json` 增加 UC-R01–17、UC-T01–16 和本票用户故事。
映射只选择证据；每个新条目固定完整的必需用例清单，缺任一项不能 PASS；状态由实际结果决定。normal/race 分别要求 UC09 两公开入口的
四调用双租户双协议图片负载，核对原 64 MiB、4 并发、引用/完整请求用量、
结果/图片身份/Attempt/Observer、完整清单和资源归零。原 local/cloud 负载仍完整运行。
`ai/release/classifier-extensions.json` 登记有限 case/path 的新引用、图片结果、Box
和严格校验语义；门禁核对其固定内容并复制进发布包；它不提供差分豁免，也不将其算作 Pi 差分 PASS。
唯一 Pi 1.1.0 基线、原 P07/P10 投影和差分账本保持。

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
