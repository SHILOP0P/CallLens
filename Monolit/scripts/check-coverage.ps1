param(
    [double]$Minimum = 80,
    [string]$GoCommand = "go"
)

$ErrorActionPreference = "Stop"

$module = (& $GoCommand list -m).Trim()
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

# These packages contain no executable business logic or are generated/test-only code.
$excludedPackages = @(
    "$module/cmd",
    "$module/internal/API",
    "$module/internal/API/dto",
    "$module/internal/models",
    "$module/internal/repository",
    "$module/internal/repository/models",
    "$module/internal/repository/repositorytest",
    "$module/internal/service",
    "$module/internal/storage"
)

$packages = & $GoCommand list ./... |
    Where-Object {
        $_ -notin $excludedPackages -and
        $_ -notmatch '/mocks($|/)'
    }

if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

$output = & $GoCommand test $packages -cover 2>&1
$testExitCode = $LASTEXITCODE
$output | ForEach-Object { Write-Host $_ }

if ($testExitCode -ne 0) {
    exit $testExitCode
}

$coverageByPackage = @{}
foreach ($line in $output) {
    if ($line -match '^\s*(?:ok\s+)?(?<package>\S+).*coverage:\s+(?<coverage>\d+(?:\.\d+)?)%') {
        $coverageByPackage[$Matches.package] = [double]$Matches.coverage
    }
}

$missing = $packages | Where-Object { -not $coverageByPackage.ContainsKey($_) }
$below = $coverageByPackage.GetEnumerator() |
    Where-Object { $_.Value -lt $Minimum } |
    Sort-Object Name

if ($missing.Count -gt 0) {
    Write-Error "Coverage result is missing for: $($missing -join ', ')"
    exit 1
}

if ($below.Count -gt 0) {
    foreach ($package in $below) {
        Write-Error ("{0}: {1:N1}% is below {2:N1}%" -f $package.Name, $package.Value, $Minimum)
    }
    exit 1
}

Write-Host ("Coverage gate passed: all {0} executable packages are at least {1:N1}%." -f $packages.Count, $Minimum)
