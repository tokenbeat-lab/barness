# Provider 交付准备：MiniMax 旧证据与 Kimi 待批准范围

Status: claimed
Type: task

已授权：只读核验 MiniMax 国际 Messages 失败包中的其余13个成功场景；公共 Client 受控复现 thinking-on 的合法793字节前缀后断流，保留部分推理、错误/资源/零流内重试事实；列最小补采范围。整理六份非可执行的Kimi待批准范围及29条操作×Provider×协议路径交付矩阵。保留历史、授权、账本、旧失败和暂停状态；不签发付费grant、不读凭据、不启动厂商live/race/压力、不恢复暂停轮次、不激活目录。

实施克制：当前产品行为未发现缺陷，先复现并验证，不为单次真实连接中断修改adapter或增加重试。不新增通用组合/续期/控制器。旧13项的分析只返回逐项适用性，不绕过完整消费者、不改失败报告或补造第14项。

失败清单（先于测试）：①合法前缀后非EOF读错被误判成功或吞掉Thinking；②正常EOF缺终态错误归类不符；③流开始后重放、重复终态、Observer/Result分歧、body/许可未释放；④原件/目录/SDK/来源/授权/原时刻/ledger/原生语义不符；⑤工具结果未绑定本包实际前序、跨tenant/account恢复；⑥取消的read-ahead被冒充已观察内容；⑦control已恢复的初次失败被误列额外未补项；⑧六份proposal误变成可执行grant或引用错baseline；⑨矩阵把TypeSafe两个版本算两条路径、把受控/部分测量算正式交付；⑩缺失或非0实际退出被包装成通过。

验证范围：新公共E2E只用synthetic credential、无dialer的内存HTTP传输/错误注入与保存夹具；分析探针使用当前原生逐项校验函数、当前编码器重建与响应回放，输入为固定旧原件。源码/SDK无行为变更时普通沿用既有完整来源，通过新增定向普通及来源差异相接；若产品/消费者实现有必要改动，先RED后实现并扩大到完整受影响普通。race/压力候选只列不执行。所有新合成grant/claim若确需使用只允许controlled并写前断言私有Git namespace；本任务首选不生成任何grant/claim。
