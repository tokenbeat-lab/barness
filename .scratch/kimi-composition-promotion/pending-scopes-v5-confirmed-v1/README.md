# 七份待批准范围 v5：事实确认修订1

用户已在本轮直接确认：MiniMax国际、Kimi国内、Kimi国际三个账户组均与各自旧证据一致；地区、型号权限及额度正常；均为自有standard按量API；Kimi每站三协议同账户；冻结费率仍适用且无额外收费。确认依据见[事实原件](../../../.evidence/kimi-composition-promotion/account-confirmation-v1/facts.json)。事实仅按用户确认记录，没有声称新增厂商核验。

七份均更新为 accountAndRegion/ratesAndServiceTier/modelPermissionsAndQuota=USER_CONFIRMED；付费状态仍NOT_GRANTED／AWAITING_EXPLICIT_PAID_APPROVAL，approved=false、grantIssued=false、executable=false，授权ID和有效期仍NOT_ISSUED。

| 路线 | 仅补场景 | 调用/HTTP | 拟议USD |
| --- | --- | --- | --- |
| minimax-messages | thinking-on | 1/1 | 0.01 |
| kimi-cn-chat | xai-origin-history | 1/1 | 0.10 |
| kimi-intl-chat | xai-origin-history | 1/1 | 0.10 |
| kimi-cn-responses | xai-origin-history | 1/1 | 0.10 |
| kimi-intl-responses | xai-origin-history | 1/1 | 0.10 |
| kimi-cn-messages | xai-origin-history | 1/1 | 0.10 |
| kimi-intl-messages | xai-origin-history | 1/1 | 0.10 |

合计仍为7调用／7HTTP、拟议USD0.61，七个独立预算，全部零重试。MiniMax保留2048输入/输出、调用120s/进程180s；Kimi保留4096输入/输出、调用120s/进程240s、真实首发等待60s。原54＋13项、基包manifest、SDK、目录、价格与正常冻结二进制绑定均未改。本次不新增未知MiniMax独立缓存写价格，出现正写量仍拒绝资格；其他停止条件不变。

[原v5](../pending-scopes-v5/README.md)与所有旧索引、失败、证据、授权和账本在原路径保留。原v5 README标题误写v4也作为原件保留，本修订标题已正确标注。今后需另行明确批准才可逐路线签新grant；本轮仅记录事实与修订提案，没有签付费授权、真实claim、厂商调用、目录激活、live/race/压力或暂停恢复。
