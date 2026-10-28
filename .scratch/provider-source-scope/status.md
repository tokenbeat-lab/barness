# 三条路线来源修正当前状态

已完成本次离线适用性修正。三条目标路线均通过最终标准 schema 2 消费，真实子进程退出 0；原测量不改写。本结论不授予新 live 执行权限，整体发布仍为 release.pass=false。

| 路线 | 复用真实场景 | 新的当前适用性 |
| --- | --- | --- |
| TypeSafe / Jev 1.14 | 12 个正式协议场景，另核验 12 个效果样本 | APPLICABLE / PASS |
| 智谱 Responses / GLM 5.3、5.2 | 22 个 | APPLICABLE / PASS |
| Z.AI Responses / GLM 5.3、5.2 | 18 个 | APPLICABLE / PASS |

## 改动与依据

按 [逐项审阅](source-review.md) 区分实际产品变化、其他协议/Provider 分支、注释和旧来源采集遗漏。Admission Provider 归属、缓存键 v2、效果 host 的预算前置确实是行为变化，按实际路径及当前普通测试判断；40 个 Responses 旧请求全部无 SessionID/cache key，本次不增加缓存命中或计费事实。

修正 chat.go / anthropic.go / responses.go 的协议归属；来源审阅只接受固定路线的确切前后字节。视觉消费者仍核验全部原 source/ 字节、合法原清单及实际评测执行/评分代码，不再让其他 Provider 或只读消费者的整套源码变化导致失效。GLM/地区来源采集包含未知产品子包，新增或删除、未来同文件改动和依赖变化仍拒绝。

智谱 clock/probe 旧清单遗漏由原完整执行快照追溯，绑定原授权/构建/二进制/命令；仅允许三个固定旧清单的哈希和两个固定文件。为验证该共用规则，另只读回归智谱 Chat/Messages，均 PASS；未要求其他 Provider 再次 live。90 条审阅记录包括 83 条产品字节差异、1 条评测 host 变化、6 条原清单缺项恢复。

## 最终验证

- 完整普通入口为 [实际选择](../../.evidence/provider-source-scope/full-ordinary-selection-v3.json)，受影响包覆盖 ./ai/e2e、./ai、release、supportmatrix、minimaxlive、jevformal、visualeval；[回执](../../.evidence/provider-source-scope/e2e/20261027T111447.586011Z-28bkc2h1/receipt.json) exit0 / PASS，墙钟 611.002s，179 个选定输入。主包 15437 个库存项：15223 PASS、214 NOT_RUN；50 项新增来源检查实际 PASS。辅助视觉 25、无厂商 GLM 32、MiniMax 77+6 全部 PASS，所有 audit findings=0。
- 工具进程套件 56 项、对应 go vet 和 releasegate 编译均 PASS，详见 [验证与保留索引](../../.evidence/provider-source-scope/verification-and-retention-v1.json)。所有最终回执的完整 source 集合与当前相等。
- [当前公共回放](../../.evidence/provider-source-scope/public-replay-v5/receipt.json)：40/40 请求结构与捕获相同；36 个正常响应的内容/用量/stop/response ID 一致。4 个取消保留截断响应边界，请求仍核验，取消时序/已观察帧/Observer 由原生消费者及当前受控普通测试验证，不伪造完整结束。
- [最终标准消费者](../../.evidence/provider-source-scope/consumption-v3/report.json)：三条目标及两个智谱回归均 pass=true、schema2、exit0，绑定二进制、完整输入、原生结果和真实数字退出。

HEAD 保持 `69ff4af880e68ef3ca245e5128bdae5e59a2ef7b`，本次未提交且保留此前任务/用户修改。实际来源 2007 文件，canonical compact/sorted UTF-8 文件映射 SHA256 `5d3c49ae55c67810595e2bab54f39cf059377f59224b4ae9ebf4da0b79481d2f`。本次 10 个来源文件变化均属消费者/来源采集/测试/数据，产品运行时代码变化 0；不能只用 HEAD 代表受测字节。

## 留存、缺口与阶段边界

[保护核验](../../.evidence/provider-source-scope/history-protection-check-v1.json)：9295 个历史文件全部不变，无缺失；公共授权/账本六处存储 8156 文件，无新增/删除/修改。本轮没有真实 claim、旧授权复活或暂停恢复。原支持矩阵、候选目录、原报告/失败/时间/授权和账本全部保留。

先失败后实现及所有失败/中止见 attempts-index-v1.json：原来源/根协议范围/未知产品子包均有 RED；视觉清单替换反例 FAIL 后补校验，完整普通第一次中止；发现 GLM 两个漏项后追溯原件、增加严格反例，第二次未完成全包中止；两份 CANCELLED 及智谱中间消费 FAIL 均保留。回放初版缺 Policy、第二版未复现四个原控制 hook 的失败也保留，均不改称厂商失败。

本次三路线最小新增付费补测清单为空，live 执行为 NOT_RUN；只读消费者未改变共享写状态或负载，race/压力影响 NOT_APPLICABLE，执行仍 NOT_RUN。未启动任何付费 live、race、压力或暂停轮次。Kimi 六路线各缺 xai-origin-history 场景、MiniMax 国际 Messages 原 incomplete/FAIL 仍保留，不由本次绕过。后续真实执行必须另行明确范围与授权。

[新适用性判定](../../.evidence/provider-source-scope/decisions-v1.json) 和 [合并资格索引](../../.evidence/provider-source-scope/current-qualification-index-v3.json) 引用旧原件。本轮三条补齐后是 23 个路线/型号证据条目，不能写成 23/29 条路径已交付；TypeSafe 两型号共用一条操作/协议路径，整体发布与人工阶段仍独立判断。

回滚仅撤销本次协议归属、确切字节审阅/遗漏恢复和视觉来源消费修正及对应测试，恢复此前保守拒绝；不得撤回已有历史测量、授权、claim、账本或删除原件，不覆盖工作区其他修改。
