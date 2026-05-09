"""
HEMAGON FastAPI Server
"""
from fastapi import FastAPI
from pydantic import BaseModel
from scraper import load_user_profile_async
from datetime import datetime

app = FastAPI(title="HEMAGON API")


class LoadProfileRequest(BaseModel):
    profile_link: str
    date_from: datetime.DateTime
    date_to: datetime.DateTime


@app.post("/load_user_profile")
async def load_user_profile_endpoint(request: LoadProfileRequest):
    """Load a HEMA fighter's profile from hemagon.com"""
    result = await load_user_profile_async(request.profile_link, request.date_from, request.date_to)
    return result


@app.get("/health")
async def health():
    return {"status": "ok"}