# Grok 正式兼容目标范围选择

Status: needs-info

[实际目录核对](legacy-formal-eligibility.md)确认：受影响22旧目标身份中，16条有已发布Builtin型号，六条Kimi仍为候选。`ai/catalog_kimi.json`明确active=false、models=[]；`supportmatrix/regional.go`的正式构造器因此拒绝Kimi。新增Grok历史场景前即如此，不能用Grok证据替Kimi晋升，也不能把候选报告改成正式。

## 建议范围

正式Grok交付要求自身完整18能力与16条已正式支持旧目标的完整自身报告/新增xAI来源历史场景；六条候选Kimi继续完成受控兼容回归和候选合同准备，保留实验状态及实际live NOT_RUN，不在本次正式兼容性支持承诺内。22条受控代码路径均保留，不移除能力或改成UNSUPPORTED规避测试。

正式live集合为17条：xai-responses，openai-responses，deepseek-responses，openai-chat，deepseek-chat，zhipu-chat，zai-chat，minimax-chat，minimax-cn-chat，anthropic-messages，zhipu-messages，zai-messages，minimax-messages，minimax-cn-messages，google-gemini，minimax-responses，minimax-cn-responses。

保留候选的六条：kimi-cn-chat/messages/responses、kimi-intl-chat/messages/responses。它们在本功能中仍有受控来源碰撞/复用检查，但不得宣称已具有正式服务资格。后续Kimi晋升需它自身的完整证据和单独任务/授权。

## 若坚持所有22旧身份都必须正式通过

则Grok完整发布继续被六条未晋升Kimi阻塞，需要另行完成Kimi自身候选/晋升/正式全流程；该工作不能由当前Grok授权隐含完成。现有软件不会替它们打开正式目录。

## 选择的影响

该选择明确的是正式兼容承诺的目标成熟度，不授予任何新付费请求、race/压力或发布权限。自己的Grok18能力、既有16正式旧目标的全量原生门槛与新增历史场景、全部22目标受控回归均保持。当前release.pass仍false。
