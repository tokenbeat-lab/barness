# Grok JPEG 修复后的冻结用例清单 v3

[完整机器清单](cases-v3.json)。在 [v2](cases-v2.json) 的565个必交case、65个顶层入口上增量核对：只修改既有受控候选进程case断言，没有新增、删除或重命名case/入口。515个xAI和50个旧ID回归仍全部必交。

- 代码提交：`00e7612c95f53dfaab5d5c604e62e3f21550d649`；集成合并：`cefa01f5607e219b60e115acf3378703ebdfa4cb`。
- `ai/e2e/xai_live_process_test.go` 实际源码哈希已更新；其余清单来源文件按当前字节重新核对。
- 既有 `E17-xai-process-finite-controlled-plan` 从真实受控HTTP捕获解码用户PNG/JPEG及工具结果PNG/JPEG，断言均至少512像素。旧8×8 JPEG在用户、工具两处实际RED；修复为32×32后整条公共进程GREEN。
- 回执：`.evidence/barness-ai-grok/jpeg-fixture-e2e-v1.json`；RED与GREEN原包保留。
- 本清单是静态枚举，不是测试PASS声明；最终完整普通验收另行运行并逐项核对全部565case/65入口。
