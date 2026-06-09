[CmdletBinding(SupportsShouldProcess = $true)]
param(
    [string]$Root,
    [switch]$Quiet
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Info {
    param([string]$Message)

    if (-not $Quiet) {
        Write-Host $Message
    }
}

if ([string]::IsNullOrWhiteSpace($Root)) {
    $scriptDir = if (-not [string]::IsNullOrWhiteSpace($PSScriptRoot)) {
        $PSScriptRoot
    } else {
        Split-Path -Parent $MyInvocation.MyCommand.Path
    }
    $Root = (Resolve-Path -LiteralPath (Join-Path $scriptDir '..')).Path
}

$resolvedRoot = (Resolve-Path -LiteralPath $Root).Path
$gitRoot = (& git -C $resolvedRoot rev-parse --show-toplevel 2>$null)
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($gitRoot)) {
    throw "Not a git repository: $resolvedRoot"
}

$gitRoot = (Resolve-Path -LiteralPath $gitRoot.Trim()).Path
$untrackedFiles = @(& git -C $gitRoot ls-files --others --exclude-standard -- plugin)
if ($LASTEXITCODE -ne 0) {
    throw 'Failed to list untracked plugin files.'
}

$dataDirs = New-Object 'System.Collections.Generic.HashSet[string]'
foreach ($file in $untrackedFiles) {
    if ($file -match '^plugin/[^/]+/data(/|$)') {
        [void]$dataDirs.Add($Matches[0].TrimEnd('/'))
    }
}

if ($dataDirs.Count -eq 0) {
    Write-Info 'No untracked plugin data directories found.'
    return
}

$removed = 0
$skipped = 0
foreach ($dir in ($dataDirs | Sort-Object)) {
    $trackedFiles = @(& git -C $gitRoot ls-files -- $dir)
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to inspect tracked files under $dir."
    }

    if ($trackedFiles.Count -gt 0) {
        Write-Info "Skipping $dir because it contains tracked files."
        $skipped++
        continue
    }

    $fullPath = Join-Path $gitRoot ($dir -replace '/', [System.IO.Path]::DirectorySeparatorChar)
    if (-not (Test-Path -LiteralPath $fullPath)) {
        continue
    }

    $resolvedFullPath = (Resolve-Path -LiteralPath $fullPath).Path
    $rootPrefix = $gitRoot.TrimEnd('\', '/') + [System.IO.Path]::DirectorySeparatorChar
    if (-not $resolvedFullPath.StartsWith($rootPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to remove path outside repository: $resolvedFullPath"
    }

    if ($PSCmdlet.ShouldProcess($resolvedFullPath, 'Remove untracked plugin data directory')) {
        Remove-Item -LiteralPath $resolvedFullPath -Recurse -Force
        Write-Info "Removed $dir"
        $removed++
    }
}

Write-Info "Removed $removed untracked plugin data director$(if ($removed -eq 1) { 'y' } else { 'ies' }); skipped $skipped."
