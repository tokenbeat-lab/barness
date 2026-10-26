复核范围：`a3c3a269...e3ad302d5bffc234a7ad3563235bffdd95d4d819`，按指定 pin 读取；工作树较新的提交未混入判断。

- **已解决：备用缓存路由头。** `ai/hooks_run.go:70–74` 在 header 名归一化后，仅对 xAI 拒绝任何实际携带值的 `X-Grok-Conv-Id`，包括空字符串、多值与大小写变体；nil/空值集合不产生传输值。新增公共 E2E 同时覆盖零 HTTP 拒绝及旧 Provider 自定义头语义。符合 R16/R19 和设计第316行。

- **已解决：重复协议控制键。** `ai/responses_xai_json.go:13–84` 在 map 折叠之前检查请求、reasoning、工具、输入 item/part 与 text.format，识别转义后的同名键。`ai/responses_xai_request.go:46–64` 保留授权检查优先级，并按来源给出 invalid_request 或 callback_failed。业务 schema、参数/结果字符串和未知 native 扩展内部未被递归误判，也不重建原生数字或密文。新增公共场景覆盖四种可达来源、授权扩大和业务 JSON。符合 R07/R08/R19。

- **已解决：向旧 Responses 降级的工具关联碰撞。** `ai/opaque_history_ids.go:146–147` 将 xAI Responses 纳入既有完整来源 ID 的碰撞分配；调用与结果沿原重命名映射同步投影。新增 E2E 覆盖并行、跨轮和占用旧 ID，检查关联唯一、可见内容保留、私有状态移除与输入不变。符合 R12/R13/E18。

新增 shorthand 原生授权及候选报告最终落盘修复未发现新的规格问题。前者统一受担保的 message ID 识别；后者先保存原非PASS报告，再暂存/审计/晋升，晚期失败将 canonical/pending 作废并返回失败，新增子进程场景保留真实 I/O 故障和正式消费者拒绝证据。

本次为只读静态复核，未执行测试。剩余规格发现：0；范围扩张：0。完整冻结普通回执、真实 live、正式晋升和人工 race/压力仍分别待完成，不能据此宣称发布通过。
