# barness

## 项目定位

barness 是 [pi](https://github.com/earendil-works/pi) 的 Go（Golang）复刻项目，兼顾本地和远端使用场景。

## 开发约定

- 实现复刻功能时，查阅 pi 对应的文档与源码，以其行为作为参考；有意调整的行为应说明原因。
- 设计和修改功能时，同时考虑本地与远端场景，明确各场景的适用条件和差异。
- 按 Go 的惯用方式组织代码；目录结构、依赖和接口随具体需求确定。
- 将改动限定在当前任务范围内；新增模块、脚手架或基础设施应有明确的任务依据。
- 修改 Go 代码后使用 `gofmt`，并运行与改动相关的测试；交付时说明验证结果及未验证项。

## Agent skills

### Issue tracker

Issues and specs live as local markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Uses the five default label names (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
