#!/bin/bash
# 把 A 目录下的文件夹结构和其中的 SRT 文件复制到 B 目录
# 保持子文件夹结构，只复制 .srt 文件，其他类型文件不复制
#
# 用法: ./copy_srt.sh <源目录> <目标目录>
#   源目录:   必选，包含 SRT 文件的目录（A）
#   目标目录: 必选，复制到的目录（B），不存在时自动创建
# 退出码：0 = 成功；2 = 参数或目录错误

set -e
export LC_ALL=en_US.UTF-8

src="$1"
dst="$2"

if [ -z "$src" ] || [ -z "$dst" ]; then
    echo "用法: $0 <源目录> <目标目录>" >&2
    exit 2
fi

if [ ! -d "$src" ]; then
    echo "错误: 源目录不存在: $src" >&2
    exit 2
fi

# 转为绝对路径，避免后面 cd 进源目录后相对路径失效
mkdir -p "$dst"
src_abs=$(cd "$src" && pwd)
dst_abs=$(cd "$dst" && pwd)

# 1. 先按源目录结构创建所有子目录（包括空目录，保证结构完整）
# 2. 再复制所有 .srt 文件到对应位置（已存在的同名文件会被覆盖）
# find 的 -exec ... {} 形式对含空格/日文的文件名是安全的
cd "$src_abs"
find . -type d -exec mkdir -p "$dst_abs/{}" \;
find . -type f -name '*.srt' -exec cp {} "$dst_abs/{}" \;

count=$(find . -type f -name '*.srt' | wc -l | tr -d ' ')
echo "完成: 已复制 $count 个 SRT 文件，从 $src_abs 到 $dst_abs"
exit 0
