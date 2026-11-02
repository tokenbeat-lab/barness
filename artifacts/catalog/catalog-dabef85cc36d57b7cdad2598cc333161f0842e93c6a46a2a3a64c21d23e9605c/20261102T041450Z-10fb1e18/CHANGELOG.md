# 型号准入策略变更（未封存候选）

Kimi chat、xAI Responses chat、xAI image 和 TypeSafe SystemOne image/Box 的型号集合由审核目录决定，不再以具体名称准入。代码仍校验协议能力、非空唯一身份、Host 授权和响应/历史身份；Gemini wire 与 OpenAI 服务等级计价原行为保持。

真实 125 条目录和版本、维护源及基线不变。现有程序继续使用这份相同目录无需因本次策略升级；使用未来名称的新策略需先升级 SDK 并重建。后续同能力外部 JSON 经审核并由 Host 授权加载后可用于新 Client；编译内置默认目录更新仍需重建。无新厂商资格或软件版本发布。

在 `8d97cb6002722aceeb7fb3ccfaf5eb27e0991f11` 完整普通 PASS/exit0，329 固定必需场景通过，主包 17232 PASS、226 NOT_RUN，四个 audit 为0；生成、构建和 vet 各 exit0。53 个动态场景对应 48 个独立 packet；历史 archive 2189 文件及9条CLI检查已核验。文档收尾后2171个源码和84个输入与受测原件相等。live、race、压力分别因无新厂商事实、共享状态或取消变化、设计负载变化而不适用；见本批次 verification/。
