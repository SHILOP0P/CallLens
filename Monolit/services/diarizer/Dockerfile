FROM python:3.12-slim-bookworm

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    MODEL_CACHE=/models

RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY services/diarizer/requirements.txt ./requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY services/diarizer/app.py ./

RUN useradd --system --create-home --uid 10001 diarizer \
    && mkdir -p /models \
    && chown -R diarizer:diarizer /app /models
USER diarizer

EXPOSE 8090
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8090"]
