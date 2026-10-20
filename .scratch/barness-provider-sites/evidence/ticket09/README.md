# 工单 09：独立正式续期授权的离线实现

这是 2026-10-20 的实现与普通验证记录；当前验收状态只见[规格](../../spec.md)。实际 Key、厂商 HTTP 和 race 均未执行。本记录不构成付费批准或正式支持证明。

固定新轮次为 `provider-sites-final-acceptance-1`，只允许国际 MiniMax / standard / M3 的三协议。旧 ID 永久关闭活动执行，保留 schema 1 的历史只读检查；新 schema 2 同时绑定原正式 39/36 和已关闭首轮续期 39/36 的完整 15 份 canonical 事实。只读 launcher preflight 与子进程 `BeginRenewal` 分开，后者在读取所选 Key 前领取持久单进程 claim。随机进程 nonce 阻止复制相同 session 续跑；pending claim 在崩溃、失败或删除账本后保持关闭。[ADR-0050](../../../../docs/adr/0050-final-acceptance-renewal-authority.md)记录边界和取舍。

先在独立 Git owner 的公共进程接缝取得新 ID 拒绝 RED，再完成三身份 39/36 GREEN。随后保留了同 session 换进程、缺少 pre-secret begin、美元 guard / 区域价格依赖漂移、旧 schema 偷带新字段、原 formal 父目录 symlink、非 standard 身份 claim 的实际 RED→GREEN。并发竞争、linked worktree、目录集合、审批 / 来源 / cap、未注册场景、残写、删除账本及重复执行也按公共进程验证；所有合成 authority 都只属于临时独立 Git owner。

合并集成分支 `1413408` 后，authority 普通包 66 PASS / 4 私有 child SKIP，MiniMax controlled / replay / research / 原件消费者进程 70 PASS，supportmatrix / release 相关消费者 60 PASS。最后两项边界修复后，新轮次进程集合 45 PASS；并入仅文档的 `3e7f9ad` 后，4 个子场景及 2 个父测试再次 PASS，普通 live vet PASS。每项都保留实际 argv、数字退出、起止时间及源码 SHA；没有把并行包时间相加。一次后并入复核选错 `-run`，实际为 0 tests，明确排除，不计验证；更正的 6 PASS 原件另存。

三份本轮原生普通 bundle 连同子 capture 原样留存，18 份直接 `audit.json` 及 45 份子原件归档内审计均 0 发现；原目录与保留副本逐文件 SHA 相等。原 15 份 canonical 前序在三次实际只读 `-renewal-source` 之间均保持相同摘要 `0aa629db8674437316645295f4000237c958c85d81df9c337cc2a4be5935448e`。未复制、重封、刷新或改写这些历史证明。

[全部本轮命令、原件与索引](/Users/cyber/RestoX/barness/.evidence/provider-sites/final-authority-20261020/index.json)、[文件 SHA 清单](/Users/cyber/RestoX/barness/.evidence/provider-sites/final-authority-20261020/files.json)。完整来源审查、实际人类预算、资格采集和正式发布判定仍由[工单 07](../../issues/07-final-acceptance.md)承接。

建议人工 race 范围是原生授权进程的 claim / 取消 / 并发读取与 MiniMax controlled 进程回归，或最终完整发布 race 的相应覆盖；本轮 `race: NOT_RUN`。不得将该建议当作自动执行许可。
