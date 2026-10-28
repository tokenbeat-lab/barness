复核 `e3ad302d...8f77ecc8d126ce8d1fc9006c3171bd63f79c493d` 的证据增量，新增规格发现0。

- `ai/e2e/xai_options_test.go:115–124` 只将空 effort 的路径标签改为 default。实际选项仍使用原 level，且第154–155行继续断言默认请求不带 reasoning；没有 wire 行为变化或场景删减。
- 新公开回归真实运行限定的选项测试，经既有 collector 取得原生 bundle，再编译并调用 releasegate。记录绑定实际 receipt/source/hash，必交的两个 default case 来自真实 producer，没有合成case或改写旧manifest。
- 消费者断言分别检查 source/e2e=true、live/race/pressure=false、release.pass=false及实际exit1，能够区分证据拒绝与正确的未完成阶段拒绝。
- 此增量没有生产实现、原生 importer或发布规则变更；旧无效证据保持历史事实。

本次仅静态只读复核，未执行测试。最终完整普通回执仍需重取；原五项修复的复核结论不变，真实live/正式及人工阶段仍未完成。
