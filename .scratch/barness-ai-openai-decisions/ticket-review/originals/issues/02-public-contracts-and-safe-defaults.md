# 02 — 公共契约、目录能力及图片安全默认

Status: ready-for-agent
Type: task
Depends on: 01
Updated: 2026-10-12

## What to build

新增 APIOpenAIDecisions、空协议 Options、请求 Images、RefusalAnswer 公共 JSON、Classifier Model/Capabilities、Classifier 图片 policy 和 Binding.AllowClassifierImages。依据 [R03/R05/R06/R07/R10](../spec.md)、[设计 §2/5](../design.md)、[OD01–08](../acceptance.md)。

## Ownership and context

负责公共类型、请求 JSON decoder、catalog 校验/clone/hash/查询、policy 校验与 Binding 新字段；E2E 从公开构造/查询/序列化及既有 Jev Classify 检查。尚不注册未完成的执行路线、不纳入内置 Luna、不更改 Jev metadata 的既有省略态。协作不回退他人编辑。

## Failure scenarios first

- 新字段必填导致旧 Host 自带 Jev Catalog 或 Binding 无效。
- 省略 Input/图片权限/图片政策被解释为可接收图片。
- Classifier 图片与 GenerateImages 子策略互相启用。
- 错/重复模态、说明格式、负数/矛盾计数、图片 MIME 元数据漏验。
- 请求 images 未进入 strict JSON、copy 或总字节检查；返回目录 slices 污染 Client。
- 新字段未进入 Hash，或默认解释强制写回旧目录而改变其内容。

## Acceptance criteria

- [ ] 先写 OD01–08 公共 E2E/default 对照和 closed public JSON 证据，再生产字段与校验。
- [ ] 公共 shape 及已接受的旧 MaxChoices/score 构造范围不收紧；旧省略默认行为照旧。
- [ ] 图片授权 false，图片 policy nil；非 classifier 开图片 flag 以 invalid_request/binding 拒绝。
- [ ] description typed enum、Input/MinChoices/MaxInputImages/MIME 的默认与协议声明规则一致，错误不回显值。
- [ ] 新 slices/map/raw data 在入口和全部目录返回边界独立，Hash 与 NewClient 共用规则。
- [ ] Request.images 的省略/null/空数组为无图，非数组和未知/case-alias image 字段拒绝；复用 Image 的 data/mimeType 形状。
- [ ] 新字段落地时立即在公共入口失败关闭：Jev 收到 Images 通过政策/模态检查拒绝，不得在 05 之前静默丢图；未实现的图片执行能力始终 capability 拒绝且零凭据/发送。
- [ ] RefusalAnswer 输出 only type=refusal；未实现的协议结果行为不在此工单声称支持。

## Verification and handoff

运行新 public contract/Catalog/default E2E，再运行 01 的旧声明/JSON/Jev 调用对照；保存字段、版本/hash 和独立修改证据。对尚未开放的 Decisions Classify 返回实际 unsupported route，不添加空 Adapter、假成功或环境凭据路径。

## Comments

2026-10-12：待实施，公共字段目前尚未增加。
