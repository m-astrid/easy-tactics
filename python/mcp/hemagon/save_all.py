from playwright.sync_api import sync_playwright
import os
import json
import sys

script_dir = os.path.dirname(os.path.abspath(__file__))

profile_url_arg = sys.argv[1] if len(sys.argv) > 1 else "eugeniya.shumakova"
output_dir = os.path.join(script_dir, f"hemagon_data_{profile_url_arg}")
os.makedirs(output_dir, exist_ok=True)

def save_page_text(page, filename):
    filepath = os.path.join(output_dir, filename)
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(page.inner_text('body'))

def save_tournament_with_links(page, filename):
    filepath = os.path.join(output_dir, filename)
    content = {
        'text': page.inner_text('body'),
        'links': []
    }
    all_links = page.query_selector_all('a[href]')
    for link in all_links:
        href = link.get_attribute('href')
        text = link.inner_text().strip()
        if href and text:
            content['links'].append({'text': text, 'href': href})
    with open(filepath, "w", encoding="utf-8") as f:
        json.dump(content, f, ensure_ascii=False, indent=2)
    
    vk_links = [l for l in content['links'] if 'vk.com' in l['href']]
    for vk in vk_links:
        vk_filename = f"vk_{vk['href'].split('/')[-1].replace('?from=groups', '')}.json"
        vk_filepath = os.path.join(output_dir, vk_filename)
        with open(vk_filepath, "w", encoding="utf-8") as f:
            json.dump({
                'source_tournament': filename,
                'vk_link': vk['href'],
                'vk_name': vk['text']
            }, f, ensure_ascii=False, indent=2)
        print(f"  Saved VK: {vk_filename}")

def set_fights_per_page(page, count=50):
    page.wait_for_timeout(1000)
    parent_div = page.query_selector('text=Per page')
    if parent_div:
        btn_group = parent_div.query_selector('xpath=../div[contains(@class, "btn-group")]')
        if btn_group:
            buttons = btn_group.query_selector_all('button')
            for btn in buttons:
                if btn.inner_text().strip() == str(count):
                    btn.click()
                    page.wait_for_timeout(2000)
                    return

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    context = browser.new_context(
        user_agent='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
        viewport={'width': 1920, 'height': 1080},
    )
    page = context.new_page()
    
    profile_url = f"https://hemagon.com/users/{profile_url_arg}"
    
    print("=== 1. Profile page ===")
    page.goto(profile_url, wait_until="domcontentloaded", timeout=60000)
    page.wait_for_timeout(3000)
    
    btn = page.query_selector('button:has-text("YES, I AGREE")')
    if btn:
        btn.click()
        page.wait_for_timeout(1000)
    
    save_page_text(page, "profile.txt")
    print("  Saved: profile.txt")
    
    print("\n=== 2. Stats page - iterating through weapons ===")
    page.goto(profile_url + "/stats", wait_until="domcontentloaded", timeout=60000)
    page.wait_for_timeout(3000)
    
    weapon_names = ['Dussak - Women', 'Katana', 'Longsword - Women', 'Longsword & Rondel', 'Rapier - Women', 'Rapier & Dagger', 'Rapier & Dagger - Women', 'Saber - Woman', 'Spear', 'Sword & Buckler - Women']
    
    processed_tournaments = set()
    
    for weapon_name in weapon_names:
        print(f"  Processing weapon: {weapon_name}")
        
        links = page.query_selector_all('a')
        print(f"===============================")
        print(f"Found {len(links)} links: {[link.inner_text() for link in links]}")
        weapon_link = None
        for link in links:
            if link.inner_text().strip() == weapon_name:
                weapon_link = link
                break
        
        if weapon_link:
            weapon_link.click()
            page.wait_for_timeout(2000)
            
            save_page_text(page, f"weapon_{weapon_name.replace(' ', '_').replace('&', 'and')}.txt")
            print(f"    Saved: weapon_{weapon_name.replace(' ', '_').replace('&', 'and')}.txt")
            
            tournament_links = page.query_selector_all('a[href*="/tournament/"]')
            weapon_tournaments = set()
            for link in tournament_links:
                href = link.get_attribute('href')
                if href and '/tournament/' in href and '/nomination/' not in href:
                    slug = href.split('/')[-1]
                    if slug and slug not in processed_tournaments:
                        weapon_tournaments.add(slug)
            
            for slug in weapon_tournaments:
                processed_tournaments.add(slug)
                page.goto(f"https://hemagon.com/tournament/{slug}", wait_until="domcontentloaded", timeout=60000)
                page.wait_for_timeout(2000)
                save_tournament_with_links(page, f"tournament_{slug}.json")
            
            if weapon_tournaments:
                print(f"    Saved {len(weapon_tournaments)} tournament pages")
            
            page.goto(profile_url + "/stats", wait_until="domcontentloaded", timeout=60000)
            page.wait_for_timeout(2000)
            
            show_fights_btn = page.query_selector('button:has-text("SHOW FIGHTS WITH ME")')
            if show_fights_btn:
                show_fights_btn.click()
                page.wait_for_timeout(2000)
                
                page.evaluate('window.scrollTo(0, document.body.scrollHeight)')
                page.wait_for_timeout(1000)
                set_fights_per_page(page, 50)
                
                save_page_text(page, f"weapon_{weapon_name.replace(' ', '_').replace('&', 'and')}_fights.txt")
                print(f"    Saved fights")
    
    print(f"\n=== Done! Files saved to {output_dir} ===")
    browser.close()