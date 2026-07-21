param(
    [string]$Target
)

$ErrorActionPreference = "Stop"
$Source = (Resolve-Path (Join-Path $PSScriptRoot "fixture")).Path
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$WorkspaceRoot = Split-Path -Parent $RepoRoot
if (-not $Target) {
    $Target = Join-Path $WorkspaceRoot "demon-docs-adoption-demo"
}
$Target = [System.IO.Path]::GetFullPath($Target)
$RepoPrefix = $RepoRoot.TrimEnd('\') + '\'

if ($Target -eq $RepoRoot -or $Target.StartsWith($RepoPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw @"
Refusing to create the disposable demo inside the Demon Docs checkout.
Tracked source fixture: $Source
Requested target:       $Target
Use the default sibling target or another directory outside $RepoRoot.
"@
}

Write-Host "Tracked source fixture: $Source"
Write-Host "Disposable workspace:   $Target"
Write-Warning "The disposable workspace will be deleted and recreated."

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

Write-Host ""
Write-Host "Demo workspace ready."
Write-Host "Open this directory as the Obsidian vault: $Target"
Write-Host "Do NOT open the tracked fixture under the Demon Docs checkout."
Write-Host "Next: Set-Location '$Target'; ddocs init --root docs"
