# 原生证据路径修正复核 v3

源码 pin：`8f77ecc8d126ce8d1fc9006c3171bd63f79c493d`。原v1/v2审查保持；本次仅修正省略effort的测试路径标签，并以真实producer/collector/原生releasegate锁定证据消费接缝。业务及导入器合同保持。

## Standards

**Standards: no material findings** in the narrow evidence delta at `8f77ecc8d126ce8d1fc9006c3171bd63f79c493d`.

The empty effort now receives the explicit evidence label `default`; the actual option value and wire omission remain unchanged. No production implementation or native consumer validation was weakened.

The new public regression invokes the actual option producer through the existing collector, builds and executes the native release CLI, and separately asserts source/E2E acceptance and live/race/pressure refusal. It retains real process argv, exits, binary identity, receipt/hash pointers, manifest and consumer report. Its exact producer selector avoids recursion; contexts/test timeouts and the existing fresh-HOME/minimal-environment helper bound execution and isolate ambient credentials.

Ticket09 and its versioned index preserve the rejected original receipt/record and distinguish collector PASS from native evidence acceptance. The final targeted report records 65 roots, 606 tests and 565 passing cases, with native exit1 and release=false because paid/human stages remain pending. Full frozen integration acceptance is explicitly still outstanding.

No additional actionable baseline smell identified. The scope addresses an observed native-consumer failure and accords with AGENTS.md §11's restraint requirement.

Read-only source/evidence-pointer review; no checks executed or repository files edited.


## Spec

复核 `e3ad302d...8f77ecc8d126ce8d1fc9006c3171bd63f79c493d` 的证据增量，新增规格发现0。

- `ai/e2e/xai_options_test.go:115–124` 只将空 effort 的路径标签改为 default。实际选项仍使用原 level，且第154–155行继续断言默认请求不带 reasoning；没有 wire 行为变化或场景删减。
- 新公开回归真实运行限定的选项测试，经既有 collector 取得原生 bundle，再编译并调用 releasegate。记录绑定实际 receipt/source/hash，必交的两个 default case 来自真实 producer，没有合成case或改写旧manifest。
- 消费者断言分别检查 source/e2e=true、live/race/pressure=false、release.pass=false及实际exit1，能够区分证据拒绝与正确的未完成阶段拒绝。
- 此增量没有生产实现、原生 importer或发布规则变更；旧无效证据保持历史事实。

本次仅静态只读复核，未执行测试。最终完整普通回执仍需重取；原五项修复的复核结论不变，真实live/正式及人工阶段仍未完成。


Standards剩余0，Spec剩余0；最终完整普通来源需以cases-v2重新取得回执。
