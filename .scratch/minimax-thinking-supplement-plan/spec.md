# MiniMax 国际 Messages thinking-on 最小补采设计

Status: claimed
Type: task
Date: 2026-10-28

范围：基于已完成的13项离线适用性与断流受控复现，交付固定单场景采集、13+1只读组合消费的最小设计，明确实现位置、独立授权、预算、验证与停止条件。只写本目录和同名.evidence中的新设计/核验记录，不修改采集器/消费者/产品/ADR/SOP/旧状态，不签发可执行grant，不读凭据、不运行厂商live/race/压力、不恢复暂停或修改账本。

保留原失败包全部15条Call/14个唯一场景、原producer exit1、旧授权和ledger；只复用13个成功场景及其恢复/取消/预检的既有语义。新thinking-on须单独真实成功，设计通过不产生新资格。提案数据不得作为当前可执行Scope/Grant。
