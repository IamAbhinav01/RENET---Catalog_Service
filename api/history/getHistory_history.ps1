
Write-Host "========================================"
Write-Host "       ReNet API - History Test"
Write-Host "========================================"
Write-Host ""

# ========================================
# TEST 1: GET /history
# ========================================

Write-Host "[TEST 1] GET /history"
Write-Host ""

$response = curl.exe -s `
    -w "|||%{http_code}" `
    -X GET `
    "http://localhost:3000/history" `
    -H "Authorization: Bearer YOUR_TOKEN_HERE"

$parts = $response -split '\|\|\|'
$body = $parts[0]
$status = $parts[1]

Write-Host "Status:"
Write-Host "$status"
Write-Host ""

Write-Host "Response:"
try {
    $body | ConvertFrom-Json | ConvertTo-Json -Depth 10
}
catch {
    Write-Host $body
}

Write-Host ""

if ($status -eq "200") {
    Write-Host "PASS - User history retrieved successfully"
}
elseif ($status -eq "401") {
    Write-Host "FAIL - User is not authenticated"
}
else {
    Write-Host "FAIL - Unexpected status: $status"
}

Write-Host ""
Write-Host "========================================"
Write-Host ""

# ========================================
# TEST 2: GET /history?limit=10
# ========================================

Write-Host "[TEST 2] GET /history?limit=10"
Write-Host ""

$response = curl.exe -s `
    -w "|||%{http_code}" `
    -X GET `
    "http://localhost:3000/history?limit=10" `
   
$parts = $response -split '\|\|\|'
$body = $parts[0]
$status = $parts[1]

Write-Host "Status:"
Write-Host "$status"
Write-Host ""

Write-Host "Response:"
try {
    $body | ConvertFrom-Json | ConvertTo-Json -Depth 10
}
catch {
    Write-Host $body
}

Write-Host ""

if ($status -eq "200") {
    Write-Host "PASS - History retrieved with limit=10"
}
elseif ($status -eq "401") {
    Write-Host "FAIL - User is not authenticated"
}
elseif ($status -eq "400") {
    Write-Host "FAIL - Invalid limit"
}
else {
    Write-Host "FAIL - Unexpected status: $status"
}

Write-Host ""
Write-Host "========================================"
