# 票13 — 8fbd 的实际完整门禁：FAIL

实际源码提交为 `8fbd17006677774368d4d77d924a8672fd205c51`。公开releasegate由同一源码直接编译后运行，2026-10-18 08:42:46Z开始，09:34:10Z完成。四项**真实数字退出码**为vet0、vet-live0、ordinary1、race1；外层退出1，正式报告存在且 `pass:false`。没有通过证书，也未自动重跑。

两次完整suite各12,448 cases：12,439 PASS /8 FAIL /1 NOT_RUN，audit0。各868项真实 `PIDIFF-` case和29项pressure case全部PASS，目录metadata case独立保留并不计入差分。两次各98,703个原生manifest artifact引用都与归档字节的SHA匹配。

八个失败是三路GLM的formal-import/known-usage六项，以及 `MM10-release-contract`、`P0-GLM13-release-trace-complete` 两项映射pin失败。`OD57-release-recorded-bundle-diagnostic` 为未配置的可选历史诊断，却先注册而留下NOT_RUN；这条原始记录不删除。另有**未注册进E2E manifest的release包测试失败**：Decisions valid fixture缺artifact inventory，以及两个release映射单元测试；公开报告保留完整命令tail，不把12,448个case当作全部Go测试的通过证明。后续离线修复单独留存，不能重写此次实际失败。

八项输入公开验证命令均退出0；公共门禁仍如实得到17身份16 PASS /1 FAIL，唯一失败为GLM Responses实际3项能力失败。其余组合不会因为aggregate LIVE/trace失败而被宣称独立失败。394项trace为354 PASS /40 FAIL；23项公共audit无findings。最终报告、原始根目录输入pin、源码、编译/观察器与二轴审查收据都在小型verdict归档。

## 只读复验

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-13/final-gate/verify.py --common-root /Users/cyber/RestoX/barness
python3 -B .scratch/barness-ai-minimax/evidence/ticket-13/final-gate/check-boundaries.py --common-root /Users/cyber/RestoX/barness
```

仅流式读取固定原始归档；不解包到磁盘，不执行捕获的命令，不读取Key，不访问网络。当前HEAD不要求等于历史8fbd，历史native来源保持原样，校验器不调用旧付费guard伪装当前适用性。安全读取原语固定复用票13 `implementation/verify.py` 的SHA；本提交是这些程序与SHA256SUMS的信任锚。

## 原始字节与边界

共同根目录 `.evidence/minimax-retained/ticket-13-final-gate/` 中三份独立只读归档：

| 归档 | 常规文件 | 压缩字节 | SHA256 |
|---|---:|---:|---|
| ordinary-raw.tar.gz |133171|114322540|`9f72789f54f61f40cba8880c9a42f7be818df02f431c1360e250dbafd1329467`|
| race-raw.tar.gz |133169|113453106|`b5b96d819cd097c6db13ebbc2101f1a3988ecd05ad4f607a1f32d7b561a0371f`|
| verdict-raw.tar.gz |45|407502|`5988535e18f3cefcf793cd6ea126b16ecd06b062e20c45256eb434d69910078e`|

ordinary在结束后先冻结，race结束后另冻；最终外层关闭后才保存verdict。三个原始集合快照前后逐文件SHA相同；先前ordinary/race归档没有因最终报告或离线修复重写。30MB编译二进制已在[launch history](../keyless-launch-history/README.md)归档，当前仅固定同一SHA/大小与原始编译收据，避免重复blob；它不作为修复后当前可执行程序复用。

每个完整suite采用独立固定上限：256 MiB压缩、2 GiB解码tar、160,000文件/200,000条目、64 MiB单成员、40 MiB解码inventory/8 MiB压缩inventory、16 MiB单观察元数据/64 MiB观察总量、1600 MiB原始payload。实际ordinary/race解码tar分别1,538,959,360 /1,516,769,280字节。小型verdict另限原始32 MiB/256文件，保存45文件4,619,704 payload字节。旧implementation的32MiB/192MiB/32k协议从未扩大；`capsule.py`是明确的新全suite协议。

验证逐级安全打开固定相对路径，拒绝链接/非普通文件/越界或重复成员/JSON重复字段/改动的hash、source或compiled observer身份。14项公共CLI反例覆盖archive损坏/超限/链接、父链接、洗白FAIL/真实race1/额外unit失败、错路径/扩大上限/tar链接、错helper/source/binary和隐藏suite失败数。所有反例只写临时目录；前后复验原始三份归档。

此处新增付费调用为0。原本次MiniMax-release renewal已知545/545及原39收据保持原样，早期独立GLM及MiniMax自身候选/正式账本另计。任何未来验证范围与新授权都由新的已审查合同决定，本历史留存不授予调用或新的来源适用性。
