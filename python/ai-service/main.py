"""
HEMAGON AI Service - FastAPI Server
"""
import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from app.load_user_data import load_and_analyze_sync
from app.read_user_data import analyze_user_data_sync

app = FastAPI(title="HEMAGON AI Service")

HEMAGON_API_URL = os.getenv("HEMAGON_API_URL", "http://localhost:8000")


class AnalyzeProfileRequest(BaseModel):
    profile_link: str


class AnalyzeExistingRequest(BaseModel):
    data_dir: str


@app.post("/analyze_profile")
async def analyze_profile(request: AnalyzeProfileRequest):
    """
    Load profile from hemagon.com, save files, and analyze with LLM.
    """
    try:
        result = load_and_analyze_sync(
            profile_link=request.profile_link,
            hemagon_api_url=HEMAGON_API_URL
        )
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/analyze_existing")
async def analyze_existing(request: AnalyzeExistingRequest):
    """
    Analyze already saved profile data with LLM.
    """
    if not os.path.isdir(request.data_dir):
        raise HTTPException(status_code=404, detail="Directory not found")
    
    try:
        result = analyze_user_data_sync(request.data_dir)
        result["target_dir"] = request.data_dir
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/health")
async def health():
    return {"status": "ok"}