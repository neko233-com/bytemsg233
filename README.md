# bytemsg233

`bytemsg233` 是一个面向 Agent 和客户端 SDK 的二进制 schema 工具链：单文件 `.bmsg` 定义、多语言代码生成、双语注释、紧凑编码，以及现在内置的各语言对象池支持。

## 当前定位

- 单文件 schema：不走 `.proto` import 链，Agent 和人都更容易读。
- 多语言输出：Go、C#、Java、TypeScript、Python。
- 注释原生生成：message / class 注释、field 注释会按 `--locale` 输出。
- 对象池内建：每门语言都会生成自带的 acquire / release / reset 机制，减少 client 端重复造轮子和对象残留。
- enum 原生适配：目标语言会生成更顺手的 enum helper，而不是把业务层逼回裸整数。

## 快速开始

```bash
# 初始化 schema
bytemsg233 init user

# 生成多语言 SDK
bytemsg233 compile user.bmsg \
  -l go,csharp,java,typescript,python \
  -o ./gen

# 输出中文注释
bytemsg233 compile user.bmsg -l go --locale zh -o ./gen
```

## 文档规范

这个仓库从现在开始按下面的边界维护文档：

- `README.md`、`docs/*.md` 面向开发者、维护者、协作者。
- `docs/**/*.html` 面向最终用户、演示访问者、产品展示场景。
- `.html` 必须是可独立阅读的完整页面，不允许依赖或引用 `.md` 内容。
- `.html` 可以链接站内其它 `.html` 页面，也可以直接内嵌说明，但不能把“去看某个 Markdown”当成主路径。
- `.md` 可以解释实现、约束、流程、计划、CLI 细节；`.html` 负责可视化表达、品牌感和直接可读性。
- 当同一主题同时存在 `.md` 和 `.html` 时：
  `.md` 负责源码仓库语境；
  `.html` 负责外部访问语境；
  两者内容可以同步，但不能互相当作渲染依赖。

## 生成代码约定

### 注释

- schema 中的 message 描述会生成类 / 结构体注释。
- schema 中的字段描述会生成字段注释。
- `--locale zh` 输出中文注释，默认输出英文注释。

### 对象池

每门语言都内置了一套原生风格的对象复用接口：

- Go：`AcquireUser()` / `ReleaseUser()` / `(*User).Reset()`
- C#：`User.Rent()` / `User.Return()` / `user.Release()` / `user.Reset()`
- Java：`User.acquire()` / `User.release()` / `user.reset()`
- TypeScript：`User.acquire()` / `user.release()` / `user.reset()`
- Python：`User.acquire()` / `user.release()` / `user.reset()`

默认 reset 行为会把标量恢复为零值，把 `list` / `map` / `bytes` 重建为干净状态，避免复用对象时残留旧数据。

### Enum 适配

生成代码会优先贴近目标语言自己的 enum 使用方式：

- Go：生成 `String()`、`Parse<Type>()`、`IsValid()`
- C#：生成 `FromValue()` 和 `IsDefinedValue()`
- Java：生成 `fromValue()` 和 `isDefined()`
- TypeScript：生成 `enum` + `namespace fromValue()`
- Python：生成 `IntEnum` + `from_value()`

目标是让业务代码一直写 enum，而不是在业务层到处传裸整数。

## Schema 示例

```bmsg
schema: bymsg/v1
package: com.example.game

enum HeroState {
    IDLE = 0
    MOVING = 1
    DEAD = 2
}

message Hero {
    uint32 id = 1 // "英雄 ID" | "Hero ID"
    string name = 2 // "名称" | "Name"
    list<string> tags = 3 // "标签" | "Tags"
    map<string, string> attrs = 4 // "属性" | "Attributes"
    HeroState state = 5 // "状态" | "State"
}
```

## 输出特点

- Go 生成 `struct`、字段注释和 `sync.Pool` 风格池化辅助函数。
- C# 生成 `sealed class`、XML 注释、原生集合默认值和 `ConcurrentBag` 对象池。
- Java 生成独立 `*.java` 文件、getter/setter、Javadoc 和 `ConcurrentLinkedQueue` 对象池。
- TypeScript 生成可实例化 `class`，不再只是 `interface`，并自带轻量池实现。
- Python 生成 `dataclass`、字段注释、默认值和类级对象池。

## 目录

- 开发者文档：[docs/使用文档.md](docs/使用文档.md)
- 在线演示页源码：[docs/demo/index.html](docs/demo/index.html)

## 子仓库

主仓库通过 Git submodule 统一挂载各端接入库和编辑器插件：

- `libs/typescript`: https://github.com/neko233-com/bytemsg233-lib-typescript
- `libs/csharp`: https://github.com/neko233-com/bytemsg233-lib-csharp
- `editors/vscode`: https://github.com/neko233-com/bytemsg233-plugin-vscode
- `editors/jetbrains`: https://github.com/neko233-com/bytemsg233-plugin-jetbrains

```bash
git submodule update --init --recursive
```

## 开发

```bash
go test ./...
```
