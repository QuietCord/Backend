# Run from an interactive PowerShell after: fly auth login
# Usage: .\scripts\deploy-fly.ps1 [-AppName quietcord-api] [-DiscordClientSecret "..."]

param(
    [string]$AppName = "quietcord-api",
    [string]$DiscordClientId = "1553407390182019282",
    [string]$DiscordClientSecret = "",
    [string]$Region = "dfw"
)

$ErrorActionPreference = "Stop"
$fly = Join-Path $env:USERPROFILE ".fly\bin\flyctl.exe"
if (-not (Test-Path $fly)) { throw "flyctl not found. Run: iwr https://fly.io/install.ps1 -useb | iex" }

& $fly auth whoami | Out-Null

function New-Pepper {
    -join ((1..64) | ForEach-Object { "{0:x2}" -f (Get-Random -Maximum 256) })
}

if (-not $DiscordClientSecret) {
    $DiscordClientSecret = Read-Host "Discord OAuth Client Secret (portal → OAuth2)"
}

$pepperSecrets = New-Pepper
$pepperSettings = New-Pepper
$base = "https://$AppName.fly.dev"

Write-Host "Creating app $AppName (skip if exists)..."
& $fly apps create $AppName 2>$null

Write-Host "Creating Redis in $Region..."
& $fly redis create --name "$AppName-redis" --region $Region --no-replicas --enable-eviction --org personal 2>&1

Write-Host "Paste the Private URL from above (rediss://...):"
$redisUri = Read-Host "REDIS_URI"

Write-Host "Setting secrets..."
& $fly secrets set `
    "REDIS_URI=$redisUri" `
    "DISCORD_CLIENT_ID=$DiscordClientId" `
    "DISCORD_CLIENT_SECRET=$DiscordClientSecret" `
    "DISCORD_REDIRECT_URI=$base/v1/oauth/callback" `
    "ROOT_REDIRECT=https://github.com/QuietCord/Backend" `
    "PEPPER_SECRETS=$pepperSecrets" `
    "PEPPER_SETTINGS=$pepperSettings" `
    -a $AppName

Write-Host "Deploying..."
Push-Location (Split-Path $PSScriptRoot -Parent)
& $fly deploy -a $AppName
Pop-Location

Write-Host "Done. API: $base/"
Write-Host "Add OAuth redirect in Discord: $base/v1/oauth/callback"
Write-Host "Set CLOUD_API_URL in Quiet brand.ts to $base/"
