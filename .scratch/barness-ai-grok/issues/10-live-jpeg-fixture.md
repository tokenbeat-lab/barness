# 真实 JPEG 夹具满足已观测像素要求

Type: task
Status: claimed
Blocked by: 09

## Contract

[spec](../spec.md)；[SOP](../../../docs/agents/verification.md)。保留用户及工具结果 JPEG 场景，修正 live 合成夹具，不以删除图片能力或增加未经证实的公共约束规避失败。

## Ownership

ai/live/xai_scenarios_test.go 的 JPEG 夹具及公共进程 E2E 捕获回归。

## Verification

先以真实 controlled candidate 请求捕获验证图片可解码且总像素至少512，保留 RED，再修正夹具，取得 GREEN、独立两轴复核及冻结完整普通验收。重新编译绑定来源的候选二进制；不自动发起新付费运行。

## Answer

`00e7612c` 已修正JPEG为32×32；公共进程RED/GREEN原件保留，独立[两轴v4复核](../review/review-v4.md)0/0。最终完整普通冻结待采集。当前事实见 [spec](../spec.md#唯一当前状态)。

## Comments

- 2026-10-26：代理真实运行15场景通过，user-jpeg返回 invalid_image：8×8=64 total pixels，低于上游报告的512像素下限。Calls20、HTTP20、零自动重试，其余2场景未运行；原件和消费结果见 [不可变汇总](../evidence/live-attempts-v1.json)。
