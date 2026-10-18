# 票13实际 legacy 十五项

失败边界：原始归档、成员或完整性清单被改写；路径越界、符号链接、重复/遗漏成员、超限解压；canonical/source/command/report pins不符；场景或预算求和不符；UNSUPPORTED被改成PASS；消费或质量指标被补造；M Plan NOT_RUN被误记已运行；Jev派生内容与原protocol/effects不同，均必须拒绝。

固定15项实际进程/集合全部PASS，实际180 calls / 180 HTTP Attempts。该数字不是180个场景PASS。OpenAI Chat为9 PASS、1 UNSUPPORTED：reasoning-history因Chat Completions无可重放的reasoning内容/签名，零调用、零Attempt；gpt-5.4-pro不可用探针是另一项PASS，用于验证错误分类。每个集合的固定场景身份、状态分布和实际调用另列于summary与scenarios。

| 集合 | 固定场景状态 | calls / Attempts | usage 完整 / 部分 / 未报告 |
|---|---|---:|---:|
| openai-responses | 8 PASS | 14 / 14 | 13 / 0 / 1 |
| openai-chat | 9 PASS, 1 UNSUPPORTED | 14 / 14 | 12 / 0 / 2 |
| anthropic-messages | 11 PASS | 20 / 20 | 20 / 0 / 0 |
| google-gemini | 6 PASS | 8 / 8 | 8 / 0 / 0 |
| deepseek-responses | 8 PASS | 10 / 10 | 9 / 0 / 1 |
| deepseek-chat | 11 PASS | 13 / 13 | 11 / 0 / 2 |
| openai-images | 3 PASS | 3 / 3 | 3 / 0 / 0 |
| google-interactions-image | 2 PASS | 2 / 2 | 0 / 2 / 0 |
| openai-decisions | 9 PASS | 9 / 9 | 8 / 0 / 1 |
| typesafe-classifier | 3 PASS | 3 / 3 | 2 / 0 / 1 |
| protocol114 | 12 PASS | 12 / 12 | 12 / 0 / 0 |
| text-jev | 24 answered | 24 / 24 | 24 / 0 / 0 |
| text-luna | 24 answered | 24 / 24 | 24 / 0 / 0 |
| visual-jev | 12 answered | 12 / 12 | 12 / 0 / 0 |
| visual-decisions | 12 answered | 12 / 12 | 12 / 0 / 0 |

中国区DeepSeek直连，其余这15项通过显式`http://127.0.0.1:7890`代理。GLM三路由[独立GLM归档](../actual-glm/README.md)保存：65/65，其中Responses失败；当前固定续期总轮245/245，加原300/300为当前MiniMax发布续期作用域545/545。更早独立GLM任务与MiniMax自身candidate/formal历史另存，不在该数值内；不授予新一轮调用。

原始包和canonical15、冻结source/plan/checks/review/authorization、M3.1三路NOT_RUN、parent命令、Jev派生support与原A成功导出/临时矩阵验证/默认矩阵导入记录一并保留。较大binary/no-key产物仅留原位置、hash和有界清单引用。原始归档存于common root新建`.evidence/minimax-retained/ticket-13-actual-legacy/`，不重复提交Git blob。

校验器复用并固定[实现归档校验器](../implementation/verify.py)的字节，独立重算原始Attempt/wire、usage下界与场景/预算；逐项验证原seal和审计、独立source review及canonical pins、派生包与原始内容。质量保留原metrics：中文24条/路、视觉12样本/路，均为固定合成数据，不能推及生产总体。partial或unreported保留未知，不等于免费。

```sh
python3 -B .scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/verify.py
python3 -B .scratch/barness-ai-minimax/evidence/ticket-13/actual-legacy/check-boundaries.py
```

在主工作树运行；其他工作树增加`--common-root <主工作树绝对路径>`。两者只读原始归档；负向检查仅修改临时副本，不读.env、不联网、不重放实际请求。完整发布门禁与中断/探针历史由另一个入口保存；本目录不声称整份spec通过。

实际最终Attempt共170 complete、2 partial、8 unreported；输入侧已知63176、输出已知9486，是明确口径的下界，不是账单。Jev/Luna中文各24/24正确；视觉两路common分类/存在性均8/8，Jev本地hooked4/4，Luna supplemental-context4/4；Jev box两条present的平均IoU为0.7268501277721419，absentFalsePositive为0。原始分母、box/置信度与延迟指标完整保留在summary的quality字段，不混并不同任务指标。
