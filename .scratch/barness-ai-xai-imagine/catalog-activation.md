# xAI图片内置目录激活范围

用户指令：2026-10-29“激活xAI目录”。本上下文激活grok-imagine-image-2.0图片目录；已有grok-4.7聊天数据保持。唯一当前状态见[spec](spec.md#唯一当前状态)。

## 目录边界

- image × xai × xai-images × grok-imagine-image-2.0，全球https://api.x.ai/v1。
- 文生图最多2张；编辑仍单输出，最多5张参考图。
- resolution为1k/1.5k/2k；quality为low/medium/auto；显式比例仅1:1/16:9。单图编辑仍拒绝显式比例。
- PerImage输入USD0.01/张，low×1k/1.5k/2k为0.04/0.05/0.06，medium为0.06/0.07/0.08，来源日期2026-10-29；auto仍按区间估算，不把本次实际收费反推为固定质量。
- 根目录content revision为2026-10-29.1。其他Provider条目和事实日期不改，历史全局版本由原Git快照保留。

## 自身依据与转换

原真实native candidate位于`.evidence/barness-ai-xai-imagine/live/xai-images/20261029T012001.157782Z-n96g7rg4`，6/6场景、7图、原始ticks合计5100000000、USD0.51；完整授权/claim/二进制/来源与摘要见不可变candidate-live-20261029-03.json。原测量stage=candidate、mode=live、formalQualified=false保持，直连失败与原授权也保持。

逐场景观察到N=1/2、参考图1/2/5、三个分辨率、三个质量值与两个显式比例；无需额外付费笛卡尔积。内置MaxOutput2不沿用候选设计上限10；比例列表不沿用候选16值。能力取已测边界与协议约束的交集，不声称视觉语义质量。候选JSON、版本与hash继续独立，以便重验原测量；生产目录不依赖testkit。

普通公共Client用默认Catalog重放六份留存原始请求/响应：最终wire、7图hash/MIME、reportedCost/ticks、token未报告、估算轴与价格快照一致。捕获只转为必要协议夹具，未复制整个胶囊。fixture/evidence原始来源与hash在`ai/e2e/testdata/xai-images/official.json`与`live-contract.json`中固定。

## 验证与发布边界

本次变化为目录metadata、根版本/快照与测试夹具；协议转换、网络、凭据、重试和资源生命周期不改。公共NewClient/Catalog/GenerateImages先红后绿；目录深复制、未授权扩展预拒绝、其他Provider字段不变及旧候选独立性均验证。完整受影响普通包最终冻结一次，相关定向回归不冒充全包。

新增厂商调用不适用：当前六份最终wire/结果以旧真实捕获回放验证，型号、原生payload与计价单位不扩张。原阶段事实不重写，本次不签发新付费授权。激活仅把已测受限型号加入BuiltinCatalog，不把源码可用、正式原生资格、race、压力或完整发布等同；后续人工阶段仍沿feature当前状态记录。回滚可禁用Host image Binding；源码回滚同时恢复该次Builtin内容与版本/快照。
