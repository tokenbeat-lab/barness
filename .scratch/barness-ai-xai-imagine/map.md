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

- 04 resolved：双轴审查、修复、完整普通交付与只读发布事实见 [当前状态](spec.md#唯一当前状态) 和 [索引](evidence/record-20261029-01.json)。live/race/压力仍NOT_RUN，release.pass=false。

- 05 resolved：首次真实候选失败后仅修显式本地代理harness，受控29事件PASS、两轴0问题。原授权已消费，新真实范围待批准，见[新提案](retry-plan-20261029-02.md)。

- 新真实候选完成：用户明确选择代理127.0.0.1:7890并签发新一次性scope，6/6场景PASS、7图、原始ticks合计USD0.51。正式晋升/race/压力未批准，详见[候选索引](evidence/candidate-live-20261029-03.json)。

- 06 resolved：受限xAI图片条目进入Builtin，默认Client回放真实六场景；当前中文评估显式升级v2并保留v1。范围、普通证据及回滚见[工单06](issues/06-activate-catalog.md#answer)，正式资格/race/压力仍未完成。

- 07 resolved：原真实候选经新严格只读consumer取得正式资格，固定race与有限自身图片/Mixed压力完成；全部实际回执和只读发布结论见[唯一当前状态](spec.md#唯一当前状态)与[资格索引](evidence/qualification-20261029-06.json)。

## Fog

- 真实厂商形状与正式资格待该路线自身证据。
