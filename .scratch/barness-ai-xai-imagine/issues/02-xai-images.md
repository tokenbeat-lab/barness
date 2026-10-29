# xAI 同步生成与编辑协议

Type: task
Status: resolved
Blocked by: none

## Scope

公共 Client E2E 先失败后实现；选项、身份授权、JSON/最终 Hook 守卫、同步原子图片、ticks/token、最终请求估算、共享 unary 生命周期与资源隔离。与 01 的领域接口集成。

完整契约：[spec](../spec.md)。验证遵循 [SOP](../../../docs/agents/verification.md)。

## Comments

2026-10-29：implement-spec 根据现有规格建立工单依赖图。

## Answer

2026-10-29：公共 Client/HookedClient 同步生成、单图/多图编辑、XAIImagesOptions、严格最终请求守卫、MIME/原子输出、ticks/token、按张估算与共享生命周期接线完成。xAI 自身输入边界不继承 OpenAI URL 上限；共享输入先 Host 长度检查与快照，再按协议检查硬限及严格内联内容，凭据读取前完成。用量和报告收费在响应正文/图片后续失败时保留。候选仍不代表正式支持。

协议来源：`c28e9775`、`ae66c62b`、`e2d81cd5`、`93ef7708`。最终分支包含集成分支 tip 的 merge。

普通证据（独立工作区 `/Users/cyber/.codex/worktrees/xai-images-protocol/barness`）：

- 首个公共生成红：`.evidence/barness-ai-xai-imagine/e2e/20261029T000301.482372Z-8ed6j5np`，FAIL/exit1。
- 响应/收费红：`20261029T000504.559048Z-rvq0u012`，FAIL/exit1。
- 费用红：`20261029T000748.899223Z-x1tq4rxh`，FAIL/exit1（本轮其余响应场景通过）。
- Hook 编码前限额红：`20261029T000806.968333Z-s69l8us_`，FAIL/exit1。
- xAI大输入边界红：`20261029T001044.787223Z-k0x1bq02`；Hook 路径边界红：`20261029T001246.102028Z-34tfx2zy`，均FAIL/exit1。
- Imagine + OpenAI/Google 图片回归：`20261029T001304.063841Z-237yttb0`，PASS/exit0，命令 `python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage e2e -- go test -json ./ai/e2e -run '^(TestXAIImagine|TestOpenAIImages|TestGoogleImages)' -count=1`。
- 最后新增有界租户准入/Observer 独立性后完整 Imagine 组：`20261029T001425.657865Z-luc_6ycl`，PASS/exit0，命令 `python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage e2e -- go test -json ./ai/e2e -run '^TestXAIImagine' -count=1`。

全部历史失败（包括构造夹具/编译错误）保留原件。最终全受影响普通包由工单04冻结并验证。未执行付费、race、压力或正式支持晋升。

### Review 修复（2026-10-29）

同一实现 worker 在 `5e2ee121` 修复双轴审查的四项去重发现：

- usage 顶层和 detail 的重复字段只删除该歧义字段；其他唯一合法 token/收费事实继续保留，结果仍为 protocol 且清空 Content。
- 图片校验循环、校验失败、发布前和最终解码结束均保留 context 的取消/超时分类。
- 带类型的资源容器按实际 JSON 字段选择累计已知总请求字节；预算失败发生在 Host Marshaler/冻结编码前。
- 目录构造和发现测试通过注入 resolver 的总读取计数证明零凭据读取。

公共 E2E 先红后绿，全部历史原件保留：

- 红：`.evidence/barness-ai-xai-imagine/e2e/20261029T002223.770584Z-u8sqyh19`，FAIL/exit1。三组产品问题均复现；加强目录断言本身通过。
- 定向绿：`20261029T002315.202810Z-tbb37zab`，PASS/exit0。
- 完整 Imagine/Catalog 与相邻 OpenAI/Google 图片回归：`20261029T002333.203692Z-hsvj9jxr`，PASS/exit0，受测源码提交 `5e2ee121`。完整命令为 `python3 tools/verification/run.py --feature barness-ai-xai-imagine --stage e2e -- go test -json ./ai/e2e -run '^(TestXAIImagine|TestXAIImagesCatalog|TestOpenAIImages|TestGoogleImages)' -count=1`。

原件仍位于本工单独立工作区，父集成交付保存完整 attempt 后归档工作区。没有执行付费、race 或压力。
