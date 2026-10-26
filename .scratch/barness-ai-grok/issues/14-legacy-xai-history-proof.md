# 受影响旧目标的 xAI 来源历史与单次验证授权

Type: task
Status: claimed
Blocked by: 13

## Contract

[spec](../spec.md) R12/R13/R18/R24与E18；[SOP](../../../docs/agents/verification.md)；[22路线原生合同范围](../legacy-live-scope.md)。运行旧Provider自己的历史冒烟不能证明新的xAI来源触发条件。当前正式Record需要完整原生报告，不能以辅助单case增量包替代完整资格。

## Minimal implementation

每个已有默认/固定型号新增一个组合场景：同轮两个在旧规范化下碰撞的xAI opaque调用，以及后续轮复用同一完整pair，保留不同结果标记；只向目标Provider发送一次新增请求，不需要额外Grok请求或密钥。通过实际目标wire/Result/Observer证明唯一合法ID、正确配对与顺序、可见历史保留、原历史不变及外来密文/resource状态不泄漏。不得替换或删除原有能力。

共享数据构造/分协议检查有22个实际调用方，使用现有公共Client/Host/recorder；不导入私有allocator。源码或Go中的固定合成历史需进入每个原生来源库存，不能藏在未绑定的夹具里。

## Native contracts and budgets

- Generic6：在完整场景后添加默认型号的新场景，现有32call/4096输出边界保留；补当前缺少的本次单进程grant/claim/source/binary/catalog/account绑定。
- GLM2：每个固定glm型号添加1场景，22→24仍在24call上限；完整计划/期望/消费者同步，补本次grant/账户/地区绑定。需要显式代理时采用该原生路线专用字段，不能继承系统代理。
- Regional ZAI/MiniMaxCN：18→20与9→10，保持24cap；Kimi9→10并明确10call/10attempt/40960输出，131072输入须由实际有界夹具预留证明足够。
- International MiniMax3：13→14，保持20call/24attempt/40000输出/120000输入；采用全新单次授权/claim，不恢复旧累计账本、已关闭renewal或新建代际恢复控制器。旧原件只留历史事实。

授权只绑定当前各路线自身有限预算，不凭空声称没有的USD账单上限；缺用量保持未知。每次仅所选路线凭据，所有实际输入/构建/claim/原件保留。没有任何新厂商调用权限。

## Verification

公共E2E/CLI先RED后实现：22目标的实际wire碰撞/复用/降级检查；缺/错/过期/复用/混合路线grant在HTTP前拒绝；预算与Kimi新增10场景完整；国际MiniMax旧关闭authority不被恢复；完整/缺/篡改的新原生场景消费。受控/合成不是实际live资格。源文件不超500行，先核查已有依赖与模块，不新增通用controller或兼容shim。

## Answer

待实施；与12自身正式合同并行，最终一起审查/冻结后才提出具体付费范围。实际22路线仍NOT_RUN。

- 2026-10-26：维护者明确采用17条正式路线（Grok+16已发布旧目标），六条Kimi保留候选与受控兼容回归，不连带晋升。14所有22代码路径继续实现，六条Kimi真实live不在本次正式承诺内；[已采用范围](../formal-target-scope-plan.md)。未增加任何实际调用授权。
