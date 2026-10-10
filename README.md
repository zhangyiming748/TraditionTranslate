```shell
SRC="/Volumes/整理/分割后"; DST="/Users/zen/github/TraditionTranslate/srt"; cd "$SRC" && find . -type d -exec mkdir -p "$DST/{}" \; && find . -name '*.srt' -exec cp {} "$DST/{}" \;
```
# TraditionTranslate
传统的翻译工具，但是尽量使用工作流实现。
