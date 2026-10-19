# 当前源码十一组旧输入复核（无 Key）

2026-10-19，测量来源751ad94；AI代码与6a24ba3相同，未合入后续续期实现。当前正式资格0/11；全部实际数字退出1，0 Key读取、0厂商请求。该结果不声称旧场景重新执行失败，也不改写旧报告历史。

M3三项调用当前`minimaxverify -formal`；GLM三项调用`supportmatrix -matrix <独立临时矩阵副本>`；视觉两项调用`visualeval -verify ... -require-live`；protocol调用`jevformal -protocol ... -effects ... -out <新诊断输出> -require-live`。退出分别为来源集合变化、GLM执行来源漂移或缺旧Messages live-report、视觉source set drift、protocol formal source set drift。

文本`chineseeval -verify`只验证录制时完整性，不能证明当前执行来源。因此保存一个只调用既有公开`release.ChineseEffectsEvidence`的独立main入口；Go overlay仅用于构建该入口，未修改磁盘生产源。它使用与公开releasegate相同的currentEvaluationSource判断，两个实际结果均为FAIL/source stale。入口源码、overlay与构建/执行数字收据全部保留；这不是新的正式发布CLI认证。

[result.json](result.json)记录11项逐项拒绝原因；源码压缩快照、6个既有公开CLI的二进制pins、所有实际命令/stdout在本目录。冻结二进制仍在主仓库`.evidence/current-source-eleven-precheck-20261019-1/bin/`。执行后完整1061来源文件集合与Git751ad94保持，原正式矩阵及四GLMselector仍匹配票18执行前pins。旧M3正式39/36与旧能力续期648/648均未改变。

后续正式续期须独立新预算与来源/持久账本授权。预计189预留/186HTTP，上限216/252；目前未批准、未取Key、未建付费intent。该前置检查不会授予十一组当前正式资格，票18单独取消场景PASS也不替代完整22场景。
