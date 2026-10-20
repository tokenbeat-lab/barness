# 01 协议配置与地区身份

Status: resolved
Type: task
Blocked by: none

## Context

[规格](../spec.md)、[设计](../design.md)、[验证策略](../../../docs/agents/verification.md)。需求：R01 R05 R06 R10 R12。

## Ownership

ai/binding.go；三套协议配置及请求、解析、历史、用量与最终校验路径；新增 regional protocol E2E。真实 Provider 不改写，逐分支区分协议形状与身份语义。

## Acceptance

先在规格指定的公共 Client / Host / CLI 接缝建立失败场景，再实现；保留实际命令、退出码、时间、执行来源、工件及 Hash。先读取相关 ADR 和验证策略。不读取或输出真实秘密；不调用付费服务，不运行 race。按实际证明关闭实现，不用普通 replay 代替自身 live。

## Comments

2026-10-20：由 implement-spec 根据已提供规格建立；无 GitHub PR 关闭义务。各票完成后在 Answer 留实现/验证指针，整体状态唯一维护于 spec。


## Answer

已在 `codex/regional-protocols` 实现；实现提交 `7171b49`、`9865294`、`7883569`，并已合入集成分支当时 tip `637827d`（merge `4c98478`）。未改写实际 Provider 身份、正式目录或旧捕获来源。

- `ai/binding.go` 增加 ZAI / MiniMaxCN；三套既有能力配置拥有各自不可变实现选择，明确枚举四站；请求、解析、流顺序、工具 ID、reasoning、usage、最终 payload 与源历史语法按所属协议配置选择实现。
- 原生信任、目录索引、价格合同与结果归属继续比较实际 Provider。101 处原始生产声明/判断/上下文的分类和变更源码 SHA256 在 `.evidence/provider-sites/protocols/branch-audit.json`；原生信任的同租户/账户/型号/协议相同仍不能跨地区回放。
- 新 ZAI Messages 在 capability/options 阶段拒绝未接入选项，零凭据读取；Zhipu 的既有晚期选项阶段保留。新增服务正缓存写入缺价不得成功计为免费；保留写入用量和已报告的一小时拆分，不套用 Anthropic 原生 TTL 费率。
- 新 E2E `TestRegionalProvider*` 通过公共 Client；覆盖新增九身份四入口文本、六路线工具分片/Host 往返/原生 reasoning、相同来源/跨服务/未担保/轮换 key、ZAI unsigned thinking 信任、用量、HTTP/畸形/截断终态、取消阶段、队列背压、Hook/认证/能力守卫，以及许可/响应体/事件/调用资源归零。

普通验证事实集中于 `.evidence/provider-sites/protocols/index.json`（原件在仓库根目录外置证据树，不在临时 worktree）：

- `01b-text-red`：六新路线原厂默认与 Responses 终态丢正文，退出 1；`02-text-green` 退出 0。
- `04b-guards-failures`：ZAI Messages 选项错误仍读凭据，退出 1；`05-guards-green` 退出 0。
- `08-unpriced-red` / `10-write-lifetime-red`：ZAI 正缓存写入缺价或 TTL 被丢弃，退出 1；`09-unpriced-green` / `11-write-lifetime-green` 退出 0。
- `12-runtime-regression`：新路线和旧 GLM / MiniMax / Native / Anthropic / Chat / DeepSeek / Preflight / ProtocolTimeouts / EventQueue 定向普通回归退出 0，墙钟 35.408 秒，执行来源 `4c984782a2aa3fc0c02b3b298b8cd9477ddb6288`。
- `13-final-regional`：最终 158 个保留场景全部 PASS，退出 0；命令 `go test ./ai/e2e -run '^TestRegionalProvider' -count=1`，执行来源 `78835691eeb43a1344698daee0bf51a8818bd222`；原生工件带 request/script/events/Result/Attempt/usage/assertions/resource 记录与 Hash，TestMain 脱敏审计通过。
- `07-old-protocol-regression` 保留实际 FAIL：临时 worktree 不含原不可变 archive 的消费测试失败，不能以此回写旧证明；运行时场景无失败。集成票另在原 archive 所在根目录核对源码过期语义与证据消费者。初始工具编译错与首次失败测试 panic 亦保留原回执，不代替上述可审计红例。

边界：以上均为合成/受控离线协议证明，无真实 key 读取，无付费请求；新身份自身 live、正式晋升、race 与完整发布仍未执行，统一状态见 spec。建议人工 race 范围为上述新 `TestRegionalProvider*` 与受影响旧取消/原生续接/队列/共享生命周期；本票没有执行 race。根协调者统一维护 map，避免并行追加冲突。
