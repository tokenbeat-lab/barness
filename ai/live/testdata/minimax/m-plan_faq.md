> ## Documentation Index
> Fetch the complete documentation index at: https://platform.minimaxi.com/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# 常见问题

> 了解 M Plan 套餐、用量、API Key 与使用规则的常见问题。

<div id="available-plans" />

## 现在有哪些 M Plan 套餐可以选择？

当前订阅档位为 **Go**、**Explore** 和 **Build**。

| 档位 | 使用定位 | 额度（相对 Go） |
| :- | :- | :- |
| Go | 轻松上手 | 1 倍（基准） |
| Explore | 日常主力 | 3 倍 |
| Build | 高强使用 | 12 倍 |

***

<div id="switch-models" />

## M Plan 包含哪些模型？如何切换模型？

三个档位都包含 M3.1 Flash Preview 文本模型，以及图像与音频生成、网络搜索和多模态理解能力；Explore 和 Build 还包含 H3 视频模型，Go 不包含视频模型。如需切换模型，在 API 请求或工具配置中填写对应的模型 ID 即可。

***

<div id="minimax-code" />

## 在 MiniMax Code 中使用，需要单独订阅或配置 API Key 吗？

不需要。订阅 M Plan 后，登录 [MiniMax Code](https://agent.minimax.cn/download) 即可享有会员权益，无需另行订阅，也无需配置订阅 Key。只有在 Claude Code、Cursor 等其他 AI 工具中使用时，才需要配置订阅 Key。

***

<div id="multiple-tools" />

## 是否可以同时在多个工具中使用我的订阅套餐？

可以，您可以在所有支持的工具中使用同一订阅套餐，但额度是共享的，所有工具的使用会消耗同一套餐额度。

***

<div id="annual-billing" />

## 年付是否更优惠？

是的，年付只需支付 10 个月的费用，相当于省下 2 个月。

***

<div id="shared-usage" />

## 文本、图片、语音等额度是分开的吗？

不是。M Plan 包含的文本、图像、音频和视频等模型共用同一份额度，您可以根据需要灵活使用。需要注意的是，视频模型只受周窗口限制，不受 5 小时窗口限制。

***

<div id="usage-consumption" />

## 不同模型的额度消耗一样吗？

不一样。每次使用消耗多少额度，主要取决于以下几点：

* **模型与模态**：不同模型的消耗不同，其中视频生成的消耗远高于文本对话，经常生成视频时额度会用得更快。
* **上下文长度**：对话越长、读入的文件和代码越多，每次请求需要处理的内容就越多。
* **任务复杂度**：多步骤的 Agent 任务会连续发起多次请求并调用工具，消耗明显高于单次问答。
* **思考深度**：M3.1 Flash Preview 的思考深度越高，生成的思考内容越多，消耗也越多。

如果额度消耗较快，可以为简单任务选择较低的思考深度，适时开启新会话，并只引入完成任务所需的文件。

***

<div id="check-usage" />

## 如何查看 M Plan 用量？

您可以通过以下三种方式查看 M Plan 用量：

**方式一：访问套餐用量页面**

访问 [套餐用量](https://platform.minimax.cn/console/usage) 页面查看您的套餐、额度、积分和用量情况。

**方式二：使用 MiniMax CLI**

```bash theme={null}
mmx quota
```

**方式三：使用 API 接口查询**

```bash theme={null}
curl --location 'https://www.minimax.cn/v1/token_plan/remains' \
--header 'Authorization: Bearer <API Key>' \
--header 'Content-Type: application/json'
```

***

<div id="reset-calculation" />

## 用量是如何重置的？

M Plan 用量通过控制台用量进度条展示，并受额度窗口控制：

* **文本、图像、语音等非视频模型**：受 5 小时窗口和周窗口控制。
* **视频模型**：仅受周窗口控制，不受 5 小时窗口限制。
* **订阅周期**：未使用完的套餐内 M Plan 额度不会结转到下一个计费周期。
* **积分包**：自购买之日起 1 年内有效，不会因为套餐窗口刷新而重置有效期。

两个窗口都从首次使用时开始计时，5 小时窗口持续 5 小时，周窗口持续 7 天。窗口结束后，额度恢复到所在档位的上限，下一次使用时开启新窗口。两个窗口各自独立刷新，5 小时窗口刷新不会恢复本周已用的额度。

***

<div id="usage-rollover" />

## 没用完的额度会累积或结转吗？

不会。未用完的额度不会累积到下一个窗口，也不会结转到下一个计费周期。

***

<div id="usage-limit" />

<div id="deduction-order" />

## 额度用完后怎么办？

达到 5 小时或周额度上限后，您可以选择：

* **使用积分包**：扣费顺序为 M Plan 额度优先、积分包其次。达到上限后，超出的用量默认从积分包中扣除积分，按到期时间先后扣除。积分包仅用于其支持的模型与能力。
* **升级订阅套餐**：在 MiniMax Code 的「设置 > 用量与模型」中点击「升级」，或在 [套餐详情](https://platform.minimax.cn/console/plan) 页面点击 **管理订阅**，选择更高档位以获得更高的额度。
* **等待额度窗口重置**：文本、图像、语音等非视频模型受 5 小时窗口和周窗口控制；视频模型仅受周窗口控制，等待 5 小时窗口刷新不会恢复视频周额度。
* **切换到按量付费**：将工具中的订阅 Key 更换为普通开放平台 API Key，按资源实际使用量消耗账户余额。

***

<div id="account-balance" />

## 额度用完后，会自动扣账户余额吗？

不会。订阅 Key 只使用 M Plan 额度和积分包。额度用完后，如有积分包剩余，会继续扣除。账户余额只有在使用按量计费 API Key 时才会扣除。

***

<div id="m-plan-limits" />

## M Plan 有哪些使用限制？是否适合生产环境？

M Plan 面向个人开发者的交互式使用场景，更高的套餐等级提供更高的额度上限。生产环境建议使用按量付费。

主要限制包括：

* **速率限制（RPM / TPM）**：超出后会限流，通常约 1 分钟恢复，高峰期可能动态收紧。
* **文本、图像、语音等非视频模型额度**：受 5 小时窗口和周窗口控制。
* **视频模型额度**：仅受周窗口控制，不受 5 小时窗口限制。

***

<div id="traffic-rules" />

## 平台流量规则是什么？

为保障对所有用户的服务稳定性和可用性，MiniMax 平台可能在高峰时段实施动态限流策略。

为了避免少数异常流量挤占公共算力池，并保障大多数用户的稳定体验，平台会基于账户使用维度进行速率调控。同时，我们正在持续推进算力扩容与系统优化，努力提供更稳定、可靠的服务。

***

<div id="subscription-key" />

## 订阅 Key 是什么？

订阅 Key 是用于 M Plan 订阅套餐和积分包的 Key。您可以将它配置到 AI Agent 和编程工具中，调用可用的模型与能力。

订阅 Key 与普通按量计费 API Key 相互独立，不能混用。

***

<div id="get-subscription-key" />

## 在哪里获取订阅 Key？

订阅 M Plan 后，在 [套餐详情](https://platform.minimax.cn/console/plan) 中查看您的订阅 Key。请妥善保管，不要泄露给他人或提交到代码仓库。

***

<div id="supported-tools" />

## 订阅 Key 可以在哪些工具中使用？

支持 OpenAI 兼容或 Anthropic 兼容协议的 AI 工具都可以使用，例如 Claude Code、Codex、Cursor、OpenCode、Hermes Agent 和 OpenClaw，配置方法见各工具的接入教程。其他工具可参考 [其他工具](/docs/m-plan/other-tools) 中的通用配置。

***

<div id="api-key-interchangeable" />

## M Plan 的订阅 Key 和开放平台普通的 API Key 可以混用吗？

不可以。

* **订阅 Key**：用于 M Plan 套餐内额度和积分包。已有按量计费价格的 API 端点会按对应按量计费价格扣减套餐内 M Plan 额度。积分包可用于其支持的模型与能力，并承接订阅额度之外的合规超额用量。
* **普通开放平台 API Key**：用于按量计费访问开放平台 API 接口，按资源实际使用量计费，消耗您的账户余额。

***

<div id="balance-still-charged" />

## 为什么订阅后仍在扣账户余额？

通常是因为工具中填写的是按量计费 API Key。按量计费 API Key 会扣除账户余额，不会使用 M Plan 额度。请将工具中的 Key 更换为 [订阅 Key](https://platform.minimax.cn/console/plan)。

***

<div id="credit-pack-models" />

## 积分包支持哪些模型？

积分包支持 M3.1 Flash Preview 文本模型、H3 视频模型，以及图像与音频生成、网络搜索和多模态理解能力。

***

<div id="credit-pack-value" />

## 积分包怎么折算？

积分包是可单独购买的补充余额，用于覆盖 M Plan 资源范围内的合规超额用量。

* **1,000 积分 = ¥7**，与开放平台 API 按量付费目录价等值。
* 使用积分包调用已有按量计费价格的资源时，会按该资源的目录价折算为积分扣减。
* 如果同一用量同时可由套餐内额度和积分包覆盖，系统会优先扣除套餐内额度，超出部分再扣除积分包中的积分。
* **有效期**：积分自购买之日起 1 年内有效。

***

<div id="credit-pack-validity" />

## 积分包的有效期是多久？

积分包自购买之日起 1 年内有效，不会因套餐额度窗口刷新而重置有效期。

***

<div id="credits-without-plan" />

## 没有 M Plan 订阅也可以使用积分包吗？

可以。积分包可以单独购买，没有 M Plan 订阅也能使用。

需要注意：

* 如同时订阅了 M Plan，会先扣除套餐内额度，再扣除积分包。
* 在 MiniMax Code 中，单独购买积分包不享有 M Plan 会员权益，例如更多的云端项目、定时任务和网站发布额度，以及自定义域名、移除网站水印等。
* 如需使用 M Plan 覆盖范围之外的资源，请使用按量计费 API Key。

***

<div id="upgrade-plan" />

## 如何升级套餐？什么时候生效？

进入 [套餐详情](https://platform.minimax.cn/console/plan)，点击 **管理订阅** 并选择目标档位。当前订阅未使用的部分会按剩余时间折算抵扣，您只需支付差额。完成付款后，升级立即生效，之后按新档位的价格续费。详见 [升级与取消](/docs/m-plan/change-plan)。

***

<div id="downgrade" />

## 可以降级吗？

暂不支持降级。

***

<div id="monthly-to-annual" />

## 可以从月付改为年付吗？

月付订阅可以升级为更高档位的月付或年付；年付订阅可以升级为更高档位的年付。

***

<div id="cancel-renewal" />

## 如何取消自动续订？

您可以在 [套餐详情](https://platform.minimax.cn/console/plan) 中关闭自动续费。关闭后，当前周期内的权益照常可用，到期后不再续费。

自动续费关闭后无法恢复。如需继续订阅，请重新购买套餐；如果在当前订阅到期前升级，剩余价值可以抵扣费用。

***

<div id="token-plan-users" />

## 我是 Token Plan 用户，需要做什么？

无需任何操作，当前套餐与权益保持不变；只要自动续费保持开启，就会按原价格和原有权益续费。您也可以随时升级到 M Plan，当前订阅未使用的部分会折算抵扣，具体安排见 [Token Plan 老用户说明](/docs/m-plan/token-plan-notice)。

***

<div id="invoice" />

## 开票金额怎么计算？

可开票金额以订单实际支付的金额为准：

| 付款方式 | 可开票金额 |
| :- | :- |
| 支付宝 | 支付宝实际支付的金额 |
| 账户余额 | 余额实际扣除的金额 |
| 支付宝与余额组合 | 两种方式实际支付金额的合计 |
| 使用代金券 | 扣除代金券后实际支付的金额，代金券抵扣部分不可开票 |
| 升级订阅 | 实际补付的差额，原订阅抵扣部分不可开票 |

***

<div id="tps-calculation" />

## 语言模型的 TPS（Tokens Per Second）是如何计算的？

TPS 表示模型每秒生成的 token 数量，用于衡量模型的推理输出速度。计算公式为：

$$
\text{TPS} = \frac{\text{输出 token 数量}}{\text{最后一个 token 的生成时间} - \text{第一个 token 的生成时间}}
$$

即从模型输出第一个 token 开始计时，到最后一个 token 输出完成为止，期间生成的 token 总数除以这段时间（秒）。

<Note>
  TPS 在实际使用中可能存在波动，各页面上标注的 TPS 为参考值。
</Note>

***

<div id="error-401" />

## 登录后仍报 401 怎么办？

通常 `mmx auth login` 会根据 API Key 自动检测服务区域（若购买的是国内套餐，在进行 mmx auth login 配置时尽量不要开 VPN）。若仍报 401，大概率是 region 未自动匹配成功，可手动指定：

```bash theme={null}
mmx config set --key region --value cn       # 国内 API Key
mmx config set --key region --value global   # 海外 API Key
```

执行 `mmx auth status` 确认当前 region 与 Key 来源平台一致。
