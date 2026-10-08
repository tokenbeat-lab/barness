# 固定中文分类评估宿主

本例通过公共 `Client.Classify` 将固定合成中文支持任务提交给 jev-1.13.0，
产出独立业务报告。标签和真值属于任务；没有修改 barness-ai 分类协议或支持矩阵。

仓库根目录执行：

    go test ./ai/e2e -run '^TestChineseEvaluation' -count=1
    go run ./ai/examples/chineseeval/cmd/chineseeval

第二条默认不联网，写新的 NOT_RUN 包，24 条样本均保留，分数为 null。
`-out` 指定证据基础目录，每次仍创建新目录。`-dataset` / `-config` 指定输入文件，
上限 1 MiB / 16 KiB；拒绝未知或重复字段、缺标注、分布不符、版本/哈希漂移及无限预算。
本例只接收 synthetic 数据，当前四类明确目标的任务之外要先制定新的标注任务。

显式真实评估需命令 `-live`、`BARNESS_AI_CHINESE_EVAL=1`、
`BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY` 和 `BARNESS_AI_CHINESE_EVAL_ACCOUNT_ALIAS`。
进程拒绝其他凭据环境变量。根 `.env` 的 TypeSafe 凭据可通过隔离启动器使用：

    python3 .scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/run-live.py --live

启动器只提取 TYPESAFE_KEY，构建并运行独立宿主进程。账户别名为
local-typesafe@region-unreported；权限级别与区域未独立核实。宿主固定官方目标，
最多 24 次调用、24 个问题和 24 次尝试，零重试，单次 15 秒/总计 180 秒；
SIGINT/SIGTERM 会取消当前调用并保留剩余 cancelled 状态和已知用量。
不完整用量标为 usage_incomplete，可用答案供诊断保存，该样本不记为完整可评分结果。

`evaluation-report.json` 包含数据集/配置内容和原字节哈希、实际响应型号、全部样本状态、
尝试元数据、预算/用量、混淆分布与错误样本 ID。ID 可定位同报告中的文本、真值、标注依据
和答案。dataset.json/config.json 保留原字节；manifest.json 记录代码提交、运行环境和版本；
audit.json 为既有逐文件脱敏审计。observations-evaluation.json 只含原有 Observer 元数据。
API 失败只保存错误分类，不保存可能包含敏感数据的厂商错误正文。

文件回放不联网、不改报告：

    go run ./ai/examples/chineseeval/cmd/chineseeval -verify <bundle-directory>

核验原输入哈希、所有样本、版本、指标重算、预算/用量、审计结论与包内敏感字段。
COMPLETE 表示完整流程已执行，未规定准确率门槛；FAIL 为不完整执行，仍完整保留分母；
NOT_RUN 是未执行。fixture 只证明流程，不能代表实际型号效果。

解释指标：accuracy = correct / 全部 24 个样本；scoredAccuracy = correct / scored。
FAILED/未执行数量分别公开，混淆表按真值行、预测标签或失败状态列统计。
confidenceBins 与 topProbabilityBins 下标 i 表示 [i/10,(i+1)/10)，最后一段含 1；
均值和正确率只针对有答案且完整上报用量的 scored 样本。
ECE 为各段 |正确率−均值| 按 scored 数量加权；Brier 为四类概率与 one-hot 真值的平方差之和，
在 scored 样本上取平均，没有再除以类别数。confidence 是厂商独立信号，不能据此声称
已校准为正确概率。本集只有明确的合成目标，样本很少，不能推断生产分布效果。

变更数据集、标注规则、问题或预算时升版本并更新数据集及配置哈希；目录变化后显式更新
价格 pin。保存新包，不覆盖历史；真实答案可以变化，每次都保留日期与型号。
