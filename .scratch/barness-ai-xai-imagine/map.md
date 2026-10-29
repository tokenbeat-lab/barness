# xAI Imagine 实施工单图

## Notes

集成分支：`codex/xai-imagine-integration`。基线：`1e244ee26f024ce3534eed0600c43f6317d77adb`。

01 与 02 并行，03 依赖 01/02，04 依赖全部实施。各工单见 issues/。

## Decisions-so-far

- 公共 Client/HookedClient、NewClient、Catalog 为规格已确认的测试边界。
- 普通 E2E 授权成立。付费、race、压力与发布没有自动授权。

- 01 resolved：按张领域、合法目录、深复制、Reported Cost 与候选目录见 [工单01](issues/01-image-pricing.md#answer)，集成提交 `277520b4`。

- 工单02：同步图片协议、用量/收费与自身资源守卫完成；普通回归证据见 [工单02](issues/02-xai-images.md)。

- 03 resolved：固定候选场景、有界原生捕获、一次性授权与受控过程验证见 [工单03](issues/03-live-readiness.md#answer)，集成提交 `46add731`。

## Fog

- 真实厂商形状与正式资格待该路线自身证据。
