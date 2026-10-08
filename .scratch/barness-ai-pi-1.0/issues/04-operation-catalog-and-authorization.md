# 04: 按模型操作管理目录与 Binding 授权

**What to build:** 宿主能分别发现聊天、图像和分类型号，并为每份 Binding 固定一种模型操作；现有聊天调用继续工作，未知操作和聊天入口对其他操作的误用在读取凭据之前得到明确拒绝。

**Blocked by:** 03 — 切换唯一 oracle，完成五项聊天对齐。

**Status:** resolved

**依据：** 规格“模型操作与目录”“Binding 与授权”“词汇表”、实施设计第 6 节，以及 ADR-0001、ADR-0003、ADR-0010、ADR-0018。

- [x] 先写公共目录构造与聊天调用的失败场景：未知操作、绑定操作不匹配、白名单之外的型号、重复完整键、矛盾能力、缺费率、配置修改泄漏、跨租户同名绑定和凭据轮换冲突。
- [x] Operation 只取 chat、image、classifier；Binding 的零值规范化为 chat，其他未知值以 invalid_request/binding 失败。显式记录零值是兼容旧宿主且不扩大授权的公共契约例外。
- [x] Model 保持原有字段和聊天含义；ImageModel 与 ClassifierModel 各自携带身份、能力与价格，宿主能强类型查找和列举三种型号。型号身份为操作、Provider、协议、ID；同一 ID 可跨操作并存，重复完整键在构造 Client 时拒绝。
- [x] 目录拒绝矛盾的模态或能力、非法分类能力范围、负数或非有限价格、缺失的已声明模态费率，以及聊天型号 samplingParams 中的保留字段。
- [x] 新增切片、map、能力和价格全部深复制；构造后修改原配置或查询结果不改变 Client。目录版本与哈希覆盖三类型号，内容变化同步更新快照。
- [x] 聊天入口先核对绑定操作，再求该操作下目录与型号白名单的交集，最后读取凭据并核对一致快照。操作不匹配为 tenant_denied/capability，凭据读取次数与 Provider 请求数均为零。
- [x] CallAttribution 与 CallStarted 从聊天入口开始就记录 chat；Provider、协议、型号、账户和版本仍只在快照一致后填写，预检失败不伪造归属或 Attempt。新增观测字段通过脱敏审计白名单审阅。
- [x] 旧宿主的 Model、Target、聊天目录查询和未设置操作的 Binding 继续编译与运行；通过既有六条聊天路线回归和新目录 E2E，产出快照与审计证据。
- [x] 新增“多类型模型操作与 Binding 授权”ADR，修订 ADR-0001、目录包含准则、公共契约与词汇表。Classify、GenerateImages 的跨入口拒绝分别在其可用切片中验证。


## Comments

2026-10-08 完成：以公共目录构造/发现和四个聊天入口先写失败场景，再实现操作授权、独立型号类型、
完整身份索引、目录校验与深复制。目录升为 `2026-10-08.2`，同步 ADR-0020、ADR-0001/0018、公共契约
和词汇表。完整 3435 E2E PASS（新增 90）、611 pi 差分 PASS/0 pending、7 pressure PASS；race、vet、
目录快照逐字节核对和脱敏审计 PASS。两轴审查：Standards 无硬性违规，有一项保持已批准 []string 形状
的非阻断类型建议；Spec 无发现。Classify/GenerateImages 的反向跨入口拒绝留在其实际可用切片验收。

详细证据、先红后绿记录、三类宿主快照与重放命令见 [验证记录](../operation-catalog-evidence/README.md)。
