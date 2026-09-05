Write-Host "========================================"
Write-Host "     ReNet API - Batch Movies Test"
Write-Host "========================================"
Write-Host ""

$payloadFile = Join-Path $env:TEMP "renet-batch-payload.json"
Set-Content -Path $payloadFile -Value '{"ids": [1, 2, 364]}' -NoNewline

Write-Host "[TEST 1] POST /api/movies/batch"
$response = curl.exe -s `
    -w "|||%{http_code}" `
    -X POST `
    "http://localhost:3000/api/movies/batch" `
    -H "Content-Type: application/json" `
    --data-binary "@$payloadFile"

$parts = $response -split '\|\|\|'
$body = $parts[0]
$status = $parts[1]

Write-Host "Status: $status"
Write-Host "Response:"
try {
    $body | ConvertFrom-Json | ConvertTo-Json -Depth 5
}
catch {
    Write-Host $body
}

if ($status -eq "200") {
    Write-Host "PASS - Batch movies retrieved successfully"
} else {
    Write-Host "FAIL - Status: $status"
    exit 1
}

Write-Host ""
Write-Host "[TEST 2] POST /movies/batch (root alias)"
$response2 = curl.exe -s `
    -w "|||%{http_code}" `
    -X POST `
    "http://localhost:3000/movies/batch" `
    -H "Content-Type: application/json" `
    --data-binary "@$payloadFile"

Remove-Item $payloadFile -ErrorAction SilentlyContinue

$parts2 = $response2 -split '\|\|\|'
$status2 = $parts2[1]
Write-Host "Status: $status2"

if ($status2 -eq "200") {
    Write-Host "PASS - Root alias /movies/batch retrieved successfully"
} else {
    Write-Host "FAIL - Status: $status2"
    exit 1
}
