# 08: 历史归一化、跨模型降级与图片输入（E03 其余部分）

**What to build:** 应用开发者传入任意已授权的多轮历史（含跨模型/跨 Provider 历史、工具调用与结果、system 和工具声明变化、图片），模块按冻结 pi 的规则确定性地转换为目标协议，不修改调用方原始历史（spec I6 表格、User Stories 11–14、19）。

**Blocked by:** 07

**Status:** ready-for-agent

- [ ] content 缺失或 null 归一为空数组
- [ ] 同模型 thinking：保留 redacted；带签名的空文本 thinking 保留；其他空白删除，非空保留
- [ ] 跨模型 thinking：非空可见推理转 text，redacted 与空白丢弃；text 只保留文本；按原 truthy 条件删除非空 thoughtSignature，不合并缺失/null/空值
- [ ] 按目标规范化工具 ID 并同步 toolResult 关联；缺少工具结果按基线顺序补 `No result provided` 的 isError 结果，调用与结果之间的 system 消息按原规则延后
- [ ] error/aborted assistant 轮次按基线跳过；SystemPrompt 归一到初始 system 消息；system 与工具声明变化按输入顺序重放
- [ ] 支持图片的模型发送 user 图片与工具结果图片；不支持时生成对应占位文本并按基线合并连续占位——两条路径分别有用例
- [ ] 调用方原始历史在调用后逐字节不变
- [ ] 以上全部场景在 Responses 上接入 pi 差分，无待处理差异
