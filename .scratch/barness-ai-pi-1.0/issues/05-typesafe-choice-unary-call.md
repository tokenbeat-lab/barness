# 05: 通过 Classify 完成 TypeSafe 单选调用

**What to build:** 宿主通过 Client 或 HookedClient 的 Classify 同步提交一组 TypeSafe 单选问题，得到严格校验的类型化答案、用量和调用元数据。以这条真实调用路径和现有聊天路径共同建立操作无关的调用运行时。

**Blocked by:** 04 — 按模型操作管理目录与 Binding 授权。

**Status:** resolved

**依据：** 规格“调用管线与 adapter”“公共入口与失败契约”“重试、准入与资源”“分类：TypeSafe System One”、实施设计第 7–10 节。

- [x] 先在现有场景 fixture runner 加入 classify 和整份 JSON 响应能力，写单选成功、缺答、多答、错型、非法分布、错误响应、操作或选项不匹配、未启用策略的失败 fixture；沿用受控 Provider 和宿主替身，经公共入口验证，不新增测试接缝。
- [x] Client 与 HookedClient 提供 Classify；nil 选项表示协议默认，首期 TypeSafeOptions 为空类型，拒绝不匹配协议的选项，不接受 SimpleOptions。
- [x] 绑定为 TypeSafe × typesafe-system-one × classifier，直连 HTTP 向 System One 提交一份状态和单选问题集合，使用 Bearer 凭据；一次逻辑调用不在库内拆分或并发凑答案。
- [x] 从聊天调用抽取操作无关的 scope、deadline、绑定、目录授权、凭据快照、尝试、准入和终态观测；聊天历史、SimpleOptions 转换、原生状态、assembler 和事件队列保留在聊天路径。按操作与协议固定分派，不开放动态注册。
- [x] 请求与选项在调用边界检查已知长度后独立复制；分类子策略显式启用且各容量为正，限制问题数、状态字节和单问题字节，构造时深复制。旧聊天策略仍能构造 Client，未启用分类时以 invalid_request/scope 返回带元数据的失败结果。
- [x] 单选问题键非空、问题集合非空、选项数 1–255；先完成字符串说明的最小调用路径。每个最终请求问题必须恰有一个同类型答案，答案选项属于选项集且为最高概率项之一，概率与置信度有限并在 0–1，分布键集完整且和在 1±1e-6 内。
- [x] 三种可信回调沿用现有契约并带操作字段；请求回调每逻辑调用执行一次，回调后的最终问题集合重新解码、校验和冻结，答案按该集合校验。回调不能改型号、凭据、目标或扩大授权；无效结果为 callback_failed，扩大授权为 tenant_denied。
- [x] 复用唯一 HTTP 客户端、传输时限、初始请求发送及绑定重试；默认不重试。成功响应后执行一次只看元数据的响应回调再读取，读取与解析完成前持有许可，任何失败关闭 body 并释放。
- [x] unary 成功响应按总输出字节读取，非 2xx 按错误体上限读取；超过帧上限但低于总输出上限的 JSON 能成功。读取、解码和答案校验失败使用 response 阶段，成功响应之后不重放。
- [x] 返回 stop、error 或 aborted 和对应的 *ai.Error；先登记用量再整体校验答案，失败时答案为空但用量和元数据保留。input/output 均存在为 complete，缺项为 partial，无 usage 为 unreported；TypeSafe 仅按输入费率估价，输出费率为零，价格来源待真实纳入时确认。
- [x] 实际版本化响应型号只记作 ResponseModel，不覆盖授权 ModelID；x-typesafe-request-id 记为厂商请求 ID。Observer 记录操作、调用与尝试元数据及用量，不含状态、问题、答案或错误正文，拥塞不改变结果。
- [x] 为 P07 登记已交付场景与追溯；新增 unary 协议 ADR，修订 ADR-0002/0005/0007/0008/0009/0010 与公共契约的已交付部分。聊天六路线回归保持通过，Gemini 聊天仍不调用响应回调；产出可重复且经审计的离线 E2E 证据。

## Comments

2026-10-08 完成：公共 Client/HookedClient.Classify 交付 TypeSafe 单选 unary 调用；先扩充既有
fixture runner 并记录失败，再实现独立类型、严格边界、可信回调与最终问题集校验。聊天与分类共用
scope、deadline、绑定/目录授权、凭据快照、尝试、准入和终态观测；输出限额、取消及异常退出均释放资源。
同步 ADR-0021、ADR-0002/0005/0007/0008/0009/0010、公共契约与 P07 追溯。Options 复用现有封闭接口，
保持错协议的 capability 拒绝；取舍记录在 ADR-0021。

最终全量 race PASS：2896 场景 PASS，619 opt-in 场景 NOT_RUN；P07 的 79 场景全部 PASS，审计零发现。
此前启用 pi 差分和压力的全量 race 3510 PASS，其中差分 611 PASS、压力 11 PASS；报告明确区分修复前后
版本与每次真实运行状态。普通/live vet、目录快照逐字节核对通过。两轴审查均发现重复响应字段覆盖的
P2，先补四个公开入口 red 场景再修复，独立有效用量保留、歧义用量拒绝；两名审查者复核后无剩余问题。
真实 TypeSafe API、内置型号与价格由工单 08 验收。

可重复 E2E、先红后绿、审查处理与脱敏审计见 [验证记录](../typesafe-choice-evidence/README.md)。
