# 工单 05 两轴 code-review

固定起点 `ffe70054b0daf853ff632076a760d37e9ba063f9`，最终生产实现 `926ed8a`，当前分支 main。
按 implement 技能要求，Standards / Spec 由两个独立只读审查 agent 并行执行，再分别复审。

## Standards

初审发现一项 P2：Hook 的 raw JSON / typed RawMessage text 未在序列化前按
Classifier.MaxStateBytes 检查，最终虽拒绝但发生比例分配，违反 ADR-0002 的 known size 边界。
公共红色用例实际测得约 28–169 MiB 分配；`9ba77ae` 增加有界 unescape byte 计数，随后这些
用例均在小于 1 MiB 分配时拒绝，合法 escaped/surrogate 边界通过。

复审发现一项 P2：相邻两个 Unicode escape 的分组可能拆开跨窗口代理对，
`\u0041\ud83d\ude00` 的实际五字节字符串被计为七字节。先以公共 Hook 在五字节预算下
复现红色，再由 `e58eed1` 改为仅合并真正的 high/low UTF-16 pair，保留标准 decoder 对
孤立 surrogate 和非法 UTF-8 的替换语义。

交叉核验又发现 typed message/content 能符合单项状态/decoded 预算却超过完整 JSON 预算，
在拒绝前复制。两条公共红色分配回归确认后，`e7325e9` 对所支持的 native struct 字段、
容器、keys/scalars 累计编码字节，并预留 model/questions 的编码成本。

该补丁的独立复审发现一项 P2：普通 []byte 应按 JSON base64 string 而非 array 计数；
两条合法整请求边界和两条状态限额分配回归均先红后绿。`926ed8a` 复用既有
byteSliceEncodesBase64/jsonPayloadValueSize，root/typed text 另按 base64 characters 检查状态预算，
custom encoder 和带方法的 byte elements 继续遵循实际 encoder 的边界。

最终 `926ed8a` 独立复审：0 remaining findings；judgement-call smells 0。
全部已记录发现已解决，aggregate `ffe70054...HEAD` 无其他可执行标准问题。

## Spec

初审发现一项 P2：合法 Hook message 的显式 `type:"message"` 被封闭 decoder 误拒绝，
与 design.md 的“message type 可省略或等于 message”不符。六个公共用例覆盖省略/合法、
null/非字符串/alias/其他 type；红色确认后由 `9ba77ae` 修正，全部绿色。

复审独立发现与 Standards 相同的 Unicode 跨窗口误计；`e58eed1` 的五字节公共边界用例
及真实 high/low 判断修正此问题。

后续 encoded-budget 的复审也独立发现普通 []byte 的合法边界误拒绝；`926ed8a` 的实际
base64 JSON 计数和 root/typed 公共回归解决了它。

最终 `926ed8a` 独立复审：0 remaining findings；无其他缺失要求、错误行为或范围扩张。

Standards 0 遗留发现；Spec 0 遗留发现。
