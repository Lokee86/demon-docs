param(
    [string]$Target
)

$ErrorActionPreference = "Stop"
$Source = Join-Path $PSScriptRoot "fixture"
$WorkspaceRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
if (-not $Target) {
    $Target = Join-Path $WorkspaceRoot "demon-docs-adoption-demo"
}

if ((Test-Path (Join-Path $Target ".ddocs")) -and (Get-Command ddocs -ErrorAction SilentlyContinue)) {
    & ddocs demon run --false $Target *> $null
    Start-Sleep -Seconds 1
}
if (Test-Path $Target) {
    Remove-Item -Recurse -Force $Target
}
New-Item -ItemType Directory -Path $Target | Out-Null
Copy-Item -Path (Join-Path $Source "*") -Destination $Target -Recurse -Force
Copy-Item -Path (Join-Path $Source ".docignore") -Destination $Target -Force

Write-Host "Reset Demon Docs adoption demo at $Target"
Write-Host "Next: Set-Location '$Target'; ddocs init --root docs"
