# 第 1 节修复后的普通验收

唯一当前状态见 [status.md](status.md)；本轮实现冻结 `3aafd2d41436d222449f5b5f0e8c0596ed1154e9`，精确命令、数字退出、时间与哈希见 [verification.json](verification.json)。七项意见及 RED 原件见 [review-fixes.md](review-fixes.md)。旧失败、旧普通结果和历史压力偏差继续保留。

- 普通 Go 回执 PASS / exit0，墙钟 374.97 秒，12,890 个实际通过节点；原生 12,086 PASS / 214 NOT_RUN、audit0。与上轮完整 E2E 的通过 case 集合完全相等；本轮增加 supportmatrix 普通回归，未移除旧普通案例。
- 41 组工具进程场景 exit0，执行前后来源集合与冻结普通回执完全一致；单独保存采集脚本、实际 argv、退出、起止、工具身份、stdout/stderr 及子进程摘要。Go 配置注入仅读 go env；默认容量入口检查必须全部 NOT_RUN。
- 完整来源 1,244 文件，外层历史输入 8,636 文件。后者仍由 git ls-files .scratch 枚举并排除本功能目录，经 runpy 将逐文件 --input 交给原采集器；不存在源/输入变化。
- 普通及 live-tag 全 ai vet、live-tag 仅编译均 exit0，未运行 live 二进制。只读发布判定只重读现有证据，不启动下一阶段；race / 正式压力 NOT_RUN，release.pass=false。

修复后，普通阶段固定 GOENV=off / 空 GOFLAGS 并使用独立 HOME；只读判定核对政策。管道读取不使用阻塞线程，同一时限覆盖子进程及继承的管道。普通 E2E 同时注册宿主账户 HOME 的脱敏，消费者日志与实际 stdout/stderr 和退出一致。live 原生 tag/timeout 字符串不改写；只读消费者完整复用 native qualification，回执绑定来源、全部原件、结果、实际命令和日志，拒绝自报 true。清理必须同时核对全部成功声明和 audit，未知/失败/授权内容阻止候选隔离；失败和未知原件作为传递引用根；状态原子写入。

消费者新增公开只读形式为 releasegate -live-bundle <仓库相对 bundle> -out <attempt 内的新目录>，不执行厂商或写支持矩阵；发布 -record/-out 消费其 live-consumer.json 引用。受控合成文档仅验证消费机制，不产生真实资格。输出目录、符号链接和等价 flag 顺序均有进程用例，旧 argv/exit/files 消费者自报格式已删除。

所有 cleanup 测试仅在临时假仓库执行，含文件大小限制导致的保存失败及公开 restore；没有扫描、隔离、恢复、删除真实证据，也没有执行第 2–8 节。race、付费 live 与正式压力没有运行；此前误跑压力的事实见执行日志，不能用本轮成功覆盖。
