# 工单 09 两轴复审

固定基点：`05b8d8788a5fbb6721becd0decdea4b4da316b90`；规格来源：
`issues/09-openai-image-generation.md`、规格的图像契约与设计第 7–9、11、13 节。
两名独立、只读子代理按 code-review 技能并行检查工作区差异。

## Standards

没有确认的工程标准问题。变更保留了操作授权、协议到领域的显式转换、有界 HTTP
读取、图像整体发布、资源释放和独立 Usage 快照。共享 unary 帮助函数服务两个真实
调用方，并直接替代原分类实现。

响应解析中的 raw/base64 暂存受 MaxOutputBytes 限制，解码使用固定缓冲区，不保留
完整响应体。复审没有将此判定为规范违例或可操作的代码异味。

## Spec

初审发现两个 P2，均已先通过公共 E2E 复现，再修复：

- 指针选项在绑定解析之后才复制。现于 generateImages 入口、调用宿主解析器前复制；
  typed nil 保持明确拒绝。`TestOpenAIImagesEntryOptionsSnapshot` 验证后续修改不影响请求。
- 已声明 mime_type 为 null 或空字符串仍成功。现要求非空受支持字符串并与格式/文件头一致；
  `TestOpenAIImagesInvalidEntriesKeepUsage/{null-mime,empty-mime}` 验证整体输出为空且保留用量。

复审确认修复正确，无遗留问题。没有额外发现生命周期、整体发布、用量完整性、计价、
观测隔离、扩展登记缺陷或范围扩张。编辑由 issue 10、真实目录/live 由 issue 11 交付。

Standards：0 项；Spec：2 项已修复，0 项遗留。初审最高级别为 Spec P2。
