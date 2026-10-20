# 地区 Provider 实施图

## Notes

[规格与唯一当前状态](spec.md)。集成分支：`codex/barness-provider-sites`；起点：`68dec5a`。

## Decisions-so-far

- 01 协议、02 runtime、03 Host/目录可并行；04 和 06 依赖前三项并行；05 依赖全部实现。
- 测试接缝已由规格确认：公共 Client、Host 装配、既有 CLI/证据消费者。
- 新增付费与 race 未获本次授权；Key 更新仅表示秘密配置可用，不改变验证合同。

## Fog

新增九个身份自身 live、晋升与正式发布证明须另有授权和实际结果；保留开放，不降低合同。

- 实际冻结数据核对后拆出[06差分](issues/06-frozen-differential.md)，保留两条适用原生路线，另四条新增路线登记扩展；不升级冻结基线。
