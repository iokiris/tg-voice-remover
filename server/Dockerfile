FROM python:3.10-slim

RUN apt-get update && apt-get install -y ffmpeg && apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
# кеширование зависимостей
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install -r requirements.txt
COPY . .

LABEL org.opencontainers.image.source https://github.com/iokiris/tg-voice-remover

CMD ["uvicorn", "process_audio:app", "--host", "0.0.0.0", "--port", "8000", "--reload"]
