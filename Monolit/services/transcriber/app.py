from __future__ import annotations

import os
import shutil
import tempfile
from contextlib import asynccontextmanager
from pathlib import Path

import torch
from fastapi import FastAPI, File, HTTPException, UploadFile
from faster_whisper import WhisperModel
from pyannote.audio import Pipeline

from alignment import SpeakerTurn, TimedText, assign_speakers, dialogue_text


class Models:
    whisper: WhisperModel
    diarizer: Pipeline


models = Models()


@asynccontextmanager
async def lifespan(_: FastAPI):
    token = os.environ.get("HF_TOKEN", "").strip()
    if not token:
        raise RuntimeError("HF_TOKEN is required for pyannote Community-1")

    device = os.environ.get("WHISPER_DEVICE", "cpu")
    compute_type = os.environ.get("WHISPER_COMPUTE_TYPE", "int8")
    models.whisper = WhisperModel(
        os.environ.get("WHISPER_MODEL", "large-v3"),
        device=device,
        compute_type=compute_type,
        download_root=os.environ.get("MODEL_CACHE", "/models/whisper"),
    )
    pipeline = Pipeline.from_pretrained(
        os.environ.get("PYANNOTE_MODEL", "pyannote/speaker-diarization-community-1"),
        token=token,
        cache_dir=os.environ.get("MODEL_CACHE", "/models"),
    )
    if pipeline is None:
        raise RuntimeError("pyannote pipeline could not be loaded; accept the model conditions and verify HF_TOKEN")
    if device.startswith("cuda"):
        pipeline.to(torch.device(device))
    models.diarizer = pipeline
    yield


app = FastAPI(title="CallLens local transcription", version="1.0.0", lifespan=lifespan)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/v1/transcribe")
def transcribe(file: UploadFile = File(...)) -> dict[str, object]:
    suffix = Path(file.filename or "call-media").suffix[:12]
    with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as target:
        source_path = target.name
        shutil.copyfileobj(file.file, target)

    try:
        whisper_segments, info = models.whisper.transcribe(
            source_path,
            language=os.environ.get("TRANSCRIPTION_LANGUAGE", "ru"),
            beam_size=5,
            vad_filter=True,
            word_timestamps=False,
        )
        transcript = [
            TimedText(float(segment.start), float(segment.end), segment.text)
            for segment in whisper_segments
            if segment.text.strip()
        ]

        diarization = models.diarizer(source_path)
        annotation = getattr(diarization, "exclusive_speaker_diarization", None)
        if annotation is None:
            annotation = diarization.speaker_diarization
        turns = [
            SpeakerTurn(float(turn.start), float(turn.end), str(speaker))
            for turn, speaker in annotation
        ]
        segments = assign_speakers(transcript, turns)
        text = dialogue_text(segments)
        if not text:
            raise HTTPException(status_code=422, detail="speech or speaker turns were not detected")
        return {"text": text, "language": info.language, "segments": segments}
    except HTTPException:
        raise
    except Exception as error:
        raise HTTPException(status_code=500, detail=f"transcription failed: {error}") from error
    finally:
        Path(source_path).unlink(missing_ok=True)
