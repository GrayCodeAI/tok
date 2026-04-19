$ErrorActionPreference = "Stop"

$snippetStart = "# >>> tok statusline >>>"
$snippetEnd = "# <<< tok statusline <<<"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$snippetBody = "if (Get-Command tok -ErrorAction SilentlyContinue) { `$env:PROMPT = `"$( & '$scriptDir\tok-statusline.ps1' 2`$null ) `$env:PROMPT`" }"

function Add-Snippet($filePath) {
  if (!(Test-Path $filePath)) { New-Item -ItemType File -Path $filePath | Out-Null }
  $content = Get-Content -Raw -Path $filePath
  if ($content -like "*tok statusline*") {
    Write-Host "Already configured: $filePath"
    return
  }
  Add-Content -Path $filePath -Value "`n$snippetStart`n$snippetBody`n$snippetEnd"
  Write-Host "Configured: $filePath"
}

Add-Snippet "$HOME\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
Write-Host "Done. Restart PowerShell."
