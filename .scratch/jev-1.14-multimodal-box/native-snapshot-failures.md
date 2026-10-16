# 01 — TypeSafe native 快照重构前的失败清单

起点：`6b893c4c7e26b7d84092644ae8f4fe7aaeb84d3b`，当前分支 `main`。
在生产实现修改前记录；公开接缝由 design §8 / issue 01 已确认：
`Client.Classify`、`HookedClient.Classify`，真实运行时连接受控 HTTP Provider。
本票是等价重构，先绿的公开回归就是基线，不人为制造红灯。

## 架构与数据流

Host 公共请求在 scope/capability 边界校验并复制；TypeSafe adapter 独立编码
native 请求；headers → payload 回调后，仅一次严格 native freeze 形成独立拥有
State、最终题集和原生 criteria 的快照；sendUnary 复用固定 bytes/headers/凭据；
usage/ResponseModel 先登记，所有答案依据最终 native 题集验证后原子发布。
公共 JSON decoder 只负责 Host 请求，删除 native→公共请求的解码模式。
既有有限的回调借用值检查仍在序列化前执行，不成为第二个语义 guard。
不新增公共接口、动态注册、图片能力或 Host/Agent 依赖。

## 必须冻结的失败方式

| 失败方式 | 公开回归与期望 |
| --- | --- |
| 窄化 State/说明、空 State、null choice、单候选、可省略 bool criteria | `TestJevRegression` 两入口、`TestClassifierMixedFixtures`；原 wire、答案与用量不变 |
| 大整数/极长小数或指数经过 float64、图例只比文本 | `TestClassifierNativeLegend`、numeric-precision；精确值、等价十进制、对象无序/数组有序不变 |
| 更改 score 2–10、和/期望容差、百分位舍入范围 | `TestJevRegression`、`TestClassifierLiveScoreRounding`、mixed/validation fixtures；真实 wire replay 保留，越界拒绝 |
| 按入口旧题校验、错误接受旧名称/类型/图例/候选 | `TestJevRegressionFinalQuestions`、`TestClassifierMixedCallbacks`、新增 `TestTypeSafeNativeSnapshot`；最终改名/换 kind/重排有效，旧答案原子失败 |
| callback 持有 map/RawMessage 在 response 阶段改写最终依据 | mixed snapshots 与新增 native snapshot；HTTP bytes、criteria、答案均独立拥有 |
| 重试重新执行 Hook、重建请求或读新配置 | `TestClassifierHooksAndRetry`、unary snapshots 与新增 native retry；所有 attempts 固定最终 bytes、身份及凭据，Hook 各一次 |
| 接受 native bool、refusal/box、unknown/case alias/重复记录或嵌套重复键 | exact fields、strict envelope、closed JSON 与新增 `TestTypeSafeNativeGuard`；`callback_failed/request`、零 Attempt |
| 漏检 null criteria、kind 对应的 criteria 形状、空名/空题、说明非空及型号 kinds/数量/格式限制 | 既有 mixed callback 与新增 native guard；保留独立 native 校验，不回解为公共模型 |
| 允许 Hook 更改 model、认证、endpoint 或添加 authority | hooks/retry 与新增 native guard；`tenant_denied/request` 优先于内容损坏/容量，零 HTTP |
| 改变原错误优先级（scope→binding→capability→pin→callback），把 callback 子容量改为新错误 | `TestJevRegressionRejectionOrder`/AuthorizationOrder、strict envelope 与新增 native guard；精确 Code/Phase，resolver/credential/Attempt 数不变 |
| 改变完整编码题目计数（HTML escaping、bool 判别字、null 与省略） | 新增 native guard 的边界样例；沿用旧 canonical 编码字节预算与 callback_failed 分类，总请求超限仍 resource_limit/request |
| 单题失败发布部分答案、丢合法 usage/ResponseModel、重复 usage 被当作唯一 | Jev/strict envelope/mixed answers/unary usage；答案全空，合法 consumption/identity 按原优先级保留 |
| 2xx 失败重试、取消后发布结果、body/Admission 泄漏、Observer 泄露正文 | Jev/unary lifecycle/body limits/observation 与新增 native snapshot；2xx 单 Attempt，取消和异常释放，脱敏审计无发现 |
| 改动 Jev 1.13 wire/目录 Hash/冻结 Pi 或将本地 replay 当真实 PASS | builtin/catalog、显式 Pi differential、基线验证器；起点/终点逐文件 Hash 与目录 Hash 核对；仅声明受控 HTTP 验证 |

## 验证顺序与证据

1. 写上述失败清单与必要的公开 E2E 补充，生产实现保持不动。
2. 将 Jev/Classifier/Unary/目录及显式 Pi 差分跑绿，保存 baseline 的输入 Hash、请求/响应、结果/错误、Attempt/Observer、断言、运行配置、完整性清单和 audit。
3. 修改 native 编码/freeze/答案路径；先跑补充 E2E，再跑同一公开回归。
4. vet（普通/live）、相关 race、最后一次全仓测试；保存 final 独立证据及审查。
5. 两轴 code-review 以本票起点为固定点，修复发现后提交当前分支。

本票没有图片输入；图片 Hash 项明确不适用，文本 fixture 用 SHA-256 标识。
原固定 Jev 包、Pi pins、目录和 live/support matrix 均不重写。

## 实现前基线（2026-10-16）

`native-snapshot-evidence/replay.py --phase baseline --out .evidence/typesafe-native-snapshot/baseline`
已经退出 0：focused、显式 Jev Pi 差分、9 项 oracle 启动检查、历史固定包完整性及审计均通过。
生产源码未改动；输入 Hash 在输出 `inputs.json`，完整文件清单在 `files.json`。
补充 E2E 首次执行有一个测试目录配置错误（删除 bool kind 时仍需保留 score kind 与其等级声明）；
已在生产重构前修正测试配置，成功冻结 32 个补充 case；该失败不描述为功能红灯。
