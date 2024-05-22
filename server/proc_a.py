import os
import threading
import uuid
import queue
import asyncio
from fastapi import FastAPI, UploadFile, File, BackgroundTasks
from fastapi.responses import JSONResponse, FileResponse
from spleeter.separator import Separator
from pydantic import BaseModel

app = FastAPI()
separator = Separator('spleeter:2stems')
result_queue = queue.Queue()

OUTPUT_DIR = "/processed_audio"
TEMP_DIR = "/temp_audio"

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

def separate_audio(input_path, dirname):
    try:
        output_path = os.path.join(OUTPUT_DIR, dirname)
        if not os.path.exists(output_path):
            os.makedirs(output_path)
        separator.separate_to_file(input_path, OUTPUT_DIR, codec='mp3')
        result_queue.put((True, dirname))
    except Exception as e:
        print(f"Error processing audio: {str(e)}")
        result_queue.put((False, None))

@app.post("/process-audio/")
async def process_audio(background_tasks: BackgroundTasks, file: UploadFile = File(...)):
    try:
        init_app()
    except Exception as e:
        return JSONResponse(status_code=500, content={"message": f"Failed to initialize directories: {e}"})

    if file is None:
        return JSONResponse(status_code=400, content={"message": "no file received"})

    file_uuid = str(uuid.uuid4())
    input_path = os.path.join(TEMP_DIR, file_uuid + ".mp3")
    with open(input_path, 'wb') as f:
        f.write(await file.read())

    #
    # asyncio.create_task(separate_audio(input_path, file_uuid))
    # thread = threading.Thread(target=separate_audio, args=(input_path, file_uuid))
    # thread.start()
    # thread.join()
    background_tasks.add_task(separate_audio, input_path, file_uuid)


    return JSONResponse(status_code=200, content={
        "message": "Audio is being processed",
        "filename": file_uuid,
        "download_url": f"/download/{file_uuid}"
    })

@app.get("/download/{file_name}")
async def download_file(file_name: str):
    file_path = os.path.join(OUTPUT_DIR, file_name, 'accompaniment.mp3')
    if not os.path.exists(file_path):
        return JSONResponse(status_code=404, content={"message": "File not found"})
    response = FileResponse(file_path)

    @response.background
    def cleanup():
        temp_path = os.path.join(TEMP_DIR, file_name + ".mp3")
        output_path = os.path.join(OUTPUT_DIR, file_name)
        cleanup_directories(temp_path, output_path)

    return response
@app.get("/get_path/{file_name}")
async def download_file(file_name: str):
    file_path = os.path.join(OUTPUT_DIR, file_name, 'accompaniment.mp3')
    if not os.path.exists(file_path):
        return JSONResponse(status_code=404, content={"message": "File not found"})
    return file_path