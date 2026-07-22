FROM python:3.12-slim-bookworm

ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 MODEL_CACHE=/models

RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY services/transcriber/requirements.txt ./requirements.txt
RUN pip install --no-cache-dir -r requirements.txt
COPY services/transcriber/app.py ./

RUN useradd --system --create-home --uid 10002 transcriber \
    && mkdir -p /models \
    && chown -R transcriber:transcriber /app /models
USER transcriber

EXPOSE 8091
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8091", "--no-access-log"]
