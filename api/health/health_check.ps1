Write-Host "========================================"
Write-Host "     ReNet API - Health Check Test"
Write-Host "========================================"
Write-Host ""

$response = curl.exe -s -w "|||%{http_code}" http://localhost:3000/health
$parts = $response -split '\|\|\|'
$body = $parts[0]
$status = $parts[1]

Write-Host "GET /health -> Status: $status, Response: $body"

$responseApi = curl.exe -s -w "|||%{http_code}" http://localhost:3000/api/health
$partsApi = $responseApi -split '\|\|\|'
$bodyApi = $partsApi[0]
$statusApi = $partsApi[1]

Write-Host "GET /api/health -> Status: $statusApi, Response: $bodyApi"

if ($status -eq "200" -and $statusApi -eq "200") {
    Write-Host "PASS - All health probes OK"
} else {
    Write-Host "FAIL - Health probe status error"
    exit 1
}
