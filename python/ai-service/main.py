"""
HEMAGON AI Service - FastAPI Server
"""
import os
import json
from typing import Optional
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from app.load_user_data import load_and_analyze_sync
from app.read_user_data import analyze_user_data_sync

app = FastAPI(title="HEMAGON AI Service")

HEMAGON_API_URL = os.getenv("HEMAGON_API_URL", "http://localhost:8000")


class WeaponSchema(BaseModel):
    name: str
    rank: int | None = None
    rating: int | None = None
    events: int | None = None
    fights: int | None = None
    win_percent: int | None = None


class FightSchema(BaseModel):
    opponent: str
    club: str | None = None
    user_score: int
    opponent_score: int
    result: str
    tournament: str
    stage: str | None = None
    date: str
    weapon: str


class TournamentSchema(BaseModel):
    name: str
    url: str | None = None
    slug: str | None = None
    category: str | None = None
    vk_link: str | None = None


class SummarySchema(BaseModel):
    events: int | None = None
    categories: int | None = None
    fights: int | None = None


class ProfileDataSchema(BaseModel):
    name: str | None = None
    club: str | None = None
    country: str | None = None
    location: str | None = None
    summary: SummarySchema | None = None
    weapons: list[WeaponSchema] = []
    fights: list[FightSchema] = []
    tournaments: list[TournamentSchema] = []


class AnalyzeResponse(BaseModel):
    profile: ProfileDataSchema
    target_dir: str
    files_saved: list[str] = []


class AnalyzeProfileRequest(BaseModel):
    profile_link: str


class AnalyzeExistingRequest(BaseModel):
    data_dir: str


@app.post("/load_or_update_profile", response_model=AnalyzeResponse)
async def load_or_update_profile(request: AnalyzeProfileRequest):
    """
    Load profile from hemagon.com, save files, and analyze with LLM.
    """
    try:
        result = load_and_analyze_sync(
            profile_link=request.profile_link,
            hemagon_api_url=HEMAGON_API_URL
        )
        return AnalyzeResponse(
            profile=ProfileDataSchema(**result.get("profile", {})),
            target_dir=result.get("target_dir", ""),
            files_saved=result.get("files_saved", [])
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/get_existing_profile", response_model=AnalyzeResponse)
async def get_existing_profile(request: AnalyzeExistingRequest):
    """
    Get existing profile data. If result.json exists, return it. Otherwise analyze files.
    """
    if not os.path.isdir(request.data_dir):
        raise HTTPException(status_code=404, detail="Directory not found")
    
    result_json_path = os.path.join(request.data_dir, "result.json")
    
    if os.path.isfile(result_json_path):
        with open(result_json_path, "r", encoding="utf-8") as f:
            result = json.load(f)
        return AnalyzeResponse(
            profile=ProfileDataSchema(**result.get("profile", {})),
            target_dir=result.get("target_dir", request.data_dir),
            files_saved=result.get("files_saved", [])
        )
    
    try:
        result = analyze_user_data_sync(request.data_dir)
        return AnalyzeResponse(
            profile=ProfileDataSchema(**result.get("profile", {})),
            target_dir=request.data_dir,
            files_saved=result.get("files_saved", [])
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/health")
async def health():
    return {"status": "ok"}