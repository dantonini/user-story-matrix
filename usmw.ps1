# USM Wrapper (usmw.ps1) - A wrapper script for User Story Matrix CLI on Windows
# Similar to Maven Wrapper (mvnw) and Gradle Wrapper (gradlew)
# This script downloads and manages USM binary versions automatically

param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Arguments
)

# Configuration
$USM_HOME = if ($env:USM_HOME) { $env:USM_HOME } else { "$env:USERPROFILE\.usm" }
$USM_BIN_DIR = "$USM_HOME\bin"
$USM_CACHE_DIR = "$USM_HOME\cache"
$USM_VERSION_FILE = "$USM_HOME\version"
$GITHUB_REPO = "dantonini/user-story-matrix"
$GITHUB_API_URL = "https://api.github.com/repos/$GITHUB_REPO/releases/latest"
$GITHUB_RELEASES_URL = "https://github.com/$GITHUB_REPO/releases/download"

# Global variables
$Global:OS = ""
$Global:ARCH = ""
$Global:LATEST_VERSION = ""
$Global:CURRENT_VERSION = ""
$Global:USMW_DEBUG = $false

# Logging functions - only show output in debug mode (except errors)
function Write-Info {
    param([string]$Message)
    if ($Global:USMW_DEBUG) {
        Write-Host "[INFO] $Message" -ForegroundColor Blue
    }
}

function Write-Success {
    param([string]$Message)
    if ($Global:USMW_DEBUG) {
        Write-Host "[SUCCESS] $Message" -ForegroundColor Green
    }
}

