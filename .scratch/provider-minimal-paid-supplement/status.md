# 七个最小付费补采当前状态

2026-10-28：STOPPED。本轮首条MiniMax thinking-on及用户后续明确要求的一次同范围重试，均实际FAIL/exit1，故整个剩余付费序列停止。Kimi六路没有签发grant、没有调用，不恢复暂停轮次。失败原因均为原始SSE用量未报告cache_creation_input_tokens；场景正文/流完成，但固定单项计费用量合同未满足，不能取得补项资格，不能与原13项组合通过。

[交付与两次原件](../../.evidence/provider-minimal-paid-supplement/delivery-v1.json)、[原七路批准](../../.evidence/provider-minimal-paid-supplement/approval.json)、[原失败STOP](../../.evidence/provider-minimal-paid-supplement/STOP.json)、[仅一次MiniMax重试批准](../../.evidence/provider-minimal-paid-supplement/retry-minimax-v1/approval.json)。原STOP未清除，重试未恢复Kimi权限。

累计真实2调用/2HTTP、2份独立新grant，授权上限合计USD0.02、保守已知费率预留合计USD0.012288；这些不是已确认账单金额，实际账单未知。两次均零SDK/环境/请求/工具自动重试；第二次来自人类新指令。原全局USD0.61没有提高，剩余未签授权额度按授权上限计为USD0.59，不能擅自执行六份各USD0.10的Kimi预算或继续第三次MiniMax。

旧公共授权/账本8156文件的路径与字节全部保留；仅在真实Git授权域追加本次两份新claim对应的6个ID/scope/ledger文件。两份失败原生包、manifest、audit、封存、日志、数字退出、授权、消费与预算均保留。原v5事实确认修订1七份文件、冻结二进制、源码/SDK/型号/目录/价格保持，源码仍匹配已通过完整普通证据，不重建、不修改消费者，不重跑普通。

真实缺口仍为MiniMax thinking-on与Kimi六项；22/29路径完整适用，MiniMax13/14、Kimi各9/10，原54+13项适用性保持。release.pass=false。race/压力、Kimi目录激活、旧暂停恢复、额外Provider调用均未执行。

另发现两个失败包都有原生收尾完整性缺口：finishMiniMax先封存验证前报告，随后用量校验失败改写报告但未更新seal；各包只有minimax-report.json与seal不符，其他封存文件一致。原FAIL报告和seal均保持原样；仅在diagnostics路径复原验证前报告字节，并以原seal SHA核对成功，用于追溯收尾缺陷，不能作为PASS资格或替换原包。

下一步先离线审阅上述收尾缺陷、MiniMax缓存写计数省略语义与当前消费者规则，区分实际计费证据缺口和过宽拦截；现有捕获是否可复用另行判定。不得把缺失计数改写为0、把失败改成PASS或继续盲目付费重试；后续付费需重新明确范围及累计预算。
