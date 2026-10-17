# 03：矩阵历史迁移 E2E 错误禁止后续合法真实重测

Status: resolved

## 问题与影响

2026-10-17，用户授权 TypeSafe 当前目录四项真实验证。51 Calls/Attempts、零重试，
各原始生产者、严格 verifier、正式封装及支持矩阵合并均成功。
当前矩阵仅 Jev1.13、Jev1.14 两条对应报告的记录更新，其他九行保持不变。

完整门禁的 normal 阶段却为 9,097 PASS、1 FAIL，失败仅为
`UC13-model-specific-matrix-history`。`ai/e2e/jev_formal_matrix_test.go` 从当前矩阵删除
1.14 行后，与最初纳入1.14前的历史快照逐字段比较，错误要求旧十行今后永远不变。
这会拒绝1.13合法的独立重测，不是本轮将1.14的PASS借给1.13。
发现时race仍在执行；用户追加授权后主动中止旧门禁，race未完成，不声明PASS。

## 可重复证据

- 完整执行根：`.evidence/ticket14-typesafe-20261017T055619.971418Z/`，
  `gate/offline/*/UC13-model-specific-matrix-history/assertions.json`。
- 无凭据单例复现：`go test ./ai/e2e -count=1 -run '^TestJevFormalMatrixHistory$'`，
  退出1，稳定复现同一断言；包和脱敏日志位于
  `.evidence/ticket14-typesafe-matrix-repro-20261017/`。
- 本轮 `matrix-before.json` 删除1.14后与旧fixture一致；`matrix-after.json` 中，
  旧十行仅1.13自身依据它的真实报告改变。原fixture和历史包没有改写。

## 修复与验收方向

- 保留原冻结fixture，不刷新它，不删除时间字段或跳过比较。
- 当前目录继续断言1.13/1.14独立身份。
- 将“其他型号历史不变”绑定到临时矩阵上的实际公开合并命令：输入1.14的报告
  只允许1.14自身变化；输入1.13报告只允许1.13自身变化；原矩阵和其他行保持一致。
- 先以公开E2E暴露合法连续更新场景的失败，再修改实现；保留新的manifest和审计。
- 完整normal/race、差分和原压力验收均不得缩减。

本轮先冻结源码和保留失败。修改Go测试也会触发正式包的完整源码一致性规则；
不能将现有真实包宣布对修改后的源码仍有效。需要用户确定是否修复并追加一次
TypeSafe四项、最多54 Calls/Attempts的真实采集；不得自动重跑收费任务。

2026-10-17：用户已明确授权修复并追加同一key四项一轮、最多54 Calls/Attempts。
首轮gate已主动中止，normal失败完整保留、race未完成；新轮另建证据目录。

修复从首次纳入提交`527f633aa735005c706933349cad0c2929671e21`提取after快照，
原JSON SHA256为`8efcf72986ce0f56e8f5cbcef23f8c826626206471a1d6a3f77dcd257f98a15c`，
仅压缩空白的fixture SHA256为`8d4cac102e5de9dfab71dc98c5fb16868723c764030ba583bba7db26f106bc59`。
原before未改；完整旧行比较移到迁移前后快照，公开CLI新增合法后续1.13重测、
其他十行精确保留和陈旧报告拒绝/文件不变。原单例RED退出1，新单例GREEN退出0，
两轴只读复核无阻断发现。待新源码下真实四项及完整门禁结束后再完成验收。

2026-10-17最终验收：源码`541ef2d`、发布执行`4d191fa`，新独立真实四项51/51、
零重试且全部严格核验通过。完整normal/race各9098 PASS、0 FAIL；两阶段UC13均30条
断言PASS，各861Pi/33UC14通过，原压力和图片设计负载未缩减。两种vet、P0、差分、
快照及聚合审计PASS，17个输入审计零发现。本缺陷按这些当前证据关闭。
首轮normal9097/1、race中止及旧51次仍保留；#14仍因其他9条真实路线及2项效果NOT_RUN
保持needs-info，整体报告pass=false/CLI退出1，不以本缺陷关闭替代整版通过。
结果与可重复封存入口见[TypeSafe验证](../../jev-1.14-multimodal-box/14-typesafe-real/verification.md)。
