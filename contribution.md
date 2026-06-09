# 贡献指南

感谢你愿意为 ZeroBot-Plugin 贡献代码。这个仓库是 ZeroBot 的插件合集，维护重点是：插件之间边界清晰、默认构建可通过、用户可按 README 直接启用和运行。

## 开始之前

1. Fork 本仓库，并从 `master` 拉出一个新分支。
2. 确认改动范围：新增或修改插件时，代码放在 `plugin/<name>/`；相关运行数据、静态资源放在 `data/<Name>/`。不要把本地运行生成的数据库、日志、二进制文件或临时目录提交上来，除非它们是插件运行所必需的版本化资源。
3. 如果改动会影响启动参数、插件启用方式、用户命令或外部依赖，请同步更新 `README.md`。

## 项目结构

- `main.go`：程序入口，同时通过空导入注册各插件。插件优先级与导入顺序相关。
- `plugin/<name>/`：插件实现目录，一个插件对应一个 Go package，目录名和包名使用短小的小写名称。
- `data/`：共享运行资产、词库、图片、数据库等资源。
- `winres/`：Windows 资源生成相关文件。
- `.github/workflows/`：PR、push、lint、构建和运行检查。
- `run.sh` / `run.bat`：本地快速启动脚本，流程为 `go mod tidy`、`go generate main.go`、`go run`。

## 本地开发流程

推荐先跑通一次完整流程：

```sh
go mod tidy
go generate main.go
go run -ldflags "-s -w" main.go
```

发布风格构建可使用：

```sh
go build -trimpath -ldflags "-s -w" .
```

如果你改动了生成资源、嵌入资源或多个插件，可改跑：

```sh
go generate ./...
```

## 插件贡献约定

新增插件时请尽量保持现有插件风格：

1. 在 `plugin/<name>/` 下创建短小小写的包名，例如 `plugin/rsshub`、`plugin/minecraftobserver`。
2. 使用 `control.AutoRegister` 注册插件，补齐 `Brief`、`Help`、`PrivateDataFolder` 等用户可见信息。
3. 在 `init` 中注册命令处理逻辑，并尽量把解析、存储、渲染、外部 API 调用拆成可测试的小函数。
4. 从 Handler 中抽出的函数只传它真正需要的值，例如 `question`、`gid`、`temperature`、抽卡结果或配置项。不要为了省事把 `*zero.Ctx` 直接传进业务函数；如果逻辑必须直接操作 `ctx`，通常说明它应该留在 Handler 里。
5. 调用大模型、外部 API 或复杂渲染时，先构造清晰的输入数据和 prompt，再把发送消息、读取用户身份、权限判断等上下文操作放在 Handler 边界处理。这样 review 时更容易确认数据流，也更容易补测试。
6. 需要持久化数据时，优先使用插件私有数据目录或 `data/<Name>/`，不要散落到仓库根目录。
7. 在 `main.go` 中添加空导入，并放到合适的优先级位置。
8. 在 `README.md` 的功能列表中补充插件说明、命令示例和必要依赖。

修改已有插件时，只改动相关插件目录和必要的资源文件。跨插件的公共行为变更需要在 PR 描述中说明影响范围。

## 代码风格

- Go 代码必须经过 `gofmt` 和 `goimports`。
- `goimports` 的本地前缀是 `github.com/FloatTech/ZeroBot-Plugin`。
- lint 禁止直接使用 `fmt.Errorf`；需要包装错误时优先使用 `github.com/pkg/errors` 的 `errors.Wrap` / `errors.Errorf` 风格。
- 用户可见错误、日志和帮助文本要说明可行动的信息，例如失败对象、失败原因和下一步。
- 保持命令匹配规则清晰，避免让一个插件拦截无关消息。
- 不要在 PR 中混入格式化、依赖更新和功能变更。确实需要时，拆成独立提交。

## 测试与验证

至少运行与你改动相关的测试：

```sh
go test ./plugin/<name>/...
```

提交 PR 前建议运行：

```sh
go test ./...
golangci-lint run
```

CI 还会执行 `go generate main.go`、`go mod tidy`、跨平台启动检查和构建检查。涉及外部服务、网络请求、定时任务、数据库或图片渲染的插件，请优先补充确定性的单元测试；无法自动化验证时，在 PR 描述里提供本地运行日志、截图或复现命令。

参考近期 `tarot` 大模型解析 PR 的 review 过程：维护者会重点看逻辑是否可测试、函数签名是否清楚、Handler 是否承担了过多业务逻辑。新增 AI 解析类功能时，至少测试 prompt 构造、文本分段、抽取结果结构等不依赖网络的部分；大模型请求本身可以留作薄封装并在 PR 中说明复用了哪套配置。

## 依赖与资源

- 新增或删除 Go 依赖后，运行 `go mod tidy` 并提交 `go.mod` / `go.sum` 的对应变化。
- 如果改动影响 Nix 构建，请同步更新 `flake.lock`、`gomod2nix.toml` 等相关文件。历史提交中这类改动通常使用 `chore(nix): ...`。
- 不要提交本地编译产物，例如 `*.exe`、临时 `zbp*` 二进制、日志文件或个人配置。
- 数据文件较大或来源不清时，请在 PR 中说明用途、来源和许可证风险。

## 提交信息

历史提交主要使用 Conventional Commit 风格。推荐格式：

```text
<type>(<scope>): <summary>
```

常用类型：

- `feat(scope): ...`：新增功能或插件能力。
- `fix(scope): ...`：修复 bug。
- `chore(lint): ...`：代码格式或 lint 修复。
- `chore(nix): ...`：Nix、gomod2nix 或依赖锁文件更新。
- `doc(README): ...` / `docs(scope): ...`：文档更新。
- `refactor(scope): ...` / `optimize(scope): ...`：不改变用户行为的重构或优化。

示例：

```text
feat(aichat): support config reasoning effort
fix(handou): 猜成语部分正确后未保存历史记录
feat(main): add option `-fb64`
chore(lint): 改进代码样式
chore(nix): update nixpkgs and gomod2nix.toml
doc(README): update ZeroBot version badge
```

scope 建议使用插件名或改动区域，例如 `aichat`、`handou`、`main`、`ci`、`nix`。release tag 和版本发布提交由维护者处理。

## Pull Request

发起 PR 前请检查：

1. PR 标题不要包含 `.go`，仓库 CI 会关闭这类标题。
2. 描述清楚行为变化、影响插件和用户可见差异。
3. 链接相关 issue；没有 issue 时写明问题背景。
4. 列出你实际跑过的验证命令，例如 `go test ./...`、`golangci-lint run`。
5. 用户可见输出变化请附截图、日志片段或命令示例。
6. 大改动请说明兼容性风险、迁移方式和回滚方式。
7. 收到 requested changes 后，优先按 review 点做最小范围修改；如果不同意某条意见，说明具体技术理由和替代方案。修改后在评论中对应说明“改了什么、为什么、跑了什么验证”。

PR 应尽量小而集中：一个 bug 修复、一个插件功能或一组强相关的维护变更。这样更容易 review，也更容易在 CI 失败时定位问题。
