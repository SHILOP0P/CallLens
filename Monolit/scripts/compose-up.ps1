$composeArgs = @("compose", "-f", "deploy/docker-compose.yaml", "--project-directory", ".")

$envFile = Join-Path $PSScriptRoot "..\.env"
$usesHybridDiarization = (Test-Path $envFile -PathType Leaf) -and (
    Select-String -Path $envFile -Pattern '^\s*TRANSCRIBER_PROVIDER\s*=\s*(hybrid|openrouter-pyannote)\s*$' -Encoding UTF8 -Quiet
)
if ($usesHybridDiarization) {
    $composeArgs += @("--profile", "diarization")
}

& docker @composeArgs up --build --wait
if ($LASTEXITCODE -eq 0) {
    exit 0
}

$exitCode = $LASTEXITCODE
Write-Warning "Docker Compose startup failed; stopping partially started services."
& docker @composeArgs down
exit $exitCode
