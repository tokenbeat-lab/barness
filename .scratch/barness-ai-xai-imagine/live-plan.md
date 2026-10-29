# xAI Images 候选真实验证提案

本文件是待人工选择的具体执行范围，不是付费授权。唯一当前状态见 [spec](spec.md#唯一当前状态)。目录身份固定为 image × xai × xai-images × grok-imagine-image-2.0，endpoint=https://api.x.ai/v1。源码、二进制 SHA、目录 hash 与实际 argv 在最终普通验证和审查后固定。

## 固定场景

| 场景 | 参考图 | 输出图 | quality | resolution | aspect_ratio | 目录费用上界 USD |
| --- | ---: | ---: | --- | --- | --- | ---: |
| generation-low | 0 | 1 | low | 1k | 1:1 | 0.04 |
| generation-medium | 0 | 2 | medium | 1.5k | 16:9 | 0.14 |
| generation-auto | 0 | 1 | auto | 2k | 1:1 | 0.08 |
| single-edit | 1 | 1 | medium | 1k | 省略 | 0.07 |
| ordered-edit | 2 | 1 | low | 1.5k | 16:9 | 0.07 |
| five-reference-edit | 5 | 1 | auto | 2k | 1:1 | 0.13 |

合计6次调用、最多6次HTTP、8张输入、7张输出。默认与Binding重试均为0。场景固定图数/档位，计划本身进入执行来源。目录费用最坏合计USD0.53；建议一次性USD0.60停止阈值。上界按2026-10-29目录计算，不代表厂商扣费承诺。每次调用前预留完整上界；失败不退还预留。厂商未报告收费、超预算或任一必交失败时停止后续调用。

每调用时限180s，进程总时限1200s。新授权须绑定本路线、场景、来源、二进制、目录、MaxCalls=6、MaxAttempts=6、MaxUSD=0.60及有效期，并经现有livegrant一次性claim。不得复用聊天授权或旧claim。此前候选的存在不授予执行权。

## 保留与验收

使用本路线唯一凭据，凭据不进入argv/产物。保留完整有界脱敏请求/响应、原整数cost_in_usd_ticks、返回图片、结果与错误、Observer、预算消费、原授权/claim、命令/退出、来源、manifest与audit。输入使用本地合成PNG/JPEG/WebP，不下载外部图片。精确顺序由请求捕获的原字节核验；语义图像质量不冒充协议验证。

controlled回放仅证明harness过程，不能取得真实资格。六场景即使真实通过，也不证明10输出、全部16比例或任意默认/厂商硬限额。正式目录继续由自身证据实际覆盖范围另行晋升。没有选择与真实执行时live=NOT_RUN、release.pass=false。

## 后续人工阶段候选

race仅在本路线所需live完成后提交精确argv，覆盖图片准入、快照、取消和Observer并发。压力在race通过后提交精确设计负载，覆盖5参考图/多输出、请求响应字节边界、并发准入和取消。当前两项均NOT_RUN；本提案不发放阶段选择。
