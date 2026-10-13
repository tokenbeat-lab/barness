# 执行前失败方式与实际保留结果

沿用票 11 的公开 CLI/process 接缝：混合/全拒答、错 route/Hash、缺 key、用量缺省、
24 次预算、取消不发后续请求、actual captures/Observer/verifier 篡改，执行前 105 PASS。
fixture、NOT_RUN 和 Jev 旧报告不能证明 Luna live 质量。

新增失败方式先由 `TestChineseEvaluationProxyProcess` 表达：外部地址、DNS 名、错误 scheme、
带凭据、缺/越界端口、路径/query/fragment、fixture 混入代理、CONNECT 故障。原 CLI 不识别
`-proxy`，所有 12 子场景及父测试 RED；实现后 13 PASS。CONNECT 502 的 24 次串行失败
无重试、无厂商发送、用量 unreported、最终 FAIL 且 artifact 可核验。没有关 TLS 验证、
改 endpoint、隐式使用环境代理或 overlay 已编译的源码。

本机无 key 的直连 `curl --noproxy '*' -I --max-time 10 https://api.openai.com` 以 code 28
超时，HTTP 000；没有调用 Decisions，也不是模型质量证据。最终独立 live 只有一次，
使用显式 `http://127.0.0.1:7890`，24 Call/Attempt、零重试、COMPLETE。

初次并行调度中，preflight 日志目录尚未创建，shell 重定向失败，测试未开始；创建目录后
独立运行的 preflight JSON 日志与完整 artifacts 保留。交付校验初次把正式 catalog 的
`{note, hash, catalog}` 封套直接与 catalog 比较导致本地断言失败，按 `catalog` 字段比较后
精确一致；未修改任何报告或再次调用厂商。

Standards 审查发现 Python launcher 在 Go 拒绝非法代理后仍将 URL 原文写入收据。
公开进程在无 key 的临时源码目录复现 userinfo/query/fragment 三种合成敏感标记持久化，
保留 `launcher-red.json`；首次 fixture 缺 matrix 的准备失败单独保留，不解释为泄漏复现。
Python 参数边界先校验后，`launcher-green.json` 三例无目录、无回显、无敏感值持久化。
本次 live 使用合法无凭据代理，不受影响；不修改实际包或追加厂商调用。
