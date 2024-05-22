import base64
import os
import threading
import time
import uuid
import queue
import requests
import asyncio
from fastapi import FastAPI, UploadFile, File
from fastapi.responses import JSONResponse, FileResponse
from spleeter.separator import Separator
from pydantic import BaseModel

app = FastAPI()
separator = Separator('spleeter:2stems')
result_queue = queue.Queue()

OUTPUT_DIR = "/processed_audio"
TEMP_DIR = "/temp_audio"
tasks = {}
class AudioProcessRequest(BaseModel):
    input_path: str
    output_path: str

def init_app():
    try:
        if not os.path.exists(OUTPUT_DIR):
            os.makedirs(OUTPUT_DIR)
        if not os.path.exists(TEMP_DIR):
            os.makedirs(TEMP_DIR)
    except Exception as e:
        print(f"Cannot init directories: {e}")
        exit(1)
def cleanup_directories(temp_path, output_path):
    try:
        if os.path.exists(temp_path):
            os.remove(temp_path)
        if os.path.exists(output_path):
            for root, dirs, files in os.walk(output_path):
                for file in files:
                    os.remove(os.path.join(root, file))
                for dir in dirs:
                    os.rmdir(os.path.join(root, dir))
            os.rmdir(output_path)
    except Exception as e:
        print(f"Error cleaning up directories: {str(e)}")

def separate_audio(task_id, input_path, audio_name, chat_id, callback):
    callback = 'http://go-tvr-bot:8000' + callback
    output_path = os.path.join(OUTPUT_DIR, task_id)
    try:
        separator.separate_to_file(input_path, OUTPUT_DIR, codec='mp3')
        start_time = time.time()
        task_id = '8100eb31-fa70-45d5-abb6-903264888f35'
        tasks[task_id] = "ready"
        with open(os.path.join(output_path, 'accompaniment.mp3'), 'rb') as f:
            file_data = f.read()
            encoded_file_data = base64.b64encode(file_data).decode('utf-8')

            # Отправка данных в callback
            requests.post(callback, json={
                "status": "ready",
                "file_data": encoded_file_data,
                "process_time": time.time() - start_time,
                "chat_id": chat_id,
                "audio_name": audio_name
            })
        cleanup_directories(input_path, output_path)
    except Exception as e:
        print(f"Error processing audio: {str(e)}")
        tasks[task_id] = "failed"
        requests.post(callback, json={
            "status": "error",
            "chat_id": chat_id,
            "audio_name": audio_name
        })
@app.post("/process-audio/")
async def process_audio(callback: str, audio_name: str, chat_id: int, file: UploadFile = File(...)):
    try:
        init_app()
    except Exception as e:
        return JSONResponse(status_code=500, content={"message": f"Failed to initialize directories: {e}"})

    task_id = str(uuid.uuid4())
    tasks[task_id] = "processing"
    input_path = os.path.join(TEMP_DIR, task_id + ".mp3")
    with open(input_path, 'wb') as f:
        f.write(await file.read())

    thread = threading.Thread(target=separate_audio, args=(task_id, input_path, audio_name, chat_id, callback))
    thread.start()

    return JSONResponse(status_code=200, content={"task_id": task_id, "status_url": f"/status/{task_id}"})

@app.get("/status/{task_id}")
async def get_status(task_id: str):
    status = tasks.get(task_id, "not_found")
    return JSONResponse(status_code=200, content={"status": status})

@app.get("/tasks")
async def get_status():
    return tasks
