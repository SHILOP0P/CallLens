$composeArgs = @("compose", "-f", "deploy/docker-compose.yaml", "--project-directory", ".")

& docker @composeArgs up --build --wait
if ($LASTEXITCODE -eq 0) {
    exit 0
}

$exitCode = $LASTEXITCODE
Write-Warning "Docker Compose startup failed; stopping partially started services."
& docker @composeArgs down
exit $exitCode
