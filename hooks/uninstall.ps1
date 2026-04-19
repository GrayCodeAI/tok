$ErrorActionPreference = "Stop"

$snippetStart = "# >>> tok statusline >>>"
$snippetEnd = "# <<< tok statusline <<<"
$profilePath = "$HOME\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"

if (!(Test-Path $profilePath)) {
  Write-Host "No PowerShell profile found."
  exit 0
}

$lines = Get-Content -Path $profilePath
$out = New-Object System.Collections.Generic.List[string]
$skip = $false
foreach ($line in $lines) {
  if ($line -eq $snippetStart) { $skip = $true; continue }
  if ($line -eq $snippetEnd) { $skip = $false; continue }
  if (-not $skip) { $out.Add($line) }
}
$out | Set-Content -Path $profilePath
Write-Host "Removed tok statusline snippet from profile."
