
Write-Host "========================================"
Write-Host "       ReNet API - History Test"
Write-Host "========================================"
Write-Host ""

Write-Host "[TEST] POST /history"
Write-Host ""

$response = curl.exe -s `
    -w "|||%{http_code}" `
    -X POST `
    "http://localhost:3000/history" `
    -H "Content-Type: application/json" `
    -d '{"item_id":3,"rating":5,"event_type":"watched"}'

# Separate response body and HTTP status
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

if ($status -eq "201") {
    Write-Host "PASS - Interaction recorded successfully"
}
elseif ($status -eq "401") {
    Write-Host "FAIL - User is not authenticated"
}
elseif ($status -eq "400") {
    Write-Host "FAIL - Invalid request body"
}
else {
    Write-Host "FAIL - Unexpected status: $status"
    exit 1
}

