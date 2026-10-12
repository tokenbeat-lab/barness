# 工单 04：失败清单与实现边界

起点：`17f2f070fa63678c310b75191dbb2ae012091468`，分支 main。工单 02 已 resolved。
沿用工单明确指定的公共 Classify / HookedClient.Classify → Result、Attempt、Observer 和
loopback Provider 接缝；不测试私有 parser。按这些已授权接缝应用 tdd。

编码前枚举失败方式：

- usage 缺失/null、空对象、每项缺失/null、details 缺失/null被冒称 complete；全零不能区别未上报。
- Reasoning 未报被置 0；显式 0被省略；免费 output 被丢弃；合法 total 被重算。
- usage/token/details 重复键、大小写别名、错误类型、负数、小数、指数、int64 溢出、NaN 被接受。
- input+output 或 cached+cache_write 溢出；缓存合计超 input；reasoning 超 output；完整 total 矛盾。
- 已知部分项求和不检查；total 与未知项被擅自修补；合法 max-int64 被拒绝或丢精度。
- input 扣缓存/重复加缓存；272000 提前升价；升价只计超额部分；output/cache 错用宿主非零费率。
- 同名 chat 价格、ResponseModel alias、目录引用迟变污染授权身份或 Price Snapshot。
- 坏答案丢唯一合法 usage；坏 usage 输出已解析前缀；2xx 错误产生第二 Attempt；资源不释放。
- Observer 增加原始 cache 明细/答案；artifact 丢原始字段 presence、expected/actual 成本、终态或审计。
- Jev 旧用量/计费和原始回归漂移。

实现边界：独立 Decisions usage 转换，复用 Usage/Attempt/ModelCost，六项 presence 仅用于
转换和 Reporting。原始 presence 保存在测试 fixture/HTTP capture，不新增 Observer 字段。
保持 Host Catalog 授权与计价，正式型号与目录由工单 10 交付；本票只证明宿主合成输入费率。
不发超长真实请求，不估图片 token。全球标准输入估算不包含区域 premium；非零缓存和实际账单未核对。
