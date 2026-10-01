# Persona installer for Windows (PowerShell).
# Downloads a prebuilt binary from GitHub Releases, falling back to a
# `go build` from source. Installs to %LOCALAPPDATA%\Programs\persona and
# adds it to the user PATH.
#
# Quick install:
#   irm https://raw.githubusercontent.com/Yescript/persona/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo    = 'Yescript/persona'
$Binary  = 'persona'
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\persona"

function Info($m) { Write-Host "==> $m" -ForegroundColor Cyan }
function Warn($m) { Write-Host "!!  $m" -ForegroundColor Yellow }
function Die($m)  { Write-Host "xx  $m" -ForegroundColor Red; exit 1 }

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        'AMD64' { return 'amd64' }
        'ARM64' { return 'arm64' }
        'x86'   { Die 'unsupported arch: x86 (need 64-bit Windows)' }
        default { return 'amd64' }
    }
}

function Install-Local {
    # If we are inside a source checkout (extracted zip / git clone), build
    # straight from it — no network, no GitHub required.
    $candidates = @()
    if ($PSScriptRoot) { $candidates += $PSScriptRoot }
    $candidates += (Get-Location).Path
    foreach ($d in $candidates) {
        if ((Test-Path (Join-Path $d 'go.mod')) -and (Test-Path (Join-Path $d 'cmd\persona'))) {
            if (-not (Get-Command go -ErrorAction SilentlyContinue)) { return $false }
            Info "Found local source in $d — building from it ..."
            New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
            Push-Location $d
            try {
                go build -ldflags "-s -w" -o (Join-Path $InstallDir "$Binary.exe") ./cmd/persona
            } finally {
                Pop-Location
            }
            return $true
        }
    }
    return $false
}

function Install-FromRelease($arch) {
    $asset = "${Binary}_windows_${arch}.zip"
    $url   = "https://github.com/$Repo/releases/latest/download/$asset"
    $tmp   = Join-Path ([System.IO.Path]::GetTempPath()) ("persona_" + [guid]::NewGuid())
    New-Item -ItemType Directory -Force -Path $tmp | Out-Null
    $zip = Join-Path $tmp $asset
    Info "Downloading $asset ..."
    Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
    Expand-Archive -Path $zip -DestinationPath $tmp -Force
    $exe = Get-ChildItem -Path $tmp -Recurse -Filter "$Binary.exe" | Select-Object -First 1
    if (-not $exe) { throw "persona.exe not found in archive" }
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item $exe.FullName (Join-Path $InstallDir "$Binary.exe") -Force
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

function Install-FromSource {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Die 'no prebuilt binary and Go is not installed — get Go from https://go.dev/dl/'
    }
    Info 'Building from source with Go ...'
    $tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("personasrc_" + [guid]::NewGuid())
    if (Get-Command git -ErrorAction SilentlyContinue) {
        git clone --depth 1 "https://github.com/$Repo.git" $tmp
    } else {
        Die 'git not found; cannot clone source'
    }
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Push-Location $tmp
    try {
        go build -ldflags "-s -w" -o (Join-Path $InstallDir "$Binary.exe") ./cmd/persona
    } finally {
        Pop-Location
    }
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

function Ensure-Path {
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath -notlike "*$InstallDir*") {
        $newPath = if ([string]::IsNullOrEmpty($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
        Warn "Added $InstallDir to your user PATH — open a new terminal to pick it up."
    }
    # also update the current session
    if ($env:Path -notlike "*$InstallDir*") { $env:Path = "$env:Path;$InstallDir" }
}

$arch = Get-Arch
Info "Target: windows_$arch  ·  install dir: $InstallDir"
if (-not (Install-Local)) {
    try {
        Install-FromRelease $arch
    } catch {
        Warn "No local source and no prebuilt binary ($($_.Exception.Message)); cloning from GitHub."
        Install-FromSource
    }
}
Ensure-Path
Info 'Verifying ...'
& (Join-Path $InstallDir "$Binary.exe") --version
Info "Done. Open a new terminal and run '$Binary' from anywhere."
