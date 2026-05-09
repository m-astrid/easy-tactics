"""
Load user data from hemagon.com, save to files, and analyze with LLM
"""
import os
import uuid
import json
import requests
from typing import Optional
from app.read_user_data import analyze_user_data
from app.llm_client import create_llm_client, LLMClient


DEFAULT_HEMAGON_API_URL = os.getenv("HEMAGON_API_URL", "http://localhost:8000")


def load_and_analyze(
    profile_link: str,
    target_dir: Optional[str] = None,
    hemagon_api_url: str = DEFAULT_HEMAGON_API_URL,
    llm_client: Optional[LLMClient] = None
) -> dict:
    """
    Load profile from hemagon.com, save files, and analyze with LLM.
    
    Args:
        profile_link: URL to hemagon profile (e.g., 'https://hemagon.com/users/nekrasova')
        target_dir: Directory to save files (auto-generated if not provided)
        hemagon_api_url: URL of hemagon API service
        llm_client: Optional LLM client
    
    Returns:
        Dictionary with analysis result and file paths
    """
    if target_dir is None:
        target_dir = os.path.join("/tmp", "hemagon_data", str(uuid.uuid4()))
    
    os.makedirs(target_dir, exist_ok=True)
    
    response = requests.post(
        f"{hemagon_api_url}/load_user_profile",
        json={"profile_link": profile_link, "target_dir": target_dir},
        timeout=300
    )
    response.raise_for_status()
    
    scrape_result = response.json()
    
    result = analyze_user_data(target_dir, llm_client)
    
    result["target_dir"] = target_dir
    result["files_saved"] = scrape_result.get("files_saved", [])
    
    result_json_path = os.path.join(target_dir, "result.json")
    with open(result_json_path, "w", encoding="utf-8") as f:
        json.dump(result, f, ensure_ascii=False, indent=2)
    result["files_saved"].append("result.json")
    
    return result


def load_and_analyze_sync(
    profile_link: str,
    target_dir: Optional[str] = None,
    hemagon_api_url: str = DEFAULT_HEMAGON_API_URL
) -> dict:
    """Synchronous wrapper for load_and_analyze."""
    return load_and_analyze(profile_link, target_dir, hemagon_api_url)