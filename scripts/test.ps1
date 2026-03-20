param(
    [string]$ApiBaseUrl = "http://localhost:5000",
    [string]$ClientId = "dev-platform",
    [string]$ClientSecret = "change-me-dev-secret-min-length-16"
)

$ErrorActionPreference = "Stop"

function Assert-True {
    param(
        [bool]$Condition,
        [string]$Message
    )
    if (-not $Condition) {
        throw "ASSERT FAILED: $Message"
    }
}

Write-Host "Running smoke tests against $ApiBaseUrl"

$health = Invoke-RestMethod -Method GET -Uri "$ApiBaseUrl/healthz"
Assert-True ($health.status -eq "ok") "healthz should return status=ok"

$ready = Invoke-RestMethod -Method GET -Uri "$ApiBaseUrl/readyz"
Assert-True ($ready.status -eq "ready") "readyz should return status=ready"

# Unversioned /api should be gone
$legacyGone = $false
try {
    Invoke-RestMethod -Method GET -Uri "$ApiBaseUrl/api/users" | Out-Null
}
catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 410) {
        $legacyGone = $true
    }
    else {
        throw
    }
}
Assert-True $legacyGone "legacy /api/users should return 410 Gone"

# v1 users without token -> 401
$unauth = $false
try {
    Invoke-RestMethod -Method GET -Uri "$ApiBaseUrl/api/v1/users" | Out-Null
}
catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 401) {
        $unauth = $true
    }
    else {
        throw
    }
}
Assert-True $unauth "GET /api/v1/users without bearer should return 401"

$tokenBody = @{
    grant_type    = "client_credentials"
    client_id     = $ClientId
    client_secret = $ClientSecret
} | ConvertTo-Json

$tokenResp = Invoke-RestMethod -Method POST -Uri "$ApiBaseUrl/api/v1/auth/token" -ContentType "application/json" -Body $tokenBody
Assert-True ($null -ne $tokenResp.data) "token response should include data envelope"
Assert-True ($tokenResp.data.access_token.Length -gt 10) "access_token should be present"
Assert-True ($tokenResp.data.token_type -eq "Bearer") "token_type should be Bearer"

$access = $tokenResp.data.access_token
$authHeaders = @{
    Authorization = "Bearer $access"
}

$newUser = @{
    name    = "SmokeAlice"
    pennies = 77
} | ConvertTo-Json

$created = Invoke-RestMethod -Method POST -Uri "$ApiBaseUrl/api/v1/users" -ContentType "application/json" -Body $newUser -Headers $authHeaders
Assert-True ($null -ne $created.data) "create user should use envelope"
Assert-True ($created.data.id -gt 0) "created user should include numeric id"
Assert-True ($created.data.name -eq "SmokeAlice") "created user name mismatch"

$uid = $created.data.id

$users = Invoke-RestMethod -Method GET -Uri "$ApiBaseUrl/api/v1/users" -Headers $authHeaders
Assert-True ($users.data.Count -ge 1) "users list should have at least one user after create"

$search = Invoke-RestMethod -Method GET -Uri "$ApiBaseUrl/api/v1/users/search?name=SmokeAli" -Headers $authHeaders
Assert-True ($search.data.Count -ge 1) "search should return created user"

$update = @{
    name    = "SmokeAliceUpdated"
    pennies = 99
} | ConvertTo-Json

$updated = Invoke-RestMethod -Method PUT -Uri "$ApiBaseUrl/api/v1/users/$uid" -ContentType "application/json" -Body $update -Headers $authHeaders
Assert-True ($updated.data.name -eq "SmokeAliceUpdated") "update should change user name"
Assert-True ($updated.data.pennies -eq 99) "update should change pennies"

$delResp = Invoke-WebRequest -Uri "$ApiBaseUrl/api/v1/users/$uid" -Method DELETE -Headers $authHeaders
Assert-True ($delResp.StatusCode -eq 204) "DELETE should return 204 No Content"

$notFound = $false
try {
    Invoke-RestMethod -Method GET -Uri "$ApiBaseUrl/api/v1/users/$uid" -Headers $authHeaders | Out-Null
}
catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 404) {
        $notFound = $true
    }
    else {
        throw
    }
}
Assert-True $notFound "deleted user should return 404 on fetch"

Write-Host "Smoke tests passed."
