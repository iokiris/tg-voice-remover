import os
from fastapi import FastAPI, HTTPException, UploadFile, File, Form
from fastapi.responses import JSONResponse
from spleeter.separator import Separator
from pydantic import BaseModel
import threading

app = FastAPI()
separator = Separator('spleeter:2stems')


class AudioProcessRequest(BaseModel):
    input_path: str
    output_path: str

def separate_audio(input_path, output_path):
    try:
        separator.separate_to_file(input_path, output_path, codec='mp3')
        return True
    except Exception as e:
        print(f"Error processing audio: {str(e)}")
        return False

@app.post("/process-audio/")
async def process_audio(request: AudioProcessRequest):
    output_dir = request.output_path
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    thread = threading.Thread(target=separate_audio, args=(request.input_path, output_dir))
    thread.start()
    thread.join()

    if thread.is_alive():
        return JSONResponse(status_code=500, content={"message": "Processing failed to complete in time"})

    return {"message": "Audio is being processed", "input": request.input_path, "output": output_dir}