function Write-Warning {
    param([string]$Message)
    if ($Global:USMW_DEBUG) {
        Write-Host "[WARN] $Message" -ForegroundColor Yellow
    }
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Detect OS and architecture
function Get-Platform {
    $Global:OS = "windows"
    
    $arch = [System.Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
    if (-not $arch) {
        $arch = (Get-WmiObject -Class Win32_Processor).Architecture
    }
    
    switch ($arch) {
        "AMD64" { $Global:ARCH = "amd64" }
        "x86_64" { $Global:ARCH = "amd64" }
        "ARM64" { $Global:ARCH = "arm64" }
        default {
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
    
    Write-Info "Detected platform: $Global:OS/$Global:ARCH"
}

# Get the latest version from GitHub API
function Get-LatestVersion {
    Write-Info "Checking latest version from GitHub..."
    
    try {
        $response = Invoke-RestMethod -Uri $GITHUB_API_URL -ErrorAction Stop
        $Global:LATEST_VERSION = $response.tag_name -replace '^v', ''
        Write-Info "Latest version: $Global:LATEST_VERSION"
    }
    catch {
        Write-Error "Failed to get latest version from GitHub API: $_"
        exit 1
    }
}

# Get the currently installed version
function Get-CurrentVersion {
    if (Test-Path $USM_VERSION_FILE) {
        $Global:CURRENT_VERSION = Get-Content $USM_VERSION_FILE -Raw | ForEach-Object { $_.Trim() }
        Write-Info "Current version: $Global:CURRENT_VERSION"
    }
    else {
        $Global:CURRENT_VERSION = ""
        Write-Info "No current version found"
    }
}

# Check if USM binary exists and is executable
function Test-BinaryExists {
    $binaryPath = "$USM_BIN_DIR\usm.exe"
    return (Test-Path $binaryPath)
}

# Download USM binary
function Get-UsmBinary {
    param([string]$Version)
    
    $downloadName = "usm-$Global:OS-$Global:ARCH-$Version.exe"
    $downloadUrl = "$GITHUB_RELEASES_URL/v$Version/$downloadName"
    $tempFile = "$USM_CACHE_DIR\$downloadName"
    $finalPath = "$USM_BIN_DIR\usm.exe"
    
    Write-Info "Downloading USM v$Version for $Global:OS/$Global:ARCH..."
    Write-Info "Download URL: $downloadUrl"
    
    # Create directories
    if (-not (Test-Path $USM_BIN_DIR)) {
        New-Item -ItemType Directory -Path $USM_BIN_DIR -Force | Out-Null
    }
    if (-not (Test-Path $USM_CACHE_DIR)) {
        New-Item -ItemType Directory -Path $USM_CACHE_DIR -Force | Out-Null
    }
    
    # Download the binary (suppress output unless debug mode)
    try {
        if ($Global:USMW_DEBUG) {
            Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -ErrorAction Stop
        } else {
            $ProgressPreference = 'SilentlyContinue'
            Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -ErrorAction Stop
            $ProgressPreference = 'Continue'
        }
    }
    catch {
        Write-Error "Failed to download USM binary from $downloadUrl`: $_"
        exit 1
    }
    
    # Move to final location
    Move-Item -Path $tempFile -Destination $finalPath -Force
    
    # Update version file
    Set-Content -Path $USM_VERSION_FILE -Value $Version
    
    Write-Success "USM v$Version installed successfully to $finalPath"
}

# Check if update is needed
function Update-UsmIfNeeded {
    Get-Platform
    Get-LatestVersion
    Get-CurrentVersion
    
    $needsDownload = $false
    
    # Check if binary exists
    if (-not (Test-BinaryExists)) {
        Write-Warning "USM binary not found, downloading..."
        $needsDownload = $true
    }
    # Check if version is outdated
    elseif ($Global:CURRENT_VERSION -ne $Global:LATEST_VERSION) {
        Write-Warning "USM version $Global:CURRENT_VERSION is outdated. Latest is $Global:LATEST_VERSION"
        $needsDownload = $true
    }
    else {
        Write-Success "USM v$Global:CURRENT_VERSION is up to date"
    }
    
    if ($needsDownload) {
        Get-UsmBinary -Version $Global:LATEST_VERSION
    }
}

# Show version information
function Show-Version {
    Get-Platform
    Get-LatestVersion
    Get-CurrentVersion
    
    Write-Host "USM Wrapper (usmw.ps1) Information:"
    Write-Host "  Platform: $Global:OS/$Global:ARCH"
    Write-Host "  Latest version: $Global:LATEST_VERSION"
    Write-Host "  Current version: $(if ($Global:CURRENT_VERSION) { $Global:CURRENT_VERSION } else { 'Not installed' })"
    Write-Host "  USM Home: $USM_HOME"
    Write-Host "  Binary location: $USM_BIN_DIR"
    
    if (Test-BinaryExists) {
        Write-Host "  Binary status: Installed ✓"
    }
    else {
        Write-Host "  Binary status: Not installed ✗"
    }
}

# Show help information
function Show-Help {
    Write-Host @"
USM Wrapper (usmw.ps1) - User Story Matrix CLI Wrapper for Windows

Usage: .\usmw.ps1 [wrapper-options] [usm-commands]
   or: powershell -ExecutionPolicy Bypass -File usmw.ps1 [wrapper-options] [usm-commands]

Wrapper Options:
  --usmw-version    Show wrapper and USM version information
  --usmw-help       Show this help message
  --usmw-update     Force update to the latest version
  --usmw-clean      Clean downloaded cache files
  --usmw-debug      Enable debug output (shows download progress)

Any other arguments are passed directly to the USM binary.

Examples:
  .\usmw.ps1 --help                    # Show USM help (transparent)
  .\usmw.ps1 --version                 # Show USM version
  .\usmw.ps1 add "new story"           # Add a new story
  .\usmw.ps1 --usmw-version            # Show wrapper version info
  .\usmw.ps1 --usmw-update             # Force update USM
  .\usmw.ps1 --usmw-debug --help       # Show USM help with wrapper debug info

The wrapper automatically downloads and updates USM when needed.
USM binary is stored in: $USM_HOME\bin\
"@
}

# Clean cache
function Clear-Cache {
    Write-Info "Cleaning USM cache..."
    if (Test-Path $USM_CACHE_DIR) {
        Remove-Item -Path $USM_CACHE_DIR -Recurse -Force
    }
    Write-Success "Cache cleaned"
}

# Main execution logic
function Main {
    param([string[]]$Args)
    
    # Check for debug flag first (can be anywhere in arguments)
    foreach ($arg in $Args) {
        if ($arg -eq "--usmw-debug") {
            $Global:USMW_DEBUG = $true
            break
        }
    }
    
    # Handle wrapper-specific options
    if ($Args.Count -gt 0) {
        switch ($Args[0]) {
            "--usmw-version" {
                Show-Version
                exit 0
            }
            "--usmw-help" {
                Show-Help
                exit 0
            }
            "--usmw-update" {
                Get-Platform
                Get-LatestVersion
                Get-UsmBinary -Version $Global:LATEST_VERSION
                exit 0
            }
            "--usmw-clean" {
                Clear-Cache
                exit 0
            }
            "--usmw-debug" {
                # If --usmw-debug is the only argument, show help
                if ($Args.Count -eq 1) {
                    Show-Help
                    exit 0
                }
                # Otherwise, remove --usmw-debug from arguments and continue
                $Args = $Args[1..($Args.Count-1)]
            }
        }
    }
    
    # Ensure USM is available and up to date
    Update-UsmIfNeeded
    
    # Execute USM with all arguments
    $usmBinary = "$USM_BIN_DIR\usm.exe"
    
    if (-not (Test-Path $usmBinary)) {
        Write-Error "USM binary not found at $usmBinary"
        exit 1
    }
    
    # Filter out --usmw-debug from arguments when passing to USM
    $filteredArgs = @()
    foreach ($arg in $Args) {
        if ($arg -ne "--usmw-debug") {
            $filteredArgs += $arg
        }
    }
    
    # Execute USM with filtered arguments
    if ($filteredArgs.Count -gt 0) {
        & $usmBinary @filteredArgs
    }
    else {
        & $usmBinary
    }
    
    # Pass through the exit code
    exit $LASTEXITCODE
}

# Run main function with all arguments
Main -Args $Arguments 