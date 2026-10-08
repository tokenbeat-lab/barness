# 工单 13 双轴审查

基线：`52ce4feaa6001fc06895cffd7bf94eb6dcd82986`。
两个独立代理按 code-review 技能审查 `git diff <base>...HEAD`，规范与规格分别报告，
所有修复先补公共 Client/HookedClient E2E 红灯，再实现；未新增单元测试。

## Standards

初次提交 `16c1ada`：1 项 P2。普通回调输入的嵌入 data 字段、普通 byte slice、非字符串
TextMarshaler map 键之后的普通值，会绕过编码前单图预算。违反 ADR-0002 和 ADR-0022 的
已知长度先检查要求。`1270eae` 复用共享资源字段选择、计入 byte base64 膨胀并保留自定义
键的普通值限制，新增 promoted/byte-data/custom-key 分配量场景由红转绿；原发现已消除。

增量复核 `1270eae`：3 项 P2 编码语义问题，分别保留：

- selected field 转 interface 丢失 addressability，不能识别指针 MarshalJSON；大 backing
  string 编码成小图时误拒绝。修复保留 usesCustomJSON 对地址的识别，未知编码留给冻结后守卫。
- omitzero 只检查 reflect.IsZero，未遵守宿主 IsZero 方法；合法省略字段被预算误计。
  修复将自定义省略决定留给编码器，预检不额外执行宿主方法。
- Uint8 slice 被一律视为 base64；元素指针 JSON/Text 方法会让标准库采用数组编码。
  修复采用相同元素方法判定，保持逐元素自定义编码语义。

`a343a91` 修复三项；exact-cap pointer-marshaler/custom-zero 与 3-byte classifier
custom-byte-element 接受性 E2E 均由红转绿。相关 OpenAI 与分类已有编码/分配量 E2E 同时通过。

增量复核 `a343a91`：2 项 P2。第一项细化前轮的 IsZero 发现：不可取地址的值字段，
标准库也会临时装箱调用指针 IsZero，不能把它当普通字段计量。第二项是 selected
RawMessage 字段落入自定义编码豁免，已知 raw 长度被遗漏；三张普通 raw 图片超出总量
时仍会先发生大分配。`701f254` 保留 reflect.Value 的编码上下文和 RawMessage 计长
优先级，指针 IsZero 不执行宿主方法并交由编码器决定；pointer-zero 与 raw-struct-total
公共 E2E 均由红转绿。

增量复核 `701f254`：1 项 P2。自定义键 map 的普通值全部按单图计量，误覆盖第一个
prompt 项。`b1ac4da` 限定逐值单图检查到图片位置，提示仍受完整请求预算；
prompt-custom-key 在精确宿主预算下由红转绿。各轮共报告 7 项具体问题（含前轮 IsZero
的残余细化），最终遗留结论独立记录于 review-results.json。

## Spec

初次提交 `16c1ada`：1 项 P2。带类型输入只累计 data，漏计提示转义、MIME 与信封开销；
违反工单 13“含 prompt、全部内联图片与编码开销”“先估算再复制”。`1270eae` 改为共享
header/input 整体预算，计入所有已知字段、mandatory JSON escapes、数组标点与 byte
base64 膨胀；struct-overhead 场景由红转绿，原发现已消除。

增量复核 `1270eae`：1 项 P2。指针 marshaler 的字段 addressability 在估算时丢失，
违反 ADR-0002 自定义编码豁免，合法小请求误拒绝。`a343a91` 修复并以 exact-cap
pointer-marshaler E2E 证明。

增量复核 `a343a91`：1 项 P2。不可取地址字段的指针 IsZero 仍被漏判，合法省略项导致
小请求被拒绝。`701f254` 修复，pointer-zero 公共接受性 E2E 由红转绿。

增量复核 `701f254`：1 项 P2。自定义键的提示 map 被单图预算误拒绝，混淆了提示与
参考图边界。`b1ac4da` 修复，prompt-custom-key 公共接受性 E2E 由红转绿。
累计 4 项规格发现；最终独立复核结论记录于 review-results.json。

其余检查覆盖全部拒绝项、固定项删除/放宽、输入顺序与快照、数量交集、最终能力、
输出原子性与用量保留、无 2xx 重放、取消/资源释放、Observer 脱敏、扩展与追溯。
未发现额外范围扩张；真实型号、价格与本组合 live 仍由工单 14 验收。

## 最终复核

`b1ac4da`：Standards 累计 7 项具体发现已修复，剩余 0 项；Spec 累计 4 项发现已修复，剩余 0 项。
两个独立代理最终均未发现新增可操作问题，未合并或跨轴重排发现。
