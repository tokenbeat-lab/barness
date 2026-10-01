# 15: 用量、成本与价格版本（E11）

**What to build:** 成本管理开发者获得与冻结 pi 一致的 Usage 与估算成本（缓存读写、1h 写入、reasoning、阶梯价格），同时在 Result 元数据中看到每次尝试的用量完整性与价格版本，未上报用量不会被当成免费（spec I10 前三条、User Stories 40–41）。

**Blocked by:** 01, 02

**Status:** ready-for-agent

- [ ] Usage 保留 input/output/cacheRead/cacheWrite/totalTokens/cost、可选 cacheWrite1h/reasoning、初始化零值与 adapter 更新规则；reasoning 为 output 子集，不重复加入总量
- [ ] 移植冻结 pi 的阶梯价格、缓存读写、1h 写入与 adapter 专用调整；模型目录与价格快照带版本与哈希
- [ ] Result/观测元数据区分未上报、部分上报、完整上报，不修改兼容消息的数字字段
- [ ] 每次尝试分别记录用量；不把全部尝试合计反写为最终消息 Usage；失败请求的零值不被解释为免费
- [ ] Responses 成功、失败、重试路径上 usage 缺失 / null / 零值三态用例
- [ ] 兼容字段接入 pi 差分无待处理差异；完整性与价格版本作为扩展单独断言
