from playwright.sync_api import sync_playwright
import os

script_dir = os.path.dirname(os.path.abspath(__file__))

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()
    
    page.goto("https://hemagon.com/users/eugeniya.shumakova", wait_until="domcontentloaded", timeout=60000)
    page.wait_for_timeout(3000)
    
    btn = page.query_selector('button:has-text("YES, I AGREE")')
    if btn:
        btn.click()
        page.wait_for_timeout(1000)
    
    profile_text = page.inner_text('body')
    
    with open(os.path.join(script_dir, "profile_text.txt"), "w", encoding="utf-8") as f:
        f.write(profile_text)
    print("Saved profile_text.txt")
    
    page.goto("https://hemagon.com/users/eugeniya.shumakova/stats", wait_until="domcontentloaded", timeout=60000)
    page.wait_for_timeout(3000)
    
    show_fights_btn = page.query_selector('button:has-text("SHOW FIGHTS WITH ME")')
    if show_fights_btn:
        show_fights_btn.click()
        page.wait_for_timeout(2000)
    
    stats_text = page.inner_text('body')
    
    with open(os.path.join(script_dir, "stats_text.txt"), "w", encoding="utf-8") as f:
        f.write(stats_text)
    print("Saved stats_text.txt")
    
    browser.close()