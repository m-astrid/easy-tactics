"""
Read user data from files and analyze with LLM
"""
import os
import glob
from typing import Optional
from app.llm_client import create_llm_client, LLMClient


def get_prompt_template() -> str:
    """Load the prompt template from file."""
    prompt_path = os.path.join(os.path.dirname(__file__), "..", "prompts", "structure_loaded_user_data.md")
    with open(prompt_path, "r", encoding="utf-8") as f:
        return f.read()


def read_all_files(data_dir: str) -> dict[str, str]:
    """Read all .txt and .json files from directory."""
    files_content = {}
    
    txt_files = glob.glob(os.path.join(data_dir, "*.txt"))
    json_files = glob.glob(os.path.join(data_dir, "*.json"))
    
    for filepath in txt_files + json_files:
        filename = os.path.basename(filepath)
        with open(filepath, "r", encoding="utf-8") as f:
            files_content[filename] = f.read()
    
    return files_content


def format_files_for_prompt(files: dict[str, str]) -> str:
    """Format file contents for inclusion in prompt."""
    formatted = []
    for filename, content in sorted(files.items()):
        formatted.append(f"\n=== {filename} ===\n{content}")
    return "\n".join(formatted)


def analyze_user_data(data_dir: str, llm_client: Optional[LLMClient] = None) -> dict:
    """
    Read user data files and analyze with LLM.
    
    Args:
        data_dir: Directory containing scraped data files
        llm_client: Optional LLM client (creates default if not provided)
    
    Returns:
        Structured JSON with user profile data
    """
    if llm_client is None:
        llm_client = create_llm_client()
    
    files = read_all_files(data_dir)
    
    if not files:
        return {"error": "No files found in directory", "directory": data_dir}
    
    template = get_prompt_template()
    files_content = format_files_for_prompt(files)
    
    user_prompt = template.replace("{__filecontent__}", files_content)
    
    result = llm_client.chat(
        system_prompt="Ты — ассистент для структурирования данных о бойцах HEMA. Ты анализируешь текстовые файлы с данными о выступлениях и формируешь JSON в точности по указанной схеме.",
        user_prompt=user_prompt
    )
    
    return result


def analyze_user_data_sync(data_dir: str) -> dict:
    """Synchronous wrapper for analyze_user_data."""
    return analyze_user_data(data_dir)