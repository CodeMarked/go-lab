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

# Cookie session: register -> login -> GET /users with jar -> logout -> 401
$webSess = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$ts = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$sessionEmail = "smoke+$ts@example.com"
$regBody = @{
    email    = $sessionEmail
    password = "smoke-pass-8ch"
    name     = "SmokeSession"
} | ConvertTo-Json
$regResp = Invoke-WebRequest -Uri "$ApiBaseUrl/api/v1/auth/register" -Method POST -Body $regBody -ContentType "application/json; charset=utf-8" -WebSession $webSess
Assert-True ($regResp.StatusCode -eq 201) "register should return 201"

$loginBody = @{
    email    = $sessionEmail
    password = "smoke-pass-8ch"
} | ConvertTo-Json
$loginResp = Invoke-WebRequest -Uri "$ApiBaseUrl/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json; charset=utf-8" -WebSession $webSess
Assert-True ($loginResp.StatusCode -eq 200) "login should return 200"
$cookieJar = $webSess.Cookies.GetCookies([Uri]$ApiBaseUrl)
$glCookie = $cookieJar | Where-Object { $_.Name -eq "gl_session" }
Assert-True ($null -ne $glCookie) "login should set gl_session cookie"
$csrfCookie = $cookieJar | Where-Object { $_.Name -eq "gl_csrf" }
Assert-True ($null -ne $csrfCookie) "login should set gl_csrf cookie"

$csrfReady = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/auth/csrf" -WebSession $webSess
Assert-True ($csrfReady.data.csrf_ready -eq $true) "GET /auth/csrf should succeed with session"

$usersViaCookie = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/users" -WebSession $webSess
Assert-True ($usersViaCookie.data.Count -ge 1) "GET /users with session cookie should succeed"

$jarForLogout = $webSess.Cookies.GetCookies([Uri]$ApiBaseUrl)
$csrfForLogout = ($jarForLogout | Where-Object { $_.Name -eq "gl_csrf" }).Value
Assert-True ($null -ne $csrfForLogout) "csrf cookie should exist before logout"
$csrfHdr = @{ "X-CSRF-Token" = $csrfForLogout }
Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/auth/logout" -Method POST -WebSession $webSess -Headers $csrfHdr | Out-Null

$cookieDead = $false
try {
    Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/users" -WebSession $webSess | Out-Null
}
catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 401) {
        $cookieDead = $true
    }
    else {
        throw
    }
}
Assert-True $cookieDead "after logout, GET /users with cookie should return 401"

# Bootstrap deprecation metadata (Origin required)
$bootResp = Invoke-RestMethod -Method POST -Uri "$ApiBaseUrl/api/v1/auth/bootstrap" -Headers @{ Origin = "http://localhost:4200" } -ContentType "application/json" -Body "{}"
Assert-True ($null -ne $bootResp.data.bootstrap) "bootstrap should include data.bootstrap deprecation object"
Assert-True ($bootResp.data.bootstrap.temporary -eq $true) "bootstrap.temporary should be true"

Write-Host "Smoke tests passed."
