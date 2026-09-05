Write-Host "Testing GET /movies/search?q=toy story"

$response = curl.exe -s -w "`n%{http_code}" "http://localhost:3000/movies/search?q=toy%20story%203%20%282010%29"

$parts = $response -split "`n"
$body = $parts[0]
$status = $parts[-1]

Write-Host "Response:"
Write-Host $body

if ($status -eq "200") {
    Write-Host "PASS - Status: $status"
}
else {
    Write-Host "FAIL - Status: $status"
}