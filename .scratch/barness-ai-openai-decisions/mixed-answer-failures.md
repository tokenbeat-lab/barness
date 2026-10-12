# 工单 03：先于实现的失败清单

测试边界沿用本票明确指定的公共 `Client.Classify` / `HookedClient.Classify`，通过受控 HTTP Provider 观察 wire，通过公共结果、错误、Attempt、Observer 及资源计数验证。不测试私有 decoder。

- 三题混合、乱序答案、重复 score label、概率并列：不得按数组位置回配、排序/合并 levels、重算 confidence 或阈值化 bool。
- bool 无 criteria、两侧非空、两侧空、单侧空：固定双换行与 True/False 行；结构化说明在 capability 拒绝，未读取凭据。
- score 公共 1/11 档、型号收窄范围、Hook 改 kind/levels 超范围或 unsupported kind：scope/capability/request 错误顺序与零发送。
- 缺/多/重复/空/null 名称；重复字段和大小写别名；错误 discriminator；choice 非字符串/非最大/候选集损坏：整次 Answers 为空。
- predicate 缺/null/非数/溢出/负/大于一；score 缺/null/非数/溢出/越界、confidence/概率损坏、分布非数组、index 非整数/缺多重复/越界、label 不符、和/期望超 1e-6：原子拒绝，唯一合法 usage 保留，2xx 不重放。
- 1e-6 内外对照、Jev 两位舍入在 Decisions 拒绝、高精度不一致反例；保留原概率、score、Legend，无归一化。
- 混合和全拒答：完整最终名称集，stop/nil，空 RefusalAnswer；拒答夹带普通答案、原因、非法语义字段拒绝。
- benign additive metadata 在 envelope/answer/entry 可忽略；重复键、语义字段别名/矛盾仍拒绝；Jev strict decoder、宽 JSON、原生 refusal 拒绝和舍入 replay 不变。
- Hook Keep/Replace 合法改名、三种 kind 互换、重排问题/候选/levels：响应依赖最终冻结记录；迟改 Hook 引用不影响捕获/结果，原 public 对象不被修改。
- 错误/拒答保留唯一合法用量与单次 Attempt；资源释放、Observer 归属和整包审计有可重放证据。

实施基线：`9fa44f9ffea5252fdc0bcce71ef13e31515b90cc`。合成 Host 目录，live NOT_RUN；图片、usage presence 专项和真实服务验证不属于本票。
