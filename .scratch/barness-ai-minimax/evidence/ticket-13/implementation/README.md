# 票13实现与修复的无Key验证记录

本目录固定实现 `83ff7d6` 与审查修复 `51c3433` 的本地验证历史。包含两份原始归档的引用、逐文件清单、来源快照与脱敏日志视图；21MB原始包仍位于 Git common root 的 `.evidence/minimax-retained/`，这里不重复提交。全部调用均为无真实Key的离线 CLI / 合成账户 / loopback CONNECT 测试，付费调用为0。

最终记录：完整辅助工具检查58项 PASS；修复后5项 launcher E2E PASS；视觉代理 ordinary 与外层/子进程 race 各16项 PASS、audit0；视觉既有工作流/授权/正式效果定向回归39项 PASS；双 vet PASS。它们证明本次实现和局部修复的行为；不构成完整发布认证或厂商支持证明。当前真实验收和完整门禁由父任务另行记录。

保留历史失败及其含义：初始 CLI 缺失、缺少 run-selected、未登记新 renewal round 的 RED；最初视觉代理15项 FAIL，且断言复制 Go 帮助文本导致两条 HOME 脱敏审计 FAIL；空 `#` fragment 的额外 RED1；跨阶段报告被修改后仍报22/22的 `SPEC-13-01` RED。另有两个夹具错误：renewal输出名未满足已有 `minimax` 目录约束；第一轮审查修复测试恢复文件后未先删除就建立符号链接。它们均保留原始字节，未改成当前失败或成功。后续日志分别记录实现、断言和夹具修复后的通过。

视觉 ordinary 最终包的实际 `git_commit` 是 `2731aaa`，race 最终包为 `b327fc4`；这是提交实现之前的工作树 TDD。归档另存最终已提交源码快照83/51，不能把它们回写为当时测试的 Git 提交。Python检查使用临时合成 Git 仓库与固定公开子进程边界，不能当作厂商调用记录。

`index.json` 给出精确归档和清单 pins；`observations.json` 引用原始 manifest/audit 与日志。`logs/` 是注明转换规则的脱敏投影，并将 tab 展开为4空格；原始日志的 SHA 保存在索引中。早期审计失败只存在于不可变原始包；投影不替换或修补该原始证据。

独立程序核对逐字节清单、manifest计数和58/5项测试日志。双 vet 的退出0来自当时工具结果；只保留了空日志，没有另写独立数字命令收据，因此不能用空日志单独证明退出码。

在主工作树运行；其他工作树可增加 `--common-root <主工作树绝对路径>`。校验只读取这两份指定归档，不解包、不读取 `.env`、不调用厂商，也不要求历史 HEAD 仍为当前提交：

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-13/implementation/verify.py
```

独立校验器的失败边界是：本地文件或归档 hash 改变、任一符号链接/非普通文件、超限压缩/解压或成员大小、路径越界、重复/遗漏/额外成员及不符清单的内容均拒绝。上限固定在程序中：32MiB压缩、192MiB解压、32768成员、32层、64MiB单文件；元数据清单额外限制8MiB解压。先核对压缩文件SHA再解析，顺序读取并强制耗尽gzip至校验尾部，从不提取归档成员。

重复行为检查会创建新的本地临时证据包，不重写这里的历史记录：

```sh
python3 -B -m unittest discover -s .scratch/barness-ai-minimax/tools -p 'test_*.py' -v
go test ./ai/e2e -run '^TestVisualEvaluationProxyProcess$' -count=1
BARNESS_EXAMPLE_RACE=1 go test -race ./ai/e2e -run '^TestVisualEvaluationProxyProcess$' -count=1
```

重放结果属于运行时的源码和环境，须单独保存，不续借这份历史验证的来源身份。
