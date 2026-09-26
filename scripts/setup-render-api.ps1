# Automate Upstash Redis + Render env updates (after Blueprint exists).
# Requires: UPSTASH_EMAIL, UPSTASH_API_KEY, RENDER_API_KEY, DISCORD_CLIENT_SECRET
# Optional: RENDER_SERVICE_ID (srv-...), UPSTASH_DB_NAME (default quietcord-api)

param(
    [string]$ServiceName = "quietcord-api",
    [string]$UpstashRegion = "us-west-1"
)

$ErrorActionPreference = "Stop"

function Require-Env($name) {
    $v = [Environment]::GetEnvironmentVariable($name)
    if (-not $v) { $v = (Get-Item "Env:$name" -ErrorAction SilentlyContinue).Value }
    if (-not $v) { throw "Missing env var $name" }
    return $v
}

$upstashEmail = Require-Env "UPSTASH_EMAIL"
$upstashKey = Require-Env "UPSTASH_API_KEY"
$renderKey = Require-Env "RENDER_API_KEY"
$discordSecret = Require-Env "DISCORD_CLIENT_SECRET"

$dbName = if ($env:UPSTASH_DB_NAME) { $env:UPSTASH_DB_NAME } else { "quietcord-api" }
$auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("${upstashEmail}:${upstashKey}"))

Write-Host "Creating Upstash Redis database '$dbName'..."
$createBody = @{
    database_name = $dbName
    region = $UpstashRegion
    tls = $true
} | ConvertTo-Json

try {
    $db = Invoke-RestMethod -Method POST `
        -Uri "https://api.upstash.com/v2/redis/database" `
        -Headers @{ Authorization = "Basic $auth"; "Content-Type" = "application/json" } `
        -Body $createBody
} catch {
    if ($_.Exception.Response.StatusCode -eq 409) {
        Write-Host "Database may already exist; listing..."
        $dbs = Invoke-RestMethod -Method GET `
            -Uri "https://api.upstash.com/v2/redis/databases" `
            -Headers @{ Authorization = "Basic $auth" }
        $db = $dbs | Where-Object { $_.database_name -eq $dbName } | Select-Object -First 1
        if (-not $db) { throw $_ }
    } else { throw }
}

$hostPart = $db.endpoint -replace "^rediss?://", ""
$pass = $db.password
$port = if ($db.port) { $db.port } else { 6379 }
$redisUri = "rediss://default:${pass}@${hostPart}:${port}"
Write-Host "REDIS_URI=$redisUri"

Write-Host "Resolving Render service..."
$services = Invoke-RestMethod -Method GET `
    -Uri "https://api.render.com/v1/services?limit=100" `
    -Headers @{ Authorization = "Bearer $renderKey"; Accept = "application/json" }

$serviceId = $env:RENDER_SERVICE_ID
if (-not $serviceId) {
    $match = $services | ForEach-Object { $_.service } | Where-Object { $_.name -eq $ServiceName } | Select-Object -First 1
    if (-not $match) {
        throw "Render service '$ServiceName' not found. Create Blueprint first, then re-run."
    }
    $serviceId = $match.id
}

$envPayload = @(
    @{ key = "REDIS_URI"; value = $redisUri }
    @{ key = "DISCORD_CLIENT_SECRET"; value = $discordSecret }
)

foreach ($ev in $envPayload) {
    Write-Host "Setting $($ev.key) on $serviceId..."
    Invoke-RestMethod -Method PUT `
        -Uri "https://api.render.com/v1/services/$serviceId/env-vars/$($ev.key)" `
        -Headers @{ Authorization = "Bearer $renderKey"; "Content-Type" = "application/json" } `
        -Body (@{ value = $ev.value } | ConvertTo-Json) | Out-Null
}

Write-Host "Triggering deploy..."
Invoke-RestMethod -Method POST `
    -Uri "https://api.render.com/v1/services/$serviceId/deploys" `
    -Headers @{ Authorization = "Bearer $renderKey"; "Content-Type" = "application/json" } `
    -Body '{"clearCache":false}' | Out-Null

Write-Host "Done. Check https://$ServiceName.onrender.com/v1/"
