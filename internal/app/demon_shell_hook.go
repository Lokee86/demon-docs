package app

import "encoding/base64"

const powershellHookScript = `# Demon Docs shell integration. Add: Invoke-Expression (& ddocs demon __shell-hook powershell)
$global:__DdocsDemonRepo = ""
$global:__DdocsDemonToken = ""
function global:Leave-DdocsDemon {
  if ($global:__DdocsDemonRepo -and $global:__DdocsDemonToken) {
    & ddocs demon __leave $global:__DdocsDemonRepo $global:__DdocsDemonToken *> $null
  }
  $global:__DdocsDemonRepo = ""
  $global:__DdocsDemonToken = ""
}
function global:Invoke-DdocsDemonHook {
  $candidate = (Get-Location).Path
  $status = @(& ddocs demon --status $candidate 2>$null)
  $repo = ($status | Where-Object { $_ -like "repository: *" } | ForEach-Object { $_ -replace '^repository: ', '' } | Select-Object -First 1)
  if ($repo -eq $global:__DdocsDemonRepo) { return }
  Leave-DdocsDemon
  if (-not $repo) { return }
  $enter = (& ddocs demon __enter $repo shell 2>$null | Select-Object -Last 1).Trim()
  $token = if ($enter -match 'token=([^ ]+)') { $Matches[1] } else { "" }
  $claimed = if ($enter -match 'claimed=([^ ]+)') { $Matches[1] } else { "false" }
  if (-not $token) { return }
  $global:__DdocsDemonRepo = $repo
  $global:__DdocsDemonToken = $token
  $after = @(& ddocs demon --status $repo 2>$null)
  $count = ($after | Where-Object { $_ -like "active shells: *" } | ForEach-Object { $_ -replace '^active shells: ', '' } | Select-Object -First 1)
  if ($claimed -eq "true") { Write-Host "document demon summoned for $repo" }
  if ($count -eq "1") { Write-Host "1 active shell feeding the demon" } else { Write-Host "$count active shells feeding the demon" }
}
if (-not (Get-Variable __DdocsOriginalPrompt -Scope Global -ErrorAction SilentlyContinue)) {
  $global:__DdocsOriginalPrompt = $function:prompt
  function global:prompt { Invoke-DdocsDemonHook; & $global:__DdocsOriginalPrompt }
}`

// powershellHookOutput deliberately emits one physical line. Windows PowerShell
// converts multiline native-command output into Object[], which made the
// documented Invoke-Expression installation command fail before the hook ran.
func powershellHookOutput() string {
	encoded := base64.StdEncoding.EncodeToString([]byte(powershellHookScript))
	return "Invoke-Expression ([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('" + encoded + "')))\n"
}
