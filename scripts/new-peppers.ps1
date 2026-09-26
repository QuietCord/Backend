# One-off peppers for Render / Fly / .env (do not commit output).
function New-Pepper {
    -join ((1..64) | ForEach-Object { "{0:x2}" -f (Get-Random -Maximum 256) })
}
Write-Host "PEPPER_SECRETS=$(New-Pepper)"
Write-Host "PEPPER_SETTINGS=$(New-Pepper)"
