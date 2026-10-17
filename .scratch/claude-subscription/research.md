# Claude 订阅账号、API 与 Agent SDK：官方机制核查与 barness 方案

核查日期：2026-10-15（Asia/Shanghai）。范围：Anthropic 第一方文档和公开源码；未读取本机 Claude 凭据，未登录、调用付费 API 或验证特定账号资格。

标记：**确认**＝官方文字或公开代码直接支持；**推断**＝由已确认边界推出的工程判断；**未知**＝未找到公开契约，不能作为实现保证。

## 1. 结论：存在三条不同路径

| 路径 | 登录和凭据 | 实际扣哪一桶 | 对 barness 的意义 |
| --- | --- | --- | --- |
| Max/Team 月度 API credits | 将 Claude 计划关联到一个 Console organization；该组织的 API key | Console 组织 API 余额，先用赠送 credits | **确认**：可直接沿用 barness 的 Messages API/provider；无须 Agent SDK 或订阅 OAuth |
| Claude Code / Agent SDK 订阅运行 | Claude 账号 `/login`，或 `claude setup-token` → `CLAUDE_CODE_OAUTH_TOKEN` | Claude 计划的 subscription usage limits | **确认**：个人订阅可用于 SDK、`claude -p`；第三方产品登录权限见第 5 节的文档冲突 |
| Claude Platform Console OAuth / WIF | `ant auth login` 的 Console profile，或工作负载身份交换 | API 组织/workspace 的计费与限额 | **确认**：API 无静态 key 认证；不能等同于 Claude 订阅 usage limits |

