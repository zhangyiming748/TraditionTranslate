```shell
SRC="/Users/zen/Movies/Downie"; DST="/Users/zen/github/TraditionTranslate/srt"; cd "$SRC" && find . -type d -exec mkdir -p "$DST/{}" \; && find . -name '*.srt' -exec cp {} "$DST/{}" \;
```
# TraditionTranslate
传统的翻译工具，但是尽量使用工作流实现。

## 使用方法

把字幕文件放到 `srt/` 目录下（支持子文件夹，一个子文件夹视为一个视频），然后运行：

```shell
# 使用默认目录 srt/
go run main.go

# 指定字幕目录
go run main.go -dir /path/to/subs
```

### 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-dir` | `srt` | 字幕文件存放目录。不传时使用项目下的 `srt/` 目录。 |

> 工作流中直接 `go run main.go` 即可，使用默认的 `srt/` 目录。
> 本地运行时可通过 `-dir` 指定任意目录。

