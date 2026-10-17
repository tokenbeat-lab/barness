---
status: proposed
date: 2026-10-16
---

# 分类图片采用统一来源和显式位置引用，原生素材机制归 Adapter

Host 需要以相同图片定义使用 Jev 与 Decisions。Jev 可把图片放在状态、说明和标准内，Decisions 当前接入子集使用全局有序图片；若把 Jev assets 原样增加到公共请求，协议差异会进入宿主。决定保留 Images 为唯一来源，增加独立的 ImageReferences，以来源索引和显式放置描述用途；value 放置通过 JSON Pointer 指向现存 null 槽位，普通文字不具有隐式图片含义。

该方案为拟议设计，尚未实施。图片本体复用既有 Image。引用缺省时每张图依次成为 context；显式引用是完整声明，所有来源必须被使用。两种写法在 Adapter 前归为一份不可变计划，延续既有公共输入且不引入兼容执行栈。Jev 私有 key、assets、multipart 与 Decisions data URL 均留在协议边界；无法保持位置作用域时在凭据前拒绝。

## 取舍

- 选择显式位置，放弃公共字符串模板：避免用户内容与引用碰撞、对象键误匹配和 substring 优先级泄漏。代价是句中图片需要结构化有序片段；自然语言仍不获得机器可验证的图片归属。
- 选择来源列表加位置表，放弃全量通用内容 AST：复用现有 RawMessage 与 Image，可保留任意 JSON 和精确数字；需要在校验时识别明确的图片 null 槽位。新增引用也不强迫迁移聊天和图像生成契约。
- Jev context 图片需要包裹原始 state；该结构投影需明确记录并单独测效果。Decisions 不把局部图片提升成全局上下文。
- 多模态 Score 图例在结果中还原为 null 槽位和公共引用，并以最终图片 MIME/hash 元数据解释索引；禁止把 native key 或图片 bytes 输出为图例。新结果字段是对外 additive 演进。

## 对既有决策的影响

扩展 ADR-0020 的图片位置能力与 Box kind，修订 ADR-0021 的 TypeSafe 请求/响应转换，并收敛 ADR-0026 的公共图片契约。上述快照、租户图片授权、有限重试、usage presence、拒答和原子结果不变。ADR-0026 仍为 proposed；本记录不改写其验收状态。

实施前先写公共 Classify E2E 和失败方式，保留无图及旧 Decisions 默认图片回归；真实路线、目录版本和效果证据各自验收。当前只发布[设计](../../.scratch/jev-1.14-multimodal-box/design.md)与[规格](../../.scratch/jev-1.14-multimodal-box/spec.md)，不更新支持矩阵。
