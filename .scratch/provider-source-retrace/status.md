# 六条来源追溯当前状态

2026-10-27：离线追溯与新的适用性判定完成，**6/6 APPLICABLE，原件缺口 0**。未启动厂商请求、race、压力或暂停轮次；本次结论不是新 live 测量，也不授予新执行权限，**release.pass=false**。

本功能唯一当前交付事实：[completion-v2.json](../../.evidence/provider-source-retrace/completion-v2.json)。授权范围及先行失败清单见 [spec](spec.md)。

## 六条路线

下表为原测量场景数量；本次只读核验，不增加测量次数、不更新原测量时间或 stage/mode。

| 路线 | 原测量场景 | 新适用性 | 缺原件 |
| --- | --- | --- | --- |
| OpenAI Responses | 9 PASS | APPLICABLE | 0 |
| DeepSeek Responses | 9 PASS | APPLICABLE | 0 |
| OpenAI Chat | 10 PASS + 1 原固定 UNSUPPORTED | APPLICABLE | 0 |
| DeepSeek Chat | 12 PASS | APPLICABLE | 0 |
| Anthropic Messages | 12 PASS | APPLICABLE | 0 |
| Google Gemini | 7 PASS | APPLICABLE | 0 |

六份新的 schema2 只读消费回执与逐条哈希见 [current-applicability-v2/report.json](../../.evidence/provider-source-retrace/current-applicability-v2/report.json)。合计沿用原 59 PASS + 1 UNSUPPORTED，未把不支持场景改成通过。

与前次的 10 条可用原测量合并后，当前索引为 **16/17 可沿用**。MiniMax 国际 Messages 的原真实测量仍是 incomplete/FAIL，未补执行、不恢复暂停。前 10 条引用前次判定，本次只更新这 6 条；产品实现没有变化。见 [current-qualification-index-v1.json](../../.evidence/provider-source-retrace/current-qualification-index-v1.json)。

## 来源链与原件缺口

原执行 scope 固定 plan，plan 固定冻结构建回执；构建回执保留 regional 原来源映射，其规范 JSON SHA256 与六份原 grant、原生 manifest 的来源声明相同：`3db5898952cdfe27176820b0535b8be659d7416cbafdc6026dd87b0a14b2d83b`。

538 项来源字节全部找到：455 项与最终当前文件逐字节相同，83 项可由索引中指定的固定 Git blob 取回；缺失 0。逐文件位置及原回执的 source/selectedInputs 对照见 [source-trace-v2.json](../../.evidence/provider-source-retrace/source-trace-v2.json)。原映射单独保存为 [execution-source.json](../../.evidence/provider-source-retrace/recovered/execution-source.json)，不补写原 bundle。初次定位为 456 项当前文件 + 82 项 Git；实现改变了其中一个只读核验文件，最终索引已将其原字节位置改指固定 Git，原映射和旧追溯回执均保留。

原 collector 的 executable SHA 是 `/usr/bin/env`；已继续核对实际 argv[3] 原生二进制，与冻结构建、原 grant 的 SHA256 一致。未将 wrapper 哈希误认作测试二进制，未执行旧二进制。见 [executed-binary-link-check-v1.json](../../.evidence/provider-source-retrace/executed-binary-link-check-v1.json)。

当前只读入口显式接受 `--execution-source`；来源补充的文件哈希与参数进入回执，最终 release 消费也复核。仍核验原生身份、原始 bundle、场景、目录、wire/usage、grant/claim、预算与时间；若原内部映射存在，则必须独立一致。只读补充不进入授权或执行入口。

## 验证

基线/受测 HEAD：`69ff4af880e68ef3ca245e5128bdae5e59a2ef7b`。变更仍未提交，不能将 HEAD 视为本次实现；受测身份为最终回执的 1999 项 Source.Files，集合摘要：`caccb0ed44d02bc7529deb225581cfc8079f73cfbe4978134f6898bc15aeacba`。六条新只读回执来源与最终普通回执完全一致。

| 范围 | 结果 | 证据 |
| --- | --- | --- |
| 完整受影响普通包 | PASS，exit0，589.324 秒；179 项实际输入和执行来源前后一致 | [最终普通回执](../../.evidence/provider-source-retrace/e2e/20261027T083912.573336Z-f2go7fq2/receipt.json) |
| 主 E2E 原生包 | 15152 PASS；214 原有 opt-in/手工场景 NOT_RUN | 同上唯一 `./ai/e2e` bundle |
| 辅助离线包 | GLM 32、MiniMax 原生工具 77、探针 6、visualeval 25 全 PASS，audit 无 findings | 同上各辅助 bundle；live 编译标签不代表发起厂商请求 |
| 工具进程套件 | 56/56 PASS | [日志](../../.evidence/provider-source-retrace/tool-process-suite-02/unittest.log) |
| 静态检查 | 受影响包及当前 releasegate 的 go vet exit0 | [回执](../../.evidence/provider-source-retrace/static-validation-v2/receipt.json) |
| 来源补充及跨阶段消费 | 11 个场景通过，含缺失、篡改、符号链接、冲突、产品漂移、消费后变化与原件不变 | 主 E2E 的 `TestGenericHistoricalSourceSupplement` |

早期 FAIL、预期 RED、CANCELLED 与最终 PASS 全部保留，见 [attempt 索引](../../.evidence/provider-source-retrace/attempts-index-v1.json)及[命令日志](../../.evidence/provider-source-retrace/command-logs/)。本次未改变产品 SDK、adapter 或 catalog，没有牵动所有 Provider 重新做 live。

## 历史保留与测试隔离异常

本次 1534 项受保护原文件（包括六份旧 bundle、原回执、原二进制、授权、claim、账本、支持矩阵）前后哈希一致，未删除、重写原文件或恢复旧权限。见 [保护清单](../../.evidence/provider-source-retrace/protected-history-before.json)与[最终核对](../../.evidence/provider-source-retrace/history-protection-check-v2.json)。

第一次完整普通回归暴露两个既有 CLI 合成授权测试遗漏私有 Git 命名空间，向公共授权存储新增 25 份合成记录（包括故意的空写入记录）；无厂商调用、无真实授权消费，原有记录未改。已停止该轮，**25 份新增记录全部保留并索引**，未删除或迁移。随后在两处测试复用现有 privateGrantStore，并在领取前断言实际 Git 存储属于私有目录；先记录失败，再修复重跑。见 [异常索引](../../.evidence/provider-source-retrace/synthetic-store-exposure-v1.json)。

修复后的最终全量与工具验收前后，公共存储 1021 项的集合及每项字节完全一致；新增 25 项仍可定位。原始保护失败回执与中断轮次仍保留，不以最终 PASS 覆盖。

前次精简状态的更新前原文已保存为 [simplification-status-before-source-retrace.md](../../.evidence/provider-source-retrace/simplification-status-before-source-retrace.md)。用户已有修改未覆盖。付费 live/race/压力均 NOT_RUN，暂停轮次 NOT_RESUMED；无新的完整发布通过或自动重测义务。
