$configDir = if ($env:TOK_CONFIG_DIR) { $env:TOK_CONFIG_DIR } else { Join-Path $HOME ".config\tok" }
$flagFile = Join-Path $configDir ".tok-active"

if (!(Test-Path $flagFile)) { return }
$mode = (Get-Content -Raw $flagFile).Trim()
if ([string]::IsNullOrWhiteSpace($mode) -or $mode -eq "full") {
  Write-Output "[TOK]"
  return
}
Write-Output ("[TOK:{0}]" -f $mode.ToUpper())
