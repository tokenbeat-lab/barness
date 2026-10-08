# 工单 10：实现前失败清单与测试边界

已授权接缝为 Client / HookedClient.GenerateImages、受控 HTTP Provider 与宿主 resolver、回调和资源探针。测试全部从公共入口观察；不新增内部单元测试。

- 端点：生成保持 generations，参考图编辑走 edits JSON；有序参考图独立复制，mask 与选项在 resolver 前快照。
- 内联：严格标准 base64、MIME/文件头，拒绝路径、远程 URL、file_id、其他厂商引用、换行、无 padding、非规范尾位、损坏尾部。
- 容量：每 URL ≤20971520 字符；参考图 ≤min(16,型号上限,宿主限额)；mask 占宿主张数和单图限额；请求总字节先检查。
- mask：没有参考图、型号未允许、尺寸与第一张不同、格式/尺寸损坏；初始选项和回调追加都检查。
- 能力：保真度省略/允许值/显式 null；无编辑时拒绝；带来源的尺寸边长倍数、长宽比、像素及边长上下限；生成与编辑共用最终选项校验。
- 回调：增删参考图改变操作、改型号/endpoint/operation/stream、file_id/远程引用为 tenant_denied；未知字段/无效 schema/组合为 callback_failed；突破资源预算为 resource_limit；编码前拒绝巨大图/大量图。
- 响应：最终请求格式及真实响应格式/MIME/文件头，全部图片原子发布；2xx 解析、回调、读断失败不重放；保留模态用量及厂商请求 ID，body/call/permit 全部释放。

首批 2.5 型号能力、JSON 编辑真实接受性由工单 11 的真实证据确定；此处仅使用合成能力，不改支持目录。
