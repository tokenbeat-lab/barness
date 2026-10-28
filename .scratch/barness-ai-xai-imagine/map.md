# xAI Imagine 实施工单图

## Notes

集成分支：`codex/xai-imagine-integration`。基线：`1e244ee26f024ce3534eed0600c43f6317d77adb`。

01 与 02 并行，03 依赖 01/02，04 依赖全部实施。各工单见 issues/。

## Decisions-so-far

- 公共 Client/HookedClient、NewClient、Catalog 为规格已确认的测试边界。
- 普通 E2E 授权成立。付费、race、压力与发布没有自动授权。

## Fog

- 真实厂商形状与正式资格待该路线自身证据。
