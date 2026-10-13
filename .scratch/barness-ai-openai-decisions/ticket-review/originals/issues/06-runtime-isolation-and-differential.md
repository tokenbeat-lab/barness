# 06 — 运行时隔离、故障、Pi 差分与设计负载

Status: ready-for-agent
Type: task
Depends on: 05
Updated: 2026-10-12

## What to build

补齐最高 public Seam 的 lifecycle/mixed/race/压力证明，冻结 Pi runner 增加 Decisions classify/catalog，形成逐项差异账本和零 pending。依据 [R13–15/R19](../spec.md)、[设计 §3/8/9](../design.md)、[OD40–50](../acceptance.md)。

## Ownership and context

负责既有 runtime 上新增行为所需的克制改动、E2E controlled fault、public probe/资源证据、pioracle runner/ledger/projections 和 release traceability 新需求映射。保留唯一 Pi 1.1.0 pins 与旧 Jev/P07，不能回退他人变更或新建第二套 runtime。

## Failure scenarios first

- 504 的 Pi 差异促使全局 retry 被改、默认 2 retries、2xx 后重放、等待/错误 body 无界。
- 回调次数重复、租户/账户跨流、snapshot 随重试改变、许可/连接/body 泄漏。
- 流帧/事件队列限制误用于 unary；Observer 装入正文、图片或拒答统计。
- runner 用 chat Luna 或错 catalog type，外网/key 泄漏，丢全 body/result 来通过差分。
- 仅使用 synthetic MIME probe 证明图片，降低 pressure 负载掩盖故障，取消后无法再调用。

## Acceptance criteria

- [ ] 先写 OD40–50 与必要 red regression，后修改实际 runtime 接缝。
- [ ] shared explicit retry/header/backoff/snapshot 与 timeout/cancel/body limits 正确；2xx 后无重放。
- [ ] header/payload once、response once，permit/body/waiter 归零；图验证、hooks 与读取均可取消。
- [ ] Jev/Luna/chat/image 的 same-named Tenant/Binding/Model mixed 调用隔离，race 无发现。
- [ ] fixed oracle 新 APIS/OPENAI_CLASSIFIER_MODELS 分派可 classify/catalog；type/provenance/loopback guard 不放宽。
- [ ] 完整共同 wire/basic answer 比较；refusal/score/details/strictness/JSON/order/retry 仅有限 path/case 扩展，0 pending。
- [ ] 保留 Jev 原字段比较与官方 wire replay，新增 metadata 只做有来源的投影，不把旧差分当新路线。
- [ ] 新图片设计负载有预算依据、实际 headroom<=75% 和 bounded peak heap/吞吐/恢复证据，审计零发现。

## Verification and handoff

目标 faults/mixed → 全 Jev/Decisions E2E → 固定差分 → race/pressure。每次实际命令/source/manifest/audit 保留；出现自然失败先定位并保存，不重复直到“偶然绿”。向 07 交付离线协议可用结论，live/effect 仍未运行则 NOT_RUN。

## Comments

2026-10-12：仅为新增路由扩展当前 oracle，不升级已经冻结的 Pi 版本。
