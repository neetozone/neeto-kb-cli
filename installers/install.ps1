$ErrorActionPreference = "Stop"

$BaseUrl = "https://neeto-downloads.s3.amazonaws.com/cli/NeetoKB/latest"
$InstallDir = "$env:LOCALAPPDATA\Programs\neetokb"

$Arch = if ([Environment]::Is64BitOperatingSystem) {
  if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
  Write-Error "Unsupported architecture"
  exit 1
}

$Archive = "neetokb_windows_${Arch}.zip"
$Url = "${BaseUrl}/${Archive}"

Write-Host "Downloading NeetoKB CLI for windows/${Arch}..."
$TmpDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
$ZipPath = Join-Path $TmpDir $Archive

Invoke-WebRequest -Uri $Url -OutFile $ZipPath

Write-Host "Extracting..."
Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force

Write-Host "Installing to ${InstallDir}..."
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item (Join-Path $TmpDir "neetokb.exe") -Destination (Join-Path $InstallDir "neetokb.exe") -Force

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
  Write-Host "Added ${InstallDir} to user PATH."
}

Remove-Item -Recurse -Force $TmpDir

Write-Host "NeetoKB CLI installed successfully. Restart your terminal and run 'neetokb --help' to get started."
