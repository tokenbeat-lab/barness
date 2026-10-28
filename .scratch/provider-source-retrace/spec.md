# 六条旧测量的离线来源追溯

Status: resolved
Type: task

按用户指令，仅核验 OpenAI Responses/Chat、DeepSeek Responses/Chat、Anthropic Messages、Google Gemini 的原测量适用性。保留原 bundle、授权、claim、账本、暂停和失败状态，不执行厂商、race、压力或恢复轮次。

冻结构建回执保留原 regional 来源映射；须核对原 plan/授权 scope/实际 producer receipt+selectedInputs/manifest、原二进制 SHA 与来源 map 的完整哈希一致，并逐文件定位留存字节。找不到的原件不得伪造。新来源补充保存为独立原件，不回写旧 bundle。

当前消费者仅从 bundle 内取来源；接入显式、只读的来源补充参数，并将补充文件纳入消费者输入哈希。保持原生身份、场景、wire、usage、catalog、SDK、grant/claim、时间与预算校验。新输入只用于历史消费，不接到执行/授权入口。

失败方式先覆盖：无补充仍拒绝；完整且 scope 哈希相等的补充可沿用；缺失、篡改、其他来源或符号链接拒绝；来源文件在消费后变化不能沿用旧回执；原始 bundle 或 grant 篡改仍拒绝；传递到 release 时不能丢掉来源补充参数。

公共入口为编译后的 releasegate 与 consume_live.py 受控进程、既有主 E2E。普通范围 ai/e2e、release、supportmatrix、当前 releasegate CLI 和工具进程套件。产品 SDK/adapter 不变，live/race/压力均 NOT_RUN。当前状态见 status.md。

实际普通验收发现两个既有 CLI 合成授权测试遗漏私有 Git 命名空间，产生25份公共存储合成记录（原记录未改、无厂商调用）。失败轮已停止，新增记录保留并单独索引；新增公共 Git 探针断言须在领取前先失败，随后复用已有 privateGrantStore 隔离。全量重跑前后核验公共存储文件集合与字节完全相同。
