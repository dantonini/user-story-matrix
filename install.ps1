# USM-CLI Installation Script for Windows
# This script downloads and installs the latest version of USM-CLI

param (
    [string]$InstallDir = "$env:USERPROFILE\usm-cli",
    [string]$Version = "0.1.0"
)

# Create installation directory if it doesn't exist
if (-not (Test-Path $InstallDir)) {
    Write-Host "Creating installation directory: $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Construct download URL
$DownloadUrl = "https://github.com/user-story-matrix/usm-cli/releases/download/v$Version/usm-windows-amd64-$Version.exe"
$OutputFile = "$InstallDir\usm.exe"

Write-Host "Installing USM-CLI v$Version for Windows..."
Write-Host "Download URL: $DownloadUrl"

# Download binary
Write-Host "Downloading..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $OutputFile
}
catch {
    Write-Host "Error downloading USM-CLI: $_"
    exit 1
}

# Add to PATH if not already there
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $CurrentPath.Contains($InstallDir)) {
    Write-Host "Adding $InstallDir to PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "USM-CLI installed successfully to $OutputFile"
Write-Host "Run 'usm --help' to get started" 