# 29条 Provider 路径交付矩阵

口径：12个Provider身份、29条操作×Provider×协议路径；TypeSafe两个版本共用一条路径。22条完整适用证据、6条Kimi各9/10、1条MiniMax Messages13/14。适用证据不是整体发布完成；所有旧事实保留，release.pass=false。

来源与逐记录原件见 ../../.evidence/provider-final-readiness/delivery-matrix-v2.json。合同项数包含零请求预检/声明不支持，不等于付费调用数。

| Provider | 路线 | 协议 | 真实证据/合同项 | 当前状态与最小缺口 |
| --- | --- | --- | --- | --- |
| anthropic | anthropic-messages | anthropic-messages | claude-fable-5/claude-fable-5-1/claude-haiku-4-5/claude-opus-4-8/claude-opus-5/claude-opus-5-5: 12项 12 PASS | 完整适用；新增补采0 |
| deepseek | deepseek-chat | openai-completions | deepseek-flash/deepseek-v4-pro: 12项 12 PASS | 完整适用；新增补采0 |
| deepseek | deepseek-responses | openai-responses | deepseek-flash: 9项 9 PASS | 完整适用；新增补采0 |
| google | google-gemini | google-generative-ai | gemini-3.8-flash: 7项 7 PASS | 完整适用；新增补采0 |
| google | google-interactions-image | google-interactions | gemini-nano-banana-2.1: 2项 2 PASS | 完整适用；新增补采0 |
| moonshot-cn | kimi-cn-chat | openai-completions | 9/10 | 缺 kimi-k3/xai-origin-history；目录inactive |
| moonshot-cn | kimi-cn-messages | anthropic-messages | 9/10 | 缺 kimi-k3/xai-origin-history；目录inactive |
| moonshot-cn | kimi-cn-responses | openai-responses | 9/10 | 缺 kimi-k3/xai-origin-history；目录inactive |
| moonshot | kimi-intl-chat | openai-completions | 9/10 | 缺 kimi-k3/xai-origin-history；目录inactive |
| moonshot | kimi-intl-messages | anthropic-messages | 9/10 | 缺 kimi-k3/xai-origin-history；目录inactive |
| moonshot | kimi-intl-responses | openai-responses | 9/10 | 缺 kimi-k3/xai-origin-history；目录inactive |
| minimax | minimax-chat | openai-completions | MiniMax-M3: 14项 14 PASS | 完整适用；新增补采0 |
| minimax-cn | minimax-cn-chat | openai-completions | MiniMax-M3: 10项 10 PASS | 完整适用；新增补采0；地区项目收尾待完 |
| minimax-cn | minimax-cn-messages | anthropic-messages | MiniMax-M3: 10项 10 PASS | 完整适用；新增补采0；地区项目收尾待完 |
| minimax-cn | minimax-cn-responses | openai-responses | MiniMax-M3: 10项 10 PASS | 完整适用；新增补采0；地区项目收尾待完 |
| minimax | minimax-messages | anthropic-messages | 13/14 | 缺 thinking-on |
| minimax | minimax-responses | openai-responses | MiniMax-M3: 14项 14 PASS | 完整适用；新增补采0 |
| openai | openai-chat | openai-completions | gpt-5-mini/gpt-5.4/gpt-5.4-pro/gpt-6-sol: 11项 10 PASS,1 UNSUPPORTED | 完整适用；新增补采0 |
| openai | openai-decisions | openai-decisions | gpt-6-luna: 9项 9 PASS | 完整适用；新增补采0 |
| openai | openai-images | openai-images | gpt-image-2.5-sunburst-2026-09-08: 3项 3 PASS | 完整适用；新增补采0 |
| openai | openai-responses | openai-responses | gpt-5-mini/gpt-5.4/gpt-6-sol: 9项 9 PASS | 完整适用；新增补采0 |
| typesafe | typesafe-classifier | typesafe-system-one | jev-1.13.0: 3项 3 PASS；jev-1.14.0: 12项 12 PASS | 完整适用；新增补采0 |
| xai | xai-responses | openai-responses | grok-4.7: 18项 18 PASS | 完整适用；新增补采0；Grok发布收尾待完 |
| zai | zai-chat | openai-completions | glm-5.2/glm-5.3: 20项 20 PASS | 完整适用；新增补采0；地区项目收尾待完 |
| zai | zai-messages | anthropic-messages | glm-5.2/glm-5.3: 20项 20 PASS | 完整适用；新增补采0；地区项目收尾待完 |
| zai | zai-responses | openai-responses | glm-5.2/glm-5.3: 18项 18 PASS | 完整适用；新增补采0；地区项目收尾待完 |
| zhipu | zhipu-chat | openai-completions | glm-5.2/glm-5.3: 24项 24 PASS | 完整适用；新增补采0 |
| zhipu | zhipu-messages | anthropic-messages | glm-5.2/glm-5.3: 24项 24 PASS | 完整适用；新增补采0 |
| zhipu | zhipu-responses | openai-responses | glm-5.2/glm-5.3: 22项 22 PASS | 完整适用；新增补采0 |

最小真实事实缺口共7个：Kimi6、MiniMax thinking-on1。已具备单场景执行合同的是Kimi6次/USD0.60提案；MiniMax的执行调用数和费用尚未核定，不写成已经可执行的7调用授权。race/压力按影响另选，旧暂停轮次不恢复。