2026-10-07 的帮助中心更新明确宣布 Max/Team 月度 API credits，并明确 Agent SDK、`claude -p`、third-party apps **仍能**使用订阅限额。因此“订阅账号能用于 API 与 SDK”在当前日期成立，但 API credits 与订阅限额是两种机制，不能把所有调用理解成“订阅无限调用普通 Messages API”。[计划更新](https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan)、[月度 API credits](https://support.claude.com/en/articles/17154008-monthly-api-credits-for-max-and-team-plans)

## 2. API credits：订阅权益进入 Console 余额

**确认**：Max 5x 每月 $100；Max 20x $200；Team Standard 每席 $20，Premium 每席 $100，合并为团队余额且每月上限 $500；Free、Pro、Enterprise 不符合该赠送资格。计划须有效且加入至少七天；移动端订阅也可在 Web 领取，无须为 Console 添加信用卡。[资格和额度](https://support.claude.com/en/articles/17154008-monthly-api-credits-for-max-and-team-plans)

领取是**账务绑定**：Max 订阅者或 Team Owner/Primary Owner，在 Claude Billing 中关联一个 Console 组织；还须具有该 Console 组织的 Owner/Admin/Billing 角色。一个计划只能关联一个组织，一个组织只能接收一个计划的 credits；变更关联组织需联系支持。[领取流程](https://support.claude.com/en/articles/17154008-monthly-api-credits-for-max-and-team-plans)

credits 覆盖 Messages、Message Batches、Playground、Managed Agents、Agent SDK，以及自行通过该组织 API key 启动的 `claude -p`。不覆盖交互式 Claude Code、额外订阅用量、Bedrock/Vertex/Foundry；GitHub Action、IDE、Desktop 发起的运行按 Claude Code 使用归类。通过 Claude plan 登录的 SDK/`-p` 继续用订阅限额，不消耗 API credits。[覆盖范围与 FAQ](https://support.claude.com/en/articles/17154008-monthly-api-credits-for-max-and-team-plans)

credits 每账期到账、账期末到期且不结转，先于购买余额消耗，组织成员共享；余额耗尽后如有购买余额或自动充值则继续，否则停止 API 请求直到下次到账，不自动转为 Claude 订阅扣费。Console 的 workspace spend limits 可以限制项目用量。[余额规则](https://support.claude.com/en/articles/17154008-monthly-api-credits-for-max-and-team-plans)

**推断**：barness 不需要调用“领取 credits API”，也不应根据 token 字符串猜测 credits。运行时仍是正常 API 认证、usage 与组织账务；赠送/购买余额由 Anthropic 服务端结算。帮助中心明确给出的操作路径是 linked organization 的 API key；Console OAuth 同属 API 身份体系，但该帮助页未单列其 credits 结算示例。

## 3. 普通 API 的 OAuth：Console profile，不是订阅登录

### 3.1 公开的登录与请求机制

**确认**：`ant` 是 Anthropic 官方 Claude Platform CLI。`ant auth login` 对 **Claude Console** 发起浏览器 OAuth，不创建 API key；`--no-browser` 支持打印 URL 后粘贴 code；可以用 `--profile` 和 `--workspace-id` 显式选择目标。官方将交互式登录定位于个人机器开发/脚本，服务器、CI、容器使用 WIF。[CLI 认证](https://platform.claude.com/docs/en/cli-sdks-libraries/cli/authentication)、[官方仓库](https://github.com/anthropics/anthropic-cli)

源码的 Console 登录采用 Authorization Code + PKCE S256 和随机 `state`：浏览器访问 `https://platform.claude.com/oauth/authorize`，通过 loopback callback 或人工粘贴取 code，再调用 `https://api.anthropic.com/v1/oauth/token`。默认 scope 为 `user:profile user:inference user:developer`；token 响应提供 access/refresh token、`expires_in`、实际 scope、组织/账号/workspace 信息。生产 client ID 属于官方 CLI，公开源码中的 ID 不等于第三方应用注册授权。[固定版本源码：登录](https://github.com/anthropics/anthropic-cli/blob/66793ac5e47445019334a560776407ece22ba098/pkg/cmd/cmd_auth.go#L29)

模型调用保持 `POST /v1/messages` 和相同 JSON/SSE 协议。OAuth token 放在 `Authorization: Bearer`，还应发送 `anthropic-beta: oauth-2025-04-20`、`anthropic-version: 2023-06-01`；不能仅把 token 塞入原 `x-api-key`。官方 API skill 明确指出 Messages 需要 OAuth beta header。[第一方 HTTP 示例](https://github.com/anthropics/skills/blob/main/skills/claude-api/shared/anthropic-cli.md)、[API 概览](https://platform.claude.com/docs/en/api/overview)

```sh
# 人工一次性登录；这会写入该 profile 并可能激活它。
ant auth login --profile barness-dev --workspace-id wrkspc_example
ant auth status

# 后续由本地宿主程序捕获 stdout，绝不能打印进日志。
ant auth print-credentials --profile barness-dev
```

**确认**：当前 `ant auth login` 没有 `--subscription`、`--billing` 或 `--claude` 这类公开切换标志；它是 Console OAuth。`--scope org:admin` 是组织管理权限，普通模型调用不需要，且官方要求与日常 profile 分开。[CLI 认证](https://platform.claude.com/docs/en/cli-sdks-libraries/cli/authentication)、[命令标志源码](https://github.com/anthropics/anthropic-cli/blob/66793ac5e47445019334a560776407ece22ba098/pkg/cmd/cmd_auth.go#L82)

### 3.2 存储、刷新与凭据快照

配置目录依次为 `$ANTHROPIC_CONFIG_DIR`、Linux/macOS `~/.config/anthropic`、Windows `%APPDATA%/Anthropic`。`configs/<profile>.json` 存认证模式、client ID、org/workspace/base URL；`credentials/<profile>.json` 是秘密，保存 access token、refresh token、expiry，SDK 写入模式为 `0600`。profile 的目标配置和 token 实际获授的组织/workspace 是两种信息，不能以修改配置代替重新授权。[profile 契约](https://platform.claude.com/docs/en/manage-claude/wif-reference#profile-configuration-file)、[SDK 写入格式](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/config/writers.go#L24)

`print-credentials` 的三个输出模式如下；均读取所选 profile，环境里的 `ANTHROPIC_API_KEY` 不会改成输出 API key。[第一方说明](https://github.com/anthropics/skills/blob/main/skills/claude-api/shared/anthropic-cli.md)、[实现](https://github.com/anthropics/anthropic-cli/blob/66793ac5e47445019334a560776407ece22ba098/pkg/cmd/cmd_auth.go#L1069)

| 模式 | 输出 | 集成注意 |
| --- | --- | --- |
| 默认 | 完整 credentials JSON | 包含 `refresh_token`；仅在内存解析，不能原样日志/返回前端 |
| `--access-token` | 裸 access token | 无 expiry/org/workspace，无法单独完成完整快照校验 |
| `--env` | `ANTHROPIC_AUTH_TOKEN`，以及可选 `ANTHROPIC_BASE_URL` | 不含 workspace/expiry；不要对不可信输出使用 shell `eval` |

默认 JSON 字段包括 `version`、`type: oauth_token`、`access_token`、`expires_at`（Unix **秒**）、可选 `refresh_token`、`scope`、`organization_uuid`、`organization_name`、`account_email`、`workspace_id`、`workspace_name`。这些实际 token 元数据适合验证账号与租户绑定；邮箱不应进入通用日志。[公开类型及 JSON 映射](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/config/writers.go#L55)

**关键确认**：`print-credentials` 在 access token 距过期不超过 120 秒时按需刷新。刷新请求是 JSON `grant_type=refresh_token`、`refresh_token`、`client_id`，发到 `/v1/oauth/token`，携带 OAuth beta；新 refresh token 如存在就替换，更新结果写回 credentials 文件。具体有效期取服务器 `expires_in`，没有应硬编码的统一 access-token 时长。[刷新实现](https://github.com/anthropics/anthropic-cli/blob/66793ac5e47445019334a560776407ece22ba098/pkg/cmd/cmd_auth.go#L1069)

**关键陷阱**：该命令刷新失败时通常只向 stderr 警告，随后仍输出旧 token 并成功退出，旧 token 可能已经过期；缺少 refresh token 也会警告后继续输出。因此宿主必须校验 `expires_at`、必需 scope、实际 org/workspace、剩余有效时长，不能只看 exit code。裸 env token 不会自动刷新；profile/SDK 路径才会刷新。[刷新失败行为](https://github.com/anthropics/anthropic-cli/blob/66793ac5e47445019334a560776407ece22ba098/pkg/cmd/cmd_auth.go#L1084)、[Admin token 导出说明](https://platform.claude.com/docs/en/manage-claude/admin-api#oauth-bearer-token)

### 3.3 优先级与 Go SDK 的隐藏重发

官方 Client SDK 的大类顺序是：显式构造参数 → `ANTHROPIC_API_KEY`/`ANTHROPIC_AUTH_TOKEN` → 指定 `ANTHROPIC_PROFILE` → WIF 环境 → active/default profile。空字符串环境变量仍可能占优先位；切换 profile 前应真正 unset 不用的凭据。不能依赖同时设置 key/token 的隐式胜出行为。[优先级](https://platform.claude.com/docs/en/manage-claude/wif-reference#credential-precedence)

**确认**：barness 已锁定的 `anthropic-sdk-go v1.75.0` 具有 `option.WithProfile`、`option.WithConfig` 和动态 auth middleware。`WithAuthToken` 本身只设置 Bearer，profile middleware 才会附加 OAuth beta。动态 middleware 遇到 401 会使缓存失效、强制更新 token、再次 `next(req)`；这条隐含重发独立于模型请求的 `WithMaxRetries(0)`。[配置入口](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/option/requestoption.go)、[认证中间件](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/internal/auth/middleware.go)

**推断**：一个显式管理 Attempt、重试、账务身份快照的宿主不能直接透传 profile middleware；应在 Logical Call 建立前完成认证解析/刷新，向模型适配器交付固定 token 和固定 headers。认证刷新请求和模型请求是不同操作，需分别观测，不能让 SDK 偷偷换身份重发模型请求。

## 4. Agent SDK 的订阅路径：沿用 Claude Code 运行时

**确认**：Agent SDK 提供 Python/TypeScript，不是普通 Messages API 的薄包装；它启动并监督 `claude` 子进程，用 stdio 通信，子进程拥有工具循环、shell、cwd、JSONL 会话记录。Go 若要驱动同一个 agent，应通过 `claude -p --output-format json` 等 CLI 接口；引入它会引入第二套 agent runtime。[SDK 边界](https://code.claude.com/docs/en/agent-sdk/overview)、[子进程与部署](https://code.claude.com/docs/en/agent-sdk/hosting)

**确认**：订阅认证默认可用 Pro、Max、Team、Enterprise 的 `/login` OAuth。自动化可运行 `claude setup-token`，经与 `/login` 相同的浏览器授权流程生成**一年期** OAuth token；命令只打印，不自动保存。将其设为 `CLAUDE_CODE_OAUTH_TOKEN` 后，SDK 和脚本可进行订阅模型请求；token 不提供 Remote Control 或 claude.ai connectors，用户本地配置的 MCP 仍可用；`--bare` 不读取该变量。[订阅 token 官方说明](https://code.claude.com/docs/en/authentication#generate-a-long-lived-token)

```sh
# 人工运行并把输出保存到用户自己的秘密管理器；不要提交到项目。
claude setup-token

# 执行环境从秘密管理器注入 CLAUDE_CODE_OAUTH_TOKEN 后：
claude -p 'Reply with a short greeting.' --output-format json
```

`CLAUDE_CODE_OAUTH_TOKEN` 的环境值在会话内固定；替换过期 token 需重新生成并重启。官方还公开 `CLAUDE_CODE_OAUTH_REFRESH_TOKEN` 配合必填 `CLAUDE_CODE_OAUTH_SCOPES`，用于让 `claude auth login` 直接交换 refresh token；后者是 Claude Code 的认证供应方式，不是普通 Go SDK 的同名能力。`CLAUDE_CODE_SESSION_ACCESS_TOKEN` 未在本次核查的当前官方文档找到稳定契约，不能将其作为 barness 公开配置项。[环境变量](https://code.claude.com/docs/en/env-vars)

Claude Code 与 Client SDK 的认证顺序不能混用：Claude Code 大体是显式云 provider → `ANTHROPIC_AUTH_TOKEN` → `ANTHROPIC_API_KEY` → `apiKeyHelper` → `CLAUDE_CODE_OAUTH_TOKEN` → 显式 Anthropic profile/WIF → `/login`；隐式 `user_oauth` active profile 又排在有效 `/login` 之后。Console profile 切换可能退出原订阅登录；Desktop/cloud/bare 又有专门规则。必须通过 `/status` 验证实际来源，不能假定“账号已登录”即一定走订阅。[完整优先级及 profile 例外](https://code.claude.com/docs/en/authentication#authentication-precedence)

**确认**：SDK 输出迭代默认交付完整 block；开启 `includePartialMessages`/`include_partial_messages` 后，收到包裹原始 `message_start`、`content_block_delta`、`message_delta` 等 API SSE 事件的 `stream_event`。`result`、assistant、tool/session 事件属于 agent 运行协议，不能作为单次模型流直接混入 barness 的 provider 事件。[流式输出](https://code.claude.com/docs/en/agent-sdk/streaming-output)

## 5. 第三方订阅登录：当前第一方文档有明确张力

| 第一方页面 | 本次读取内容 | 能确认什么 |
| --- | --- | --- |
| 帮助中心，明确日期 2026-10-07 | SDK、`claude -p`、third-party apps 仍可用订阅 limits；6 月变更已经暂停 | 不能引用早期变更宣告，声称订阅 SDK 已全面禁用 |
| 当前 SDK overview 和 quickstart | 未经 Anthropic 预先批准，不允许第三方为产品提供 claude.ai login/rate limits，推荐 API key | 不能据可运行 token 推导任何第三方 SaaS 都能收集订阅 OAuth |
| 当前 Code Authentication | 公开 setup-token、CI/scripts、订阅计划与 token 权限 | 个人订阅自动化具备官方文档路径 |
| 当前 Code Legal and compliance | 允许平台托管原版 Code，每位终端用户自行认证并直接计费；禁止收集或中介 claude.ai 凭据 | 托管官方 binary 与产品自建订阅 OAuth 登录具有不同边界 |

来源：[帮助中心](https://support.claude.com/en/articles/15036540-use-the-claude-agent-sdk-with-your-claude-plan)、[SDK overview](https://code.claude.com/docs/en/agent-sdk/overview)、[SDK quickstart](https://code.claude.com/docs/en/agent-sdk/quickstart)、[Code Authentication](https://code.claude.com/docs/en/authentication)、[原版 Code 托管条件](https://code.claude.com/docs/en/legal-and-compliance#can-customers-offer-claude-code-in-their-products)

**判断**：这几页不必被解释为“技术能力互相否定”：订阅者使用自己的 SDK/脚本，与第三方产品被授权提供统一 claude.ai 登录入口，可以是不同范围。但“third-party apps 仍可用 limits”与 SDK 的广泛禁止措辞存在未被完整解释的重叠。官方未给出这几页一致的第三方授权说明，不能擅自宣布所有产品均获准，也不能用未注明更新日期的通用警告抹掉 10 月更新。

**未知/实现边界**：未找到对 barness 这类独立第三方产品公开开放的 Claude 订阅 OAuth client 注册、redirect URI 注册、订阅账户授权代理或 SaaS 多租户 BYO-subscription 契约；未找到官方承诺可直接把 `claude setup-token` token 注入任意原生 Messages API 客户端并消耗订阅限额的完整接口规范。因此不要自行使用 Anthropic 第一方 client ID 复制订阅登录，也不要把“Console OAuth 已公开”包装成“订阅 OAuth 已全面开放”。这不阻塞 API credits + API key 的方案。

## 6. 云端、多租户与计费观测

**确认**：官方建议生产服务/CI 使用独立 service account 或 WIF，而非共享个人 key/浏览器 profile。WIF 通过上游 OIDC JWT 在 `POST /v1/oauth/token` 换取短期 Anthropic Bearer；scope 可选 `workspace:inference`（模型推理）或 `workspace:developer`（更广 API），权限取 scope 与 service-account role 的交集；token lifetime 为 60–86400 秒，默认 3600 秒。这是 API 组织身份，不是 Claude 订阅令牌。[认证选型](https://platform.claude.com/docs/en/manage-claude/authentication)、[WIF 参考](https://platform.claude.com/docs/en/manage-claude/wif-reference)

API 响应提供 `request-id`、`anthropic-organization-id`、`anthropic-workspace-id`；workspace 是调用身份/限额/计费边界，模型配置中的 workspace 字符串不能改变已签发 token 的授权。[API 响应和 workspace](https://platform.claude.com/docs/en/api/overview)、[CLI workspace 绑定](https://platform.claude.com/docs/en/cli-sdks-libraries/cli/authentication)

**确认**：Agent SDK 的 `total_cost_usd`/`costUSD` 是客户端价格表估算，不是实际账单；不能据它判断赠送 credits 余额或订阅限额已扣多少。并行工具消息按 API message ID 去重；`result.usage` 不含 subagents，`modelUsage` 和 `total_cost_usd` 包含；恢复会话可能包含历史累计成本，不能逐条相加。正式账务读 Console/Usage and Cost API。[成本语义](https://code.claude.com/docs/en/agent-sdk/cost-tracking)

SDK 多租户部署要求分别隔离 cwd、`CLAUDE_CONFIG_DIR`、settings/memory、凭据和 egress；默认共享磁盘配置可能泄露别的租户上下文。其宿主模型是每个活跃 agent 一个子进程，非无状态 HTTP client。[部署与隔离](https://code.claude.com/docs/en/agent-sdk/hosting#multi-tenant-isolation)

**推断**：barness UI/审计应分开记录认证来源、API 组织/workspace、实际 usage、估算 USD、结算来源的已知程度；只知道 API key/OAuth 不足以断言调用全额使用赠送 credits。凭据 ID/用户 ID 只能用于内部关联，凭据内容和完整 HTTP headers/body 不能作为验收产物。

## 7. 尚待账号级验收的内容

- 本次不确认特定用户的 credits 是否已到账、Console 角色、rollout、模型权限或剩余额度；官方也提示分批开放。
- 当前帮助中心不公开“订阅 remaining limits API”作为第三方稳定接口；不要依赖抓取 Claude 私有接口。
- profile/SDK 自动刷新和 `print-credentials` 的行为已由源码确认，但未在用户账号上执行。跨进程同时刷新/轮换令牌的协调不能仅靠原子文件写入推断完成。
- SDK 权限并不替代 barness 的 tenant/user 身份、授权、工具隔离、取消、超时与计费上限。其商业使用和订阅产品接入是不同问题。

## 8. 基于 barness 的实现方案（设计建议，尚未实施）

本节依据本仓库当前源码和 ADR；不把“支持订阅余额”建模成新的 Provider 或 API 协议。建议按以下顺序实施，前三条路线各自有独立验收结论。

| 路线 | 业务目标 | barness 落点 | 建议 |
|---|---|---|---|
| A：Console API key + 月度 credits | 原生 Go Agent 使用 Max/Team 赠送 API 余额 | 现有 Anthropic Messages adapter | 先交付；基础协议通常无须改动 |
| B：ç profile | 原生 Go Agent 本地免 API key 登录；仍走 Console 余额 | Host 的 CredentialResolver + 小范围认证扩展 | 有免 key 需求时实施 |
| C：Claude 订阅计划限额 | 由用户自己的订阅运行官方 Claude Code harness | Host 层外部 Agent runner | 独立可选能力；保留官方登录流程 |
| D：订阅 token 直连自研 Messages adapter | 自研 Go harness 消耗 Claude 计划限额 | 需要独立协议与授权证据 | 当前不作为默认实现承诺；公开技术能力不能替代产品授权 |

A 的适用性来自 [月度 API credits 官方说明](https://support.claude.com/en/articles/17154008-monthly-api-credits-for-max-and-team-plans)；B 的认证目标来自 [ant 认证文档](https://platform.claude.com/docs/en/cli-sdks-libraries/cli/authentication)；C/D 的区别见 [官方产品集成与凭据边界](https://code.claude.com/docs/en/legal-and-compliance#can-customers-offer-claude-code-in-their-products)。这不是对未经验证的 D 路线宣告技术不可行，而是避免把它与已有明确文档的 A/B/C 混为一谈。

### 8.1 当前代码与架构约束

- 现有模块是 barness-ai；仓库尚无完整 Agent/账户服务。Host 才是身份、会话和工具执行的责任主体。[术语](../../GLOSSARY.md)、[ADR-0001](../../docs/adr/0001-barness-ai-protocol-boundary.md)。
- [Binding / Credential](../../ai/binding.go) 只支持 AuthAPIKey 和 Credential.APIKey；[preflight](../../ai/preflight.go) 拒绝其他认证类型、空 key 和不一致快照。
- [Anthropic adapter](../../ai/anthropic.go) 已用官方 Go SDK v1.75.0，显式构造服务、发送 X-Api-Key、自行解析 SSE、关闭 SDK 普通重试；不自动读取用户环境或 profile。[go.mod](../../go.mod)、[ADR-0011](../../docs/adr/0011-barness-ai-anthropic-messages-adapter.md)。
- [Request Callback 的 header guard](../../ai/hooks_run.go) 禁止回调添加/修改 Authorization、X-Api-Key；因此不能通过 TransformHeaders 把 OAuth 偷渡成 API key。
- 一次 Logical Call 固定 Binding、Credential 和账户；网络重试也固定同一快照，流开始后不得重放。OAuth 刷新必须位于 pin 之前。[ADR-0003](../../docs/adr/0003-barness-ai-snapshot-consistency.md)、[ADR-0006](../../docs/adr/0006-barness-ai-retry-policy-on-binding.md)。
- 已有 [本地装配示例](../../ai/examples/localassembly/localassembly.go) 可显式从指定环境变量或文件读取 key；云端可直接实现两个 resolver，已有 AccountScope、Actor、Admission、Observer 等接缝。

### 8.2 路线 A：最小交付，先用现有协议获得订阅价值

1. 用户在 Claude 账单页将 Max/Team 计划关联到 Console organization，并领取 credits。此操作由账户持有人在官方 UI 完成。
2. 在该组织创建**绑定到目标 workspace 的 API key**，避免新式跨 workspace key 所要求的额外请求头；云端共享工作负载优先使用 service account 的 workspace-scoped key。[API 认证与 workspace](https://platform.claude.com/docs/en/manage-claude/authentication)。
3. Host 按租户保存 key 的秘密引用；Binding 仍设 ProviderAnthropic、APIAnthropicMessages、OperationChat、AuthAPIKey，Endpoint 为 https://api.anthropic.com，AllowedModels 显式列举已入目录且获准使用的型号。
4. 本地直接使用 localassembly.Config.Key.EnvVar 或 File；核心库不隐式读取 ANTHROPIC_API_KEY。云端使用 TenantID + ActorID 授权 resolver。
5. AccountScopeID 由 Host 为实际 Console 组织/workspace 维护稳定映射。它不等于 TenantID、profile 名或 CredentialVersion；同账户的 key 轮换不改变作用域。
6. 保留现有 Usage 与 Estimated Cost。credits 只是支付来源，不能把 token 用量或模型成本置零；API 响应不能证明本次扣的是赠送、购买还是账期余额。UI 显示“Console 余额；估算成本”，实际余额以官方 Console 为准。[用量语义](../../GLOSSARY.md)。

**交付内容**：新增一份真实账户装配说明和单次低预算 live evidence；若账户满足条件，现有 adapter 即可工作。没有理由为领取 credits 新增 OAuth、Agent SDK、Provider ID 或协议代理。

**限制**：API credits 用尽后，服务端可能继续扣购买余额或 auto-reload；如产品要求“只花赠送余额”，Host 估算无法保证该属性，须结合组织的计费设置和 workspace spend limits，不能宣称有一个已文档化的“仅 credits”请求参数。回滚停用此 Binding，不能自动切换到另一个收费账户。

### 8.3 路线 B：Console OAuth 的原生 Go 接入

只在确有“本地免 key 登录”需求时增加。此处 user_oauth 是 Console 账户，不是提取 Claude 订阅登录。

职责按数据流分开：

1. 用户通过官方 ant auth login --profile barness-local --workspace-id wrkspc_... 完成浏览器登录；登录、scope 授权和 refresh token 持久化归官方 CLI。
2. 新建独立本地装配示例 ai/examples/anthropicoauth/：实现 CredentialResolver，显式指定可信 ant 可执行文件、配置目录和 profile；使用 exec.CommandContext 的参数数组调用 auth print-credentials，stdout/stderr 有界、超时可取消，不经 shell。
3. 默认 JSON 只在内存解析，取 access_token、expires_at 和可用的账户元数据；refresh_token 不传入 ai，不写日志、证据或错误。print-credentials 可能刷新失败仍 exit 0，Host 必须检查到期时间，必要时让用户重新登录；不能仅凭退出码判定成功。
4. Host 将 profile 的授权关系绑定到 Tenant/Actor 与稳定账户作用域。profile 配置和凭据返回的组织/workspace 不一致时拒绝；元数据不足时不得从 profile 名或 token 前缀猜归属，沿用受控的账户登记或要求重新登记。
5. Host 在解析期间完成必要刷新，签发版本化、有效期明确的 Credential Snapshot；完成 BindingVersion 一致性检查后进入 pin。之后的推理请求只用此静态 token。
6. ai 发送同一 Messages 请求与 SSE，新增的只有认证方式和必需认证 beta；Transcript、tools、thinking、Native State 继续走现有转换。

建议公开类型为**增量字段**，保持现有 API key 外部契约；以下为拟定形状，非已存在 API：

~~~go
const AuthOAuthAccessToken AuthKind = "oauth_access_token"

type Credential struct {
    // 既有身份、版本、Active 与 APIKey 字段保留。
    AccessToken Secret
    ExpiresAt   time.Time // OAuth 必填；零值不表示永不过期。
}
~~~

- AuthOAuthAccessToken 首期仅允许 Anthropic × Messages × Chat；按 Binding.AuthKind 校验恰好一种秘密，拒绝 APIKey/AccessToken 同时设置、token 缺失、未知认证类型和过期快照。
- preflight 用调用剩余时间与配置的安全余量判断 token 是否足够进入请求；等待 Admission、退避后到实际发送前也检查有效期，过期则失败，不能在 Attempt 里换 token。
- 对任意 stream 不承诺 token 在途永不过期；认证仅在建连检查还是会被服务端中途撤销必须实测，stream 失败保留已发生的 Usage，不自动重放。
- 内部 adapterCall 的 apiKey 可以直接重命名/替换为带 AuthKind 的私有认证值；httpFailures 的秘密脱敏应覆盖实际发送的 key/token。内部重构删除旧命名，不保留 shim。
- OAuth 请求仅有 Authorization: Bearer；不再发送 X-Api-Key。沿用 anthropic-version: 2023-06-01，并在全部 header/payload 回调之后合并必需的 oauth-2025-04-20；可选 thinking beta 仍按模型生成。单独 WithAuthToken 并不负责增加 OAuth beta。[Go SDK 静态认证选项](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/option/requestoption.go)、[profile 认证中间件](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/internal/auth/middleware.go)。
- 如确需显式 workspace header，在 Binding 中增加受限的 Anthropic workspace 配置，并将 Anthropic-Workspace-Id 加入 guardedHeaders；不开放任意认证 header map，也不从 AccountScopeID 反解析厂商 ID。
- 保留显式 Endpoint 与凭据目标校验；OAuth 首版只允许官方 API origin。token exchange 不可被 profile 中未核对的 base_url 或外部请求重定向。
- 401 返回认证失败；Host 可更新凭据后以新 RequestID 开始新 Logical Call。不能在旧调用中重新 resolve。只有未产生输出的、Host 明确允许且预算内的新调用才有重启条件。

**不能直接套官方 profile 中间件**：v1.75.0 的 NewMiddleware 在 401 后会 ForceToken 并再次 next(req)，不受普通 WithMaxRetries(0) 控制。这会让一次 barness Attempt 藏入两次网络请求、改变凭据版本，冲突于 ADR-0003/0006。WithProfile/WithConfig 的动态身份解析也绕过现有快照边界，因此仅在 Host 登录工具链中使用，不直接注入 ai adapter。[固定版本源码](https://github.com/anthropics/anthropic-sdk-go/blob/v1.75.0/internal/auth/middleware.go)。

**并发和生命周期**：本地每份 profile 的获取/刷新串行化并允许等待者取消；避免另一个 CLI 进程同时刷新同一 refresh token。原子写文件只保证文件不撕裂，不等于跨进程刷新互斥。首版采用单 owner 的专用 profile。云端生产按官方建议使用 service account key 或 WIF；若未来确需 OAuth broker，再设计集中刷新、锁与 fencing、原子轮换及撤销，不把本地 profile 文件当分布式凭据服务。

### 8.4 路线 C：使用订阅计划限额的官方 harness

如真实目标是“使用 Pro/Max/Team/Enterprise 的计划限额”，建议提供 Host 层的可选 Claude Code runner，运行 Anthropic 发布的原版二进制。Go 用子进程即可；无须为了认证引入 Python/TypeScript bridge。[Agent SDK 对比与其他语言方案](https://code.claude.com/docs/en/agent-sdk/overview)。

- 用户在官方 CLI 中完成登录；无人值守可采用官方 setup-token 与 CLAUDE_CODE_OAUTH_TOKEN，按实际到期时间安排更换。优先让原版 binary 管理登录，barness 不采集/转存用户的 claude.ai refresh token。
- runner 是一个完整外部 Agent 作业：拥有自己的工具、权限、上下文压缩和 session，不伪装成 ai.Client 的“一次 Generation Turn”，也不包装成新的 APIAnthropicMessages adapter。
- 用 claude -p 的结构化输出接收 result；实时输出采用官方 stream-json 模式，并按官方要求启用必要的 verbose/partial 选项。首次实现需固定 CLI 版本和 schema。[CLI 自动化](https://code.claude.com/docs/en/headless)。
- Host 启动前选择一种计费来源，构造独立进程环境，避免意外继承 ANTHROPIC_API_KEY、ANTHROPIC_AUTH_TOKEN 或显式 profile 改变计划限额路径；不全局修改环境变量。订阅模式不能启用会跳过订阅认证的 --bare；用受控 cwd、独立配置目录与沙箱限制自动发现的项目 hooks/MCP/插件。状态展示依据 binary 的可确认信息，未知时明确显示未知。
- 每个用户的配置、会话目录、工作目录和工具权限独立；云端使用独立 OS 身份或容器/沙箱。多租户共享进程中的环境切换不能构成隔离。
- 限制 stdout 队列/单事件/总输出，持续 drain stderr，取消时终止整个子进程组并回收；工具审批与文件访问由 runner 的宿主授权管理。
- sdk/CLI 的 session_id 属于外部作业，只能由同 Tenant/Actor 恢复；不能作为 barness-ai Native State、RequestID 或工具调用 ID。
- SDK result 的 total_cost_usd 作为其报告的成本估算单独记录，不转换成订阅剩余额度或用户实际扣款；宿主不得在 429、额度不足、登录过期时偷偷转到 API key。[成本与用量](https://code.claude.com/docs/en/agent-sdk/cost-tracking)。

产品化需同时满足官方“原版 binary + 每位终端用户自行认证/直接计费”的条件，并处理目前 overview 与支持页之间的表述差异；不据此创建 barness 自己的 claude.ai OAuth 登录入口。[公开边界](https://code.claude.com/docs/en/legal-and-compliance#authentication-and-credential-use)。

### 8.5 实施拆分、验证与回滚

| 阶段 | 主要变更 | 完成依据 |
|---|---|---|
| P0：验证 credits | 账户领取与 workspace-scoped key；现有 Host 装配说明 | 一次真实 Messages streaming + 工具往返，Console 侧核对用量；无核心改动 |
| P1：Console OAuth（可选） | ai/binding.go、preflight.go、adapter.go、call.go、anthropic.go、http_failures.go；独立本地装配示例；必要时 hooks_run.go | 先写失败场景和公共接口 E2E，再实现；证明固定快照和无隐藏重试 |
| P2：官方 harness（可选） | Host 层外部 runner；当前不存在完整 Agent 层，先定义 Job/权限/session 边界 | 自有任务 E2E、审批/取消/隔离，以及真实订阅 smoke；不计入原 Messages 支持结论 |
| 后续云端认证 | 按真实需求接入 service account/WIF | 另立范围，不与 P1 的本地 profile 混做 |

P1 编码前先列出并实现这些失败验收：错误 Tenant/Actor/profile、组织/workspace 不符、双凭据、OAuth 误用于其他协议、未知/过期 expiry、CLI 不存在/超时/非零退出、刷新失败却 exit 0、凭据轮换竞争、取消刷新、并发刷新、401 隐藏重发、429 退避后 token 过期、流中断、回调删认证 beta/改 workspace、恶意 base_url/重定向、token 被上游错误回显。沿用 [公共 Client + loopback Provider 的 E2E](../../ai/e2e/anthropic_suite_test.go)，不在实现后补镜像单元测试。

复用现有 evidence/audit 写脱敏 requests、events、result、attempts、assertions、manifest；OAuth token、refresh token、原始 CLI JSON/stderr 和个人目录均不得落盘。真实 smoke 使用专用账户和低预算，分别记录 API-key/Console-OAuth/官方-harness 的认证路线、SDK/CLI 版本、计费模式声明与 PASS/FAIL/NOT_RUN，不能用旧 API-key 通过结果替代 OAuth 证据。[E2E 证据入口](../../ai/e2e/main_test.go)。

观测在 Host 单独记录 credential_resolve/refresh 的耗时、阶段与固定错误码，关联 RequestID；刷新网络请求不冒充模型 Attempt。barness-ai 的 Observer 仍只报告原有无秘密元数据。退出登录/撤销拒绝新调用；已经发送的请求不能承诺撤回。回滚关闭对应 Binding 或 runner；保留已有消耗，重新启用用新调用读取新快照。

实现前为 P1 的 AuthKind/必需认证 beta、P2 的外部 harness 边界各补一份 proposed ADR；明确是扩展现有契约，**不修改已接受 ADR-0003 的在途快照规则**。本次仅写研究与方案，未创建 accepted ADR、未修改生产代码或运行付费请求。

### 8.6 尚需真实环境确认的事项

1. 当前订阅是否符合 credits 领取资格、Console 关联是否成功，以及目标 workspace/model 的权限：官方规则不能代替用户账户状态。
2. 计划限额与 API credits 的服务端选择没有已确认的统一请求级开关；不得从 token 前缀、SDK 名称或 Usage 推断支付来源。
3. Console profile 刷新后的账户元数据保留、多个进程的 refresh rotation 行为、实际 token 有效期与流式认证边界，必须锁定 ant/SDK 版本实测。
4. 第三方自研 harness 直接使用订阅 token 的完整授权/公开客户端注册仍缺乏无歧义官方保证；若要走路线 D，需要补齐具体授权和官方协议证据，而不是复制 Claude Code client ID、UA 或私有请求形状。
