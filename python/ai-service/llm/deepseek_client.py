"""
DeepSeek LLM Client
"""
import os
import json
import requests
from typing import Optional


class LLMClient:
    def __init__(self, api_key: Optional[str] = None, base_url: str = "https://api.deepseek.com"):
        self.api_key = api_key or os.getenv("DEEPSEEK_API_KEY")
        self.base_url = base_url
        if not self.api_key:
            raise ValueError("DEEPSEEK_API_KEY environment variable not set")
    
    def chat(self, system_prompt: str, user_prompt: str, model: str = "deepseek-chat") -> dict:
        """Send chat request to DeepSeek and return parsed JSON response."""
        url = f"{self.base_url}/v1/chat/completions"
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
        payload = {
            "model": model,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt}
            ],
            "temperature": 0.1,
            "response_format": {"type": "json_object"}
        }
        
        response = requests.post(url, headers=headers, json=payload, timeout=120)
        response.raise_for_status()
        
        result = response.json()
        content = result["choices"][0]["message"]["content"]
        
        try:
            return json.loads(content)
        except json.JSONDecodeError:
            return {"error": "Failed to parse JSON", "raw_content": content}


def create_llm_client() -> LLMClient:
    """Factory function to create LLM client."""
    return LLMClient()