#!/bin/bash
# 检查字幕文件的正文行（每块第 3 行）是否超过 100 个字符
# 字幕格式：第1行序号、第2行时间轴、第3行正文、第4行空行
# 正文行位置：3, 7, 11, 15, ...（即 lineNum % 4 == 3）
#
# 用法: ./check_srt.sh [字幕目录]
#   字幕目录: 可选参数（脚本的第一个位置参数），默认为 srt
#            工作流中传入绝对路径，例如: ./check_srt.sh "$GITHUB_WORKSPACE/srt"
# 退出码：0 = 未发现问题（无任何问题输出）；1 = 发现超长正文行

export LC_ALL=en_US.UTF-8

srt_dir="${1:-srt}"

if [ ! -d "$srt_dir" ]; then
    echo "错误: 字幕目录不存在: $srt_dir" >&2
    exit 2
fi

# 将所有问题行收集到 result 中（管道会使 while 在子 shell 中执行，
# 无法通过变量计数，因此改为收集输出后统一判断）
result=$(find "$srt_dir" -name "*.srt" -type f | sort | while IFS= read -r file; do
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
done)

if [ -n "$result" ]; then
    # 有输出即表示发现问题：打印问题明细并以非零码退出，阻断工作流
    printf '%s\n' "$result"
    exit 1
fi

# 无任何输出、退出码 0 表示检查通过
exit 0
