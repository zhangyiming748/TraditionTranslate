#!/bin/bash
# 检查 srt/ 下所有字幕文件的正文行（每块第 3 行）是否超过 16 个字符
# 字幕格式：第1行序号、第2行时间轴、第3行正文、第4行空行
# 正文行位置：3, 7, 11, 15, ...（即 lineNum % 4 == 3）

export LC_ALL=en_US.UTF-8

find srt -name "*.srt" -type f | sort | while IFS= read -r file; do
    line_num=0
    while IFS= read -r line || [ -n "$line" ]; do
        line_num=$((line_num + 1))
        if [ $((line_num % 4)) -eq 3 ]; then
            char_count=$(printf '%s' "$line" | wc -m | tr -d ' ')
            if [ "$char_count" -gt 100 ]; then
                printf "问题: %s 第%d行 (%d字符): %s\n" "$file" "$line_num" "$char_count" "$line"
            fi
        fi
    done < "$file"
done
