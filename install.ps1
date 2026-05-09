$ErrorActionPreference = "Stop"

$Repo = if ($env:DIDA_REPO) { $env:DIDA_REPO } else { "DeliciousBuding/dida-cli" }
$Version = $env:DIDA_VERSION
$InstallDir = if ($env:DIDA_INSTALL_DIR) { $env:DIDA_INSTALL_DIR } else { Join-Path $HOME ".local\bin" }

function Get-Platform {
  if (-not $IsWindows -and $PSVersionTable.PSEdition -eq "Core") {
    if ($IsLinux) { return "linux" }
    if ($IsMacOS) { return "darwin" }
  }
  return "windows"
}

function Get-Arch {
  switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()) {
    "x64" { return "amd64" }
    "arm64" { return "arm64" }
    default { throw "unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
  }
}

if (-not $Version) {
  $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
  $Version = $latest.tag_name
}
if (-not $Version) {
  throw "could not resolve latest release version"
}

$OS = Get-Platform
$Arch = Get-Arch
$Ext = if ($OS -eq "windows") { "zip" } else { "tar.gz" }
$Binary = if ($OS -eq "windows") { "dida.exe" } else { "dida" }
$Asset = "dida_${Version}_${OS}_${Arch}.${Ext}"
$BaseUrl = "https://github.com/$Repo/releases/download/$Version"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("dida-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
  Write-Host "Installing DidaCLI $Version for $OS/$Arch"
  $ArchivePath = Join-Path $TempDir $Asset
  $ChecksumsPath = Join-Path $TempDir "checksums.txt"
  Invoke-WebRequest -Uri "$BaseUrl/$Asset" -OutFile $ArchivePath
  Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath

  $Line = Get-Content $ChecksumsPath | Where-Object { $_ -match "\s$([regex]::Escape($Asset))$" } | Select-Object -First 1
  if (-not $Line) { throw "checksum not found for $Asset" }
  $Expected = ($Line -split "\s+")[0].ToLowerInvariant()
  $Actual = (Get-FileHash -Algorithm SHA256 $ArchivePath).Hash.ToLowerInvariant()
  if ($Actual -ne $Expected) { throw "checksum mismatch for $Asset" }

  $OutDir = Join-Path $TempDir "out"
  New-Item -ItemType Directory -Path $OutDir | Out-Null
  if ($Ext -eq "zip") {
    Expand-Archive -Path $ArchivePath -DestinationPath $OutDir -Force
  } else {
    tar -xzf $ArchivePath -C $OutDir
  }

  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  $Source = Get-ChildItem -Path $OutDir -Recurse -Filter $Binary | Select-Object -First 1
  if (-not $Source) { throw "binary not found in archive" }
  $Target = Join-Path $InstallDir $Binary
  Copy-Item $Source.FullName $Target -Force

  $PathParts = ($env:PATH -split [System.IO.Path]::PathSeparator)
  if ($PathParts -notcontains $InstallDir) {
    Write-Host "PATH note: add $InstallDir to PATH to run dida from any shell."
  }

  & $Target version
  & $Target doctor --json
}
finally {
  Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
}
