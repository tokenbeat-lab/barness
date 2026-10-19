# 票19：HOME 下续期拒绝诊断的审计修复

源码修复 `a25737a` 仅改 `ai/live/minimax_renewal_process_test.go`。实际执行 argv 保持原样；本用例记录的 argv/stdout/断言诊断局部替换可执行文件与 run.Dir 为 `$MINIMAX_RENEWAL_TEST_BINARY`、`$MINIMAX_RENEWAL_ROOT`，保留 selector、timeout 和证据路径后缀。Stat 诊断复用既有 errText。未改 Provider、授权、正式消费者、审计规则或全局 redactor，所有原断言保留并新增路径上下文/精确参数断言。

原 `0bbd73d` 六项前置检查为5×0/1×1，保留于父层 [preflight-first](../preflight-first/README.md)。本次用相同公开测试、真实 HOME 下的新证据目录复现：用例自身 PASS，但 unchanged audit 在4件文件找到9处 HOME，Go exit1。再先加入别名/参数断言，实际 RED 为6条断言失败并保留原9处审计发现。归档内原 RED 中的 HOME 文本是明确保留的历史失败，未重写为脱敏成功。

完成局部修复后，相同 HOME-contained 公开 ordinary/race 均 exit0、case PASS、audit0；三种 controlled 拒绝仍全部失败退出、0逻辑预留/0HTTP、completefalse、无正式发布。双 vet0。共12份原始受控子报告都保留。未读取 Key、未付费，也未建立真实 grant。

测量 HEAD 是 `0bbd73d` 加本文件 WIP；三个阶段源码原件和六份数字命令收据分别固定，source-after 字节等于提交 a257。不是事后冒充已提交源码的重新测试。此前 fbb/native 的完整检查不改写成本次源码的整套复验；父层会在新冻结目录重新完成全部前置检查。

`index.json` 和完整 inventory 固定101普通成员、440596原始成员字节。原包位于 common-root `.evidence/minimax-retained/ticket-19-native-home-redaction/raw.tar.gz`：185032压缩字节、SHA256 `c61b768bf5b0b36dd2b10415b24fc6ab5a34d00d47a9df2013c064704de453b4`。复用票13既有32MiB压缩/192MiB解压、32768项、64MiB单件、1MiB观察元数据、8MiB清单 reader；不执行捕获命令、不解包。

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-19/native-home-redaction/verify.py --common-root /Users/cyber/RestoX/barness

# 在新的、位于当前 HOME 下的证据目录重放；不覆盖原件。
BARNESS_AI_EVIDENCE_DIR="$HOME/new-owned-evidence" go test -tags live ./ai/live -run '^TestMiniMaxFormalRenewalControlledRefusal$' -count=1 -timeout=3m
```

该记录只证明离线诊断和证据审计修复，不授予当前真实厂商资格或完整发布 PASS。
