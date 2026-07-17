from __future__ import annotations

import os
import shutil
import tempfile
from contextlib import asynccontextmanager
from pathlib import Path

import torch
from fastapi import FastAPI, File, HTTPException, UploadFile
from pyannote.audio import Pipeline


class Models:
    diarizer: Pipeline


models = Models()


@asynccontextmanager
async def lifespan(_: FastAPI):
    token = os.environ.get("HF_TOKEN", "").strip()
    if not token:
        raise RuntimeError("HF_TOKEN is required for pyannote Community-1")
    pipeline = Pipeline.from_pretrained(
        os.environ.get("PYANNOTE_MODEL", "pyannote/speaker-diarization-community-1"),
        token=token,
        cache_dir=os.environ.get("MODEL_CACHE", "/models"),
    )
    if pipeline is None:
        raise RuntimeError("pyannote pipeline could not be loaded; accept the model conditions and verify HF_TOKEN")
    device = os.environ.get("PYANNOTE_DEVICE", "cpu")
    if device.startswith("cuda"):
        pipeline.to(torch.device(device))
    models.diarizer = pipeline
    yield


app = FastAPI(title="CallLens speaker diarization", version="1.0.0", lifespan=lifespan)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/v1/diarize")
def diarize(file: UploadFile = File(...)) -> dict[str, object]:
    suffix = Path(file.filename or "call-media").suffix[:12]
    with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as target:
        source_path = target.name
        shutil.copyfileobj(file.file, target)
    try:
        diarization = models.diarizer(source_path)
        annotation = getattr(diarization, "exclusive_speaker_diarization", None)
        if annotation is None:
            annotation = diarization.speaker_diarization
        turns = [
            {"start_seconds": round(float(turn.start), 3), "end_seconds": round(float(turn.end), 3), "speaker": str(speaker)}
            for turn, speaker in annotation
            if float(turn.end) > float(turn.start)
        ]
        if not turns:
            raise HTTPException(status_code=422, detail="speaker turns were not detected")
        return {"turns": turns}
    except HTTPException:
        raise
    except Exception as error:
        raise HTTPException(status_code=500, detail=f"diarization failed: {error}") from error
    finally:
        Path(source_path).unlink(missing_ok=True)
