# 09: full/simple 选项、reasoning 映射与字段 presence（E04 选项部分）

**What to build:** 高级开发者通过按 API 区分的完整选项使用 Responses 全部能力；普通开发者用 simple 入口的统一 reasoning 等级与预算；未设置、显式 null、零值在请求中得到与冻结 pi 一致的编码（spec I7 前五条、User Stories 21–23）。

**Blocked by:** 01, 02

**Status:** ready-for-agent

- [ ] full 入口校验选项与绑定协议匹配，不建立按厂商名索引的选项命名空间
- [ ] simple 支持 minimal/low/medium/high/xhigh/max、thinkingBudgets、toolChoice 与公共参数；`off` 不加入 simple 输入枚举
- [ ] 移植支持等级、clamp、thinkingLevelMap 禁用/重映射与未设置分支；默认预算 minimal=1024、low=2048、medium=8192、high=16384 及自定义规则；maxTokens、回答空间、输入估算、4096 安全余量与无 contextWindow 例外；估算不裁剪历史
- [ ] presence-aware 字段表达未设置 / null / 零值三态
- [ ] samplingParams 与调用参数逐键合并、调用值优先；OpenAI-compatible 路径保留最后覆盖已命名字段的语义（其他 API 的忽略行为在对应协议票验证）
- [ ] cacheRetention、metadata、toolChoice 完整保留并按 fixture 验证；TenantID 不自动进入厂商 metadata；缓存/亲和标识按租户与账户作用域派生
- [ ] 发送前复核模型与资源授权，采样参数不能越过认证/目标边界
- [ ] 全部 reasoning 等级、预算边界与 presence 用例接入 pi 差分，无待处理差异
