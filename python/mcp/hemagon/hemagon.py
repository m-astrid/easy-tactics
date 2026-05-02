"""
Hemagon profile scraper for HEMA tournament database
"""
import re
from playwright.sync_api import sync_playwright


def scrape_profile(profile_url: str) -> dict:
    """
    Scrape a fighter's profile from hemagon.com
    
    Args:
        profile_url: URL to the fighter's profile (e.g., https://hemagon.com/users/eugeniya.shumakova)
    
    Returns:
        Dictionary with fighter data including name, club, location, summary, weapons, fights, and tournaments
    """
    with sync_playwright() as p:
        browser = p.chromium.launch(
            headless=True,
            args=['--disable-blink-features=AutomationControlled']
        )
        
        context = browser.new_context(
            user_agent='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
            viewport={'width': 1920, 'height': 1080},
        )
        
        page = context.new_page()
        
        page.goto("https://hemagon.com/", wait_until="domcontentloaded", timeout=60000)
        page.wait_for_timeout(1000)
        
        btn = page.query_selector('button:has-text("YES, I AGREE")')
        if btn:
            btn.click()
            page.wait_for_timeout(1000)
        
        page.goto(profile_url, wait_until="domcontentloaded", timeout=60000)
        page.wait_for_timeout(3000)
        
        result = {}
        
        name_el = page.query_selector('h1')
        result['name'] = name_el.inner_text() if name_el else None
        
        text = page.inner_text('body')
        
        club_el = page.query_selector('text=HEMA TEAM')
        if club_el:
            result['club'] = "HEMA TEAM"
        
        country_match = re.search(r'(Russia|Ukraine|Belarus|Poland|Germany|USA|UK|Italy|France|Sweden)', text)
        if country_match:
            result['country'] = country_match.group(1)
        
        location_parts = text.split('/')
        if len(location_parts) >= 2:
            location_candidate = location_parts[-1].strip().split('\n')[0]
            if location_candidate and len(location_candidate) < 50:
                result['location'] = location_candidate
        
        summary = {}
        
        events_match = re.search(r'EVENTS\s*(\d+)', text)
        if events_match:
            summary['events'] = int(events_match.group(1))
        
        cats_match = re.search(r'CATEGORIES\s*(\d+)', text)
        if cats_match:
            summary['categories'] = int(cats_match.group(1))
        
        fights_match = re.search(r'FIGHTS\s*(\d+)', text)
        if fights_match:
            summary['fights'] = int(fights_match.group(1))
        
        result['summary'] = summary
        
        weapons = []
        table = page.query_selector('table')
        if table:
            rows = table.query_selector_all('tr')
            for row in rows[1:]:
                cols = row.query_selector_all('td')
                if len(cols) >= 5:
                    weapon = {
                        'name': cols[0].inner_text().strip(),
                        'rank': int(cols[1].inner_text().strip()) if cols[1].inner_text().strip().isdigit() else None,
                        'rating': int(cols[2].inner_text().strip()) if cols[2].inner_text().strip().isdigit() else None,
                        'events': int(cols[3].inner_text().strip()) if cols[3].inner_text().strip().isdigit() else None,
                        'fights': int(cols[4].inner_text().strip()) if cols[4].inner_text().strip().isdigit() else None,
                    }
                    if len(cols) > 5:
                        win_pct = cols[5].inner_text().strip().replace('%', '')
                        weapon['win_percent'] = int(win_pct) if win_pct.isdigit() else None
                    weapons.append(weapon)
        result['weapons'] = weapons
        
        page.goto(profile_url.rstrip('/') + '/stats', wait_until="domcontentloaded", timeout=60000)
        page.wait_for_timeout(3000)
        
        show_fights_btn = page.query_selector('button:has-text("SHOW FIGHTS WITH ME")')
        if show_fights_btn:
            show_fights_btn.click()
            page.wait_for_timeout(2000)
        
        page_text = page.inner_text('body')
        
        lines = [l for l in page_text.split('\n') if l.strip()]
        
        date_pattern = re.compile(r'[A-Za-z]{3}\s+\d{1,2},\s+\d{4}')
        
        fights_data = []
        my_name = result.get('name', '')
        
        for i, line in enumerate(lines):
            if date_pattern.search(line) and line.startswith('\t'):
                parts = line.split('\t')
                if len(parts) >= 5:
                    date = parts[-1].strip()
                    tournament = parts[1].strip() if parts[1].strip() else ''
                    stage = parts[3].strip() if len(parts) > 3 and ('Round' in parts[3] or 'Pool' in parts[3]) else ''
                    
                    line_before = lines[i - 1].strip()
                    line_before2 = lines[i - 2].strip() if i >= 2 else ''
                    line_before3 = lines[i - 3].strip() if i >= 3 else ''
                    
                    try:
                        if line_before.isdigit():
                            opponent_score = int(line_before)
                            if line_before2 and not line_before2[0].isupper() and not line_before2.isdigit():
                                opponent_name = line_before3
                                opponent_club = line_before2
                            elif line_before2 and line_before2[0].isupper():
                                opponent_name = line_before2
                                opponent_club = ''
                            else:
                                continue
                        else:
                            continue
                    except (ValueError, IndexError):
                        continue
                    
                    user_score_idx = i - 4
                    user_score_str = lines[user_score_idx].strip() if user_score_idx >= 0 else ''
                    
                    try:
                        user_score = int(user_score_str)
                    except ValueError:
                        continue
                    
                    if not opponent_name or len(opponent_name) < 2:
                        continue
                    
                    result_val = 'win' if user_score > opponent_score else 'lose' if user_score < opponent_score else 'draw'
                    
                    fights_data.append({
                        'opponent': opponent_name,
                        'club': opponent_club if opponent_club else None,
                        'user_score': user_score,
                        'opponent_score': opponent_score,
                        'result': result_val,
                        'tournament': tournament,
                        'stage': stage,
                        'date': date
                    })
            if date_pattern.search(line):
                parts = line.split('\t')
                if len(parts) >= 4:
                    date = parts[-1].strip()
                    rest = '\t'.join(parts[:-1])
                    score_parts = re.split(r'\s+', rest.strip())
                    
                    if len(score_parts) >= 4:
                        try:
                            my_score = int(score_parts[-2])
                            opponent_score = int(score_parts[-1])
                            
                            opponent_and_club = ' '.join(score_parts[:-2]).rsplit(' ', 1)
                            if len(opponent_and_club) >= 2:
                                opponent = opponent_and_club[0].strip()
                                club = opponent_and_club[1].strip() if len(opponent_and_club) > 1 else None
                            else:
                                continue
                            
                            if not opponent or len(opponent) < 2:
                                continue
                                
                            result_val = 'win' if my_score > opponent_score else 'lose' if my_score < opponent_score else 'draw'
                            
                            for j in range(i - 1, max(0, i - 5), -1):
                                if '\t' in lines[j]:
                                    tournament_parts = lines[j].split('\t')
                                    if len(tournament_parts) >= 3:
                                        tournament = tournament_parts[0].strip()
                                        stage = tournament_parts[1].strip() if 'Round' in tournament_parts[1] or 'Pool' in tournament_parts[1] else ''
                                        break
                                    break
                            else:
                                tournament = ''
                                stage = ''
                            
                            fights_data.append({
                                'opponent': opponent,
                                'club': club if club and '\t' not in club else None,
                                'user_score': my_score,
                                'opponent_score': opponent_score,
                                'result': result_val,
                                'tournament': tournament,
                                'stage': stage,
                                'date': date
                            })
                        except (ValueError, IndexError):
                            continue
        
        result['fights'] = fights_data[:50]
        
        page.goto(profile_url.rstrip('/') + '/stats', wait_until="domcontentloaded", timeout=60000)
        page.wait_for_timeout(3000)
        
        tournament_links = page.query_selector_all('a[href*="/tournament/"]')
        
        tournaments = []
        seen = set()
        for link in tournament_links:
            href = link.get_attribute('href') or ''
            text = link.inner_text().strip()
            if '/nomination/' in href or href in seen:
                continue
            seen.add(href)
            if text and len(text) > 1:
                nomination = ""
                all_links = page.query_selector_all(f'a[href*="{href.split("/")[-1]}"]')
                for nl in all_links:
                    if '/nomination/' in nl.get_attribute('href') or '':
                        nomination = nl.inner_text().strip()
                        break
                
                vk_link = None
                if href:
                    tourney_page = context.new_page()
                    try:
                        tourney_page.goto("https://hemagon.com" + href, wait_until="domcontentloaded", timeout=30000)
                        tourney_page.wait_for_timeout(2000)
                        
                        all_anchors = tourney_page.query_selector_all('a[href*="vk.com"], a[href*="vkontakte"]')
                        for anchor in all_anchors:
                            vk_link = anchor.get_attribute('href')
                            if vk_link:
                                break
                    except Exception as e:
                        print(f"Error fetching tournament {text}: {e}")
                    finally:
                        tourney_page.close()
                
                tournaments.append({
                    'name': text,
                    'url': href,
                    'slug': href.split('/')[-1] if href else None,
                    'category': nomination,
                    'vk_link': vk_link
                })
        
        result['tournaments'] = tournaments
        
        browser.close()
        return result


if __name__ == "__main__":
    result = scrape_profile("https://hemagon.com/users/eugeniya.shumakova")
    import json
    print(json.dumps(result, indent=2, ensure_ascii=False))