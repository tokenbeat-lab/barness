# Ticket14 — 映射 pin 与付费前置契约预检：无 Key 修复历史

代码锚点为 `c2604db6f693f5c7d3a82c656512004ec85d2249`。这是离线修复与诊断预检的留存，不是新付费轮或发布证书。原票13付费18项均已关闭，原消费545 calls /545 HTTP Attempts保持原样；该数值限本次 MiniMax-release renewal，其他早期 GLM 项目历史另计。票13原A2在8fbd测得的失败不回写为成功。

## 复验

在仓库根目录运行，`--common-root` 指向包含外部归档的真实根目录：

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-14/repair/verify.py --common-root /Users/cyber/RestoX/barness
python3 -B .scratch/barness-ai-minimax/evidence/ticket-14/repair/check-boundaries.py --common-root /Users/cyber/RestoX/barness
```

仅校验有限大小的本地元数据和归档流；不执行留存的控制器、Go命令、历史请求或链接，不读取Key，不访问网络。还依赖已提交的票13 `implementation/verify.py` 中固定SHA的安全文件读取原语。被复验脚本及SHA256SUMS自身以本提交为信任锚，不能用可修改的摘要文件替代代码审查。

## 保存内容与结果

- 原生Go契约RED在a3a9097运行，退出1：6 cases中4 PASS /2 FAIL，失败原因为 `approved evidence mappings changed`。仅替换 `approvedMappingHash` 一行后，相同公开契约ordinary/race均退出0、各6 PASS；包括GLM四类删除/缺失负向模式。
- 独立计算ad20469前后完整trace的Go规范化hash，精确7项 `cases` / `requiredCases` 仅改目录元数据case名称，394项总数、必需集合基数与其他字段不变。trace本身在票14不改。
- 公开routed CLI RED为旧四项预检遗漏契约命令并错误放行；同2项CLI测试修复后通过。全工具60项通过，双vet退出0。TDD测试阶段HEAD/工作区源码变化在各原始source快照明确记录；不能把当时HEAD标签当作当时未提交源码的全部证明。其实际selected文件hash与最终c260逐项相等。
- 第一轮实际check-only在c260失败：vet0、vet-live0、no-key-live1，主manifest为398 cases：324 PASS /12 FAIL /62 NOT_RUN。缺少隔离工作树中的6个已提交selector固定的历史ZIP。完整原始失败与日志保留。
- 从共同工作树复制且逐个核对6个ZIP的固定SHA后，另开check-only目录，五项均0：双vet、no-key-live、visual-proxy、公开release contracts。no-key-live为336 PASS /62 NOT_RUN，visual16 PASS，release contracts6 PASS。62 NOT_RUN保持原样，不代表付费验收。两轮合计四份P08/P09单case子进程manifest没有独立audit文件，摘要明确记为null；主bundle audit均0。
- 两次实际预检均只产生PENDING模板。成功的五项结果没有生成APPROVED授权、新round或paid intent；实际控制器在前后核对原18 canonical receipt hash未变。这里只保留该已执行检查的原始证明，后续来源变化仍令旧完整1047 freeze自然过期。

原始归档位于共同根目录 `.evidence/minimax-retained/ticket-14-release-mapping-pin/ticket14-repair-raw.tar.gz`，SHA `88eae30a5c322edab67bde19a709c62acd7971bf07446a62637454d7332c407c`。归档为54,792个**常规条目**，原始payload425,402,095字节、压缩81,125,177字节。包含TDD原生bundle/命令/source、两轮check-only原始输出、16项visual默认输出、6 ZIP、selector、前后源码与控制器。

两轮测试各自刻意制造了一条 `formal-corrupt-symlink/formal-integrity.json` 负向fixture。留存过程不跟随或重新创建链接；`negative-fixture-links.json` 用常规JSON记录原相对路径、链接类型范围、原始目标文本/字节数/SHA。这两条链接是元数据投影，不能称为常规文件字节副本。归档读取器拒绝所有tar链接。其余常规原文件在快照前后逐个SHA相同。

两次prepare各有3个未被check执行的付费启动二进制，未重复保存其大blob；原始binary manifest、build命令和6项精确大小/SHA外部引用仍保留，校验器核对二者。该排除不用于缩减实际测试输入。

## 有限读取与反例

这是票14专用协议，旧票13读取器上限保持原样。固定上限为192 MiB压缩、1 GiB解码tar、90,000文件/100,000条目、64 MiB每成员、24 MiB解码inventory、8 MiB压缩inventory、8 MiB每观察元数据、16 MiB观察总量、900 MiB常规payload。实际解码tar为524,359,680字节。读取从安全目录fd逐层拒绝符号链接、非普通文件、重复/外来/越界路径与JSON重复字段；不解压到文件系统。

`check-boundaries.py` 从公共验证CLI回放14个拒绝场景：损坏hash、超限、链接归档/父目录、错误归档路径、扩大上限、洗白代码RED/首次预检失败、伪造付费批准、改变链接目标投影、tar链接、改变未使用二进制引用、改变安全helper、改变代码身份。原始归档在前后均重新验证，反例只写临时目录。

后续付费需要新固定source/binaries/精确审查/请求授权及独立合同。这里没有新round工具，也没有替换原18项、刷新旧来源pin或减少场景。当前代码修复导致的适用性边界见[票14](../../../issues/14-release-evidence-mapping-pin.md)；[GLM实际失败](../../ticket-13/actual-glm/README.md)、[旧15项实际结果](../../ticket-13/actual-legacy/README.md)和[中断/observer历史](../../ticket-13/keyless-launch-history/README.md)分别保留。
