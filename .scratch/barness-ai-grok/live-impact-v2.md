# Grok 当前正式影响范围 v2

维护者已明确选择：[17条正式路线，Kimi保留候选](formal-target-scope-plan.md)。这是对正式兼容承诺范围的确认，未移除任何xAI必交能力或22条受控转换路径，也没有新付费/人工授权。

原[身份级影响分析](live-impact.md)仍作为历史设计依据：xAI来源会触发22个旧聊天目标的ID碰撞/跨轮复用分配；GLM Responses与五条非chat路径未受该行为变化。新增代码和公共受控回归仍覆盖全部22旧目标。

## 正式验收集合

Grok自身完整18能力，以及16条已发布旧目标各自的完整原生能力和新增xAI来源历史场景。所有当前正式资格仍需最终源码/目录/SDK/构建/预算/授权与原件一致；不存在借旧pin或候选报告充当正式的豁免。

```json
[
  "xai-responses",
  "openai-responses",
  "deepseek-responses",
  "openai-chat",
  "deepseek-chat",
  "zhipu-chat",
  "zai-chat",
  "minimax-chat",
  "minimax-cn-chat",
  "anthropic-messages",
  "zhipu-messages",
  "zai-messages",
  "minimax-messages",
  "minimax-cn-messages",
  "google-gemini",
  "minimax-responses",
  "minimax-cn-responses"
]
```

## 六条候选Kimi

`kimi-cn-chat/messages/responses`、`kimi-intl-chat/messages/responses` 的Builtin模型仍未发布（active=false、models=[]），不在本次正式兼容承诺中。保留来源碰撞/复用受控回归、10次调用候选预算与原生合同准备；实际live保持NOT_RUN，不标成不存在运行影响的NOT_APPLICABLE，也不宣称已正式支持。后续Kimi晋升由它自己的资格任务与单独授权处理。

目录事实和正式模板拒绝原因见[资格核对](legacy-formal-eligibility.md)。该范围避免把无关候选Provider晋升隐含加入Grok发布；Grok自己的18能力及16个正式旧目标的全量原生门槛均保持。
