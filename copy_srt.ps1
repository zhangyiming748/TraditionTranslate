# 把 A 目录下的文件夹结构和其中的 SRT 文件复制到 B 目录
# 保持子文件夹结构，只复制 .srt 文件，其他类型文件不复制
#
# 用法: .\copy_srt.ps1 <源目录> <目标目录>
#   源目录:   必选，包含 SRT 文件的目录（A）
#   目标目录: 必选，复制到的目录（B），不存在时自动创建
# 退出码：0 = 成功；2 = 参数或目录错误

param(
    [Parameter(Mandatory = $false, Position = 0)]
    [string]$Src,

    [Parameter(Mandatory = $false, Position = 1)]
    [string]$Dst
)

$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($Src) -or [string]::IsNullOrWhiteSpace($Dst)) {
    Write-Error "用法: $($MyInvocation.MyCommand.Name) <源目录> <目标目录>"
    exit 2
}

if (-not (Test-Path -LiteralPath $Src -PathType Container)) {
    Write-Error "错误: 源目录不存在: $Src"
    exit 2
}

$srcAbs = (Resolve-Path -LiteralPath $Src).ProviderPath
$dstAbs = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Dst)
if (-not (Test-Path -LiteralPath $dstAbs -PathType Container)) {
    New-Item -ItemType Directory -Path $dstAbs -Force | Out-Null
}
$dstAbs = (Resolve-Path -LiteralPath $dstAbs).ProviderPath

# 1. 先按源目录结构创建所有子目录（包括空目录，保证结构完整）
# 2. 再复制所有 .srt 文件到对应位置（已存在的同名文件会被覆盖）
$count = 0
Get-ChildItem -LiteralPath $srcAbs -Directory -Recurse -Force | ForEach-Object {
    $relative = $_.FullName.Substring($srcAbs.Length).TrimStart('\', '/')
    $targetDir = Join-Path $dstAbs $relative
    if (-not (Test-Path -LiteralPath $targetDir -PathType Container)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }
}

Get-ChildItem -LiteralPath $srcAbs -File -Recurse -Force -Filter '*.srt' | ForEach-Object {
    $relative = $_.FullName.Substring($srcAbs.Length).TrimStart('\', '/')
    $targetFile = Join-Path $dstAbs $relative
    $targetDir = Split-Path -Parent $targetFile
    if (-not (Test-Path -LiteralPath $targetDir -PathType Container)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }
    Copy-Item -LiteralPath $_.FullName -Destination $targetFile -Force
    $count++
}

Write-Host "完成: 已复制 $count 个 SRT 文件，从 $srcAbs 到 $dstAbs"
exit 0
