$composeArgs = @("compose", "-f", "deploy/docker-compose.yaml", "--project-directory", ".")

$envFile = Join-Path $PSScriptRoot "..\.env"
$usesHybridDiarization = (Test-Path $envFile -PathType Leaf) -and (
    Select-String -Path $envFile -Pattern '^\s*TRANSCRIBER_PROVIDER\s*=\s*(hybrid|openrouter-pyannote|local-pyannote)\s*$' -Encoding UTF8 -Quiet
)
if ($usesHybridDiarization) {
    $composeArgs += @("--profile", "diarization")
}
$usesLocalTranscription = (Test-Path $envFile -PathType Leaf) -and (
    Select-String -Path $envFile -Pattern '^\s*TRANSCRIBER_PROVIDER\s*=\s*(local|local-pyannote)\s*$' -Encoding UTF8 -Quiet
)
if ($usesLocalTranscription) {
    $composeArgs += @("--profile", "local-transcription")
}

& docker @composeArgs down
exit $LASTEXITCODE
