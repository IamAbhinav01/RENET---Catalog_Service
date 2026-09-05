Write-Host "Testing GET /movies/"

$response = curl.exe -s -w "`n%{http_code}" http://localhost:3000/movies/

$parts = $response -split "`n"
$body = $parts[0]
$status = $parts[1]

Write-Host "Response:"
Write-Host $body

if ($status -eq "200") {
    Write-Host "PASS - Status: $status"
}
else {
    Write-Host "FAIL - Status: $status"
}