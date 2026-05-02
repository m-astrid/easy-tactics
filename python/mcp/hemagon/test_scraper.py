import httpx

def explore_hemagon():
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
    }
    
    # Try direct URL search pattern
    urls_to_try = [
        "https://hemagon.com/search?query=evgenia",
        "https://hemagon.com/fighters?q=evgenia",
        "https://hemagon.com/fighters/search?term=evgenia",
        "https://hemagon.com/api/search?type=fighter&q=evgenia",
    ]
    
    for url in urls_to_try:
        resp = httpx.get(url, headers=headers, timeout=10)
        print(f"URL: {url}")
        print(f"Status: {resp.status_code}, Length: {len(resp.text)}")
        if resp.status_code == 200 and len(resp.text) > 100:
            print(f"Content preview: {resp.text[:500]}")
        print("-" * 50)
    
    # Check main fighters page for any scripts with data
    print("\n=== Checking main page for embedded data ===")
    resp = httpx.get("https://hemagon.com/fighters", headers=headers, timeout=10)
    print(f"Status: {resp.status_code}")
    
    # Look for any window.__DATA or similar
    import re
    data_patterns = [
        r'window\.__INITIAL_STATE__\s*=\s*({.*?});',
        r'window\.__DATA__\s*=\s*({.*?});',
        r'window\.__PRELOADED_STATE__\s*=\s*({.*?});',
        r'data-fighter="([^"]+)"',
    ]
    
    for pattern in data_patterns:
        matches = re.findall(pattern, resp.text, re.DOTALL)
        if matches:
            print(f"Found pattern: {pattern[:30]}...")
            print(f"Match: {matches[0][:500]}")

if __name__ == "__main__":
    explore_hemagon()