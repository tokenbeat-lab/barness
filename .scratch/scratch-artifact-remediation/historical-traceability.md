# 历史链接与原样快照追溯

第4项通过外部索引处理原报告349处失效候选，保留所有历史正文与既有结论。272处可从已证实的原路径基址解析，2处临时地址已有原哈希一致的永久副本；70处已删除源码链接只能确认历史版本，不能声称已恢复其未固定版本的原字节。

当前进度与实际检查/只读回执见[唯一当前状态](status.md)。本页和索引仅用于阅读追溯，不供产品/证据消费者执行；不改旧pin/授权/余额，不续期支持或发布资格，recovery7保持[当前暂停](../barness-provider-sites/tools/README.md)。

## 候选处置

| 原因 | 数量 | 处理 |
| --- | ---: | --- |
| 第2项已修正的活动入口 | 1 | 第2项已撤下活动链接；旧文原字节仍可读。 |
| 已删除仓库路径/历史版本 | 70 | 28个当前不存在的路径，找到固定Git历史blob；原目标版本/哈希未记录，70处引用身份仍未确认。 |
| 原样快照相对基址 | 272 | 按Kimi spec、Decisions出版映射或MiniMax delivery-status的原基址解析，不改快照正文。 |
| 临时地址失效，原哈希找回 | 2 | 两份引用指向同一旧/tmp地址，两个随Git原件与原文SHA逐字节相等。 |
| 厂商网站根相对地址 | 4 | 按原sources.json的来源域名解析，不当本地路径；远端当前可用性未检查。 |

导航：[349项逐项处置](historical-link-dispositions.md)、[完整处置JSON](historical-links.json)、[原样快照索引](historical-snapshot-index.md)、[快照来源JSON](historical-snapshots.json)、[固定Git历史版本](historical-git-targets.md)。

## 已确认的临时原件永久指针

旧`/tmp/barness-ticket17-current-source-applicability.json`目前仍不存在；无需恢复该临时位置。两份引用的原文均记载SHA256：

`29d7009a2f31e03199b17225b3f069b78d99777e4c0d8c21ea71f1c9e51f8ff0`

- [ticket09永久原件](../barness-ai-minimax/evidence/ticket-09/final-closure/current-source-applicability.json)，49,705 bytes。
- [ticket17审查原件](../barness-ai-minimax/evidence/ticket-17/reviews/current-source-applicability.json)，49,705 bytes。

两份均与原记录SHA一致，限定来源230691742b731c77e8167b94646219e10f78c8f7；原严格相等性false、原矩阵/原生资格和generic10证明边界均保留。没有重算、重封存、更新source/selector/pin或产生当前资格。

## 尚未确认的部分

70处旧源码引用虽然有可读Git历史版本，原链接没有目标commit/hash，不能把选择的上下文或删除前版本冒充被引用时的确切原字节。逐项保留`expectedOriginalSHA256=null`及`originalRevisionVerified=false`；若原结论依赖确切版本，仍需原pin或独立新来源证明，不能降低原接受条件。

33份原样文档中14份在限定Git历史中未找到相同文档版本，但均与原发布清单/状态回执的SHA匹配；保留“原件可读、文档Git版本未知”。另1份历史报告由原capture程序进行路径脱敏，原完整路径/原字节哈希未确认，单列为派生材料。4处网站根链接的远端内容/可用性也保持未验证。

本轮没有“已确认永久丢失”的原件，也没有完全未定位的本地候选；这不等于所有历史原件已完备。原版本未确认、派生转换和远端未验证的限制继续可见，缺少这些事实的资格保持阻塞，既有PASS/FAIL/PARTIAL均不改写。

## 查找范围与停止条件

只读取原349项的目标、证实的原基址、原发布/回执/来源清单和本地Git。快照最多80个原路径变更commit、单blob最多8MiB；源码候选最多40个唯一路径、每文档/目标最多4个commit、200个blob、总64MiB，单Git命令20秒上限。实际源码查询为28个唯一路径、49个blob、8,268,529字节，约2.98秒。

临时库存先在明确保留位置命中两个原哈希相等副本；其他当前缺失路径都有有界Git版本，因此停止扩大查找。没有穷尽TAR/ZIP、全盘/tmp、用户目录或网络，也没有宣称归档中不存在其他副本。[完整查询边界/命令与限制](../../.evidence/scratch-artifact-remediation/docs/20261023T234325Z-t6bj9jrp/search-log.json)保存原件；任何未定位结果都只能称“本范围未找到”，不能判定永久丢失。

## 保留与回滚

本项只新增外部展示索引和治理说明，没有改现有历史文档/脚本、快照正文、产品源码、夹具、原生索引、支持矩阵或暂停状态。未执行旧脚本、Go/Node测试、构建、厂商、race、压力或真实清理。回滚只回退新索引/说明，保留查询与验证原件，不恢复旧执行入口。

来源捕获另按[捕获格式说明](historical-snapshot-index.md#捕获格式不等于原始下载字节)区分：reference.txt为HTML正文提取，pi-classifier.json为npm条目提取。文件pin不冒充原HTML/完整tgz的身份；本轮未定位这些上游原字节，原格式与未检索范围保持可见。
