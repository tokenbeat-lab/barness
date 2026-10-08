# 工单 17：门禁失败方式与设计

接缝沿用规格 Testing Decisions 与 ADR-0017：release.Evaluate/文件收集命令、
supportmatrix.Merge、公共 Client 的既有离线/宿主/压力 E2E、独立 live 进程。
门禁仅读取证据，不解析凭据或调用 Provider；不新增模型运行时或兼容路径。

实现前列出以下误放行输入，保留原有 21 种门禁失败方式：

- 缺 P07/P08/P09 或 P0 场景，NOT_RUN/FAIL 被忽略；删追溯映射逃避验收。
- pending 差分、缺记录、错误 pi 版本/commit/model-data hash、空 SDK/请求哈希或发现清单；扩展路线未登记或缺离线/显式 skip。
- 空/重复矩阵能力；删必交分类/生成/编辑，改为 UNSUPPORTED，替代型号或错误 operation/provider/API。
- 共享 adapter 或同 combo 的旧包替代当前完整运行；SDK、目录/价格版本或哈希漂移；矩阵与报告状态/时间/型号不一致。
- 只审计离线包而遗漏 race/live/压力/中文效果包，审计错误或非零发现，Observer 携带正文或凭据。
- 缺本地或云端混合设计负载实际报告；仅有场景 PASS，报告超内存预算、并发不足、用量 >75%、资源未释放。
- 缺中文真实完整评估，以 fixture/NOT_RUN/协议 PASS 代替；删除样本、答案、manifest 或 Observer，版本或配置漂移。
- 快照与代码不一致，缺 fixture 完整性清单，回退映射有名无 PASS；已发请求用量丢失或转到另一操作。

发布产物分别呈现离线、pi 差分、扩展 skip、各路线 live、设计负载和中文业务效果。
当前 FAIL 与外部阻塞保留；不得删场景、换型号、放宽预算或修改父规格掩盖失败。
代码审查固定起点：553f322be9d33a0e1bad0d488d7dbae2a7006d39。
