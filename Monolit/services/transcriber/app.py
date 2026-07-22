import os
import shutil
import tempfile
from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import FastAPI, File, HTTPException, UploadFile
from faster_whisper import WhisperModel


class Models:
    transcriber: WhisperModel


models = Models()


@asynccontextmanager
async def lifespan(_: FastAPI):
    models.transcriber = WhisperModel(
        os.environ.get("WHISPER_MODEL", "large-v3"),
        device=os.environ.get("WHISPER_DEVICE", "cpu"),
        compute_type=os.environ.get("WHISPER_COMPUTE_TYPE", "int8"),
        download_root=os.environ.get("MODEL_CACHE", "/models"),
    )
    yield


app = FastAPI(title="CallLens local transcription", version="1.0.0", lifespan=lifespan)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/v1/audio/transcriptions")
def transcribe(file: UploadFile = File(...)) -> dict[str, object]:
    suffix = Path(file.filename or "call-media").suffix[:12]
    with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as target:
        shutil.copyfileobj(file.file, target)
        source_path = target.name
    try:
        segments, info = models.transcriber.transcribe(
            source_path,
            language="ru",
            vad_filter=True,
            word_timestamps=False,
        )
        result = []
        text = []
        for segment in segments:
            value = segment.text.strip()
            if not value:
                continue
            text.append(value)
            result.append({
                "start_seconds": round(float(segment.start), 3),
                "end_seconds": round(float(segment.end), 3),
                "text": value,
            })
        joined = " ".join(text).strip()
        if not joined:
            raise ValueError("transcription is empty")
        return {"text": joined, "language": info.language or "ru", "segments": result}
    except Exception as error:
        raise HTTPException(status_code=500, detail=f"transcription failed: {error}") from error
    finally:
        Path(source_path).unlink(missing_ok=True)
