"""
Hemagon profile scraper for HEMA tournament database
"""
from playwright.sync_api import sync_playwright


HEMAGON_BASE = "https://hemagon.com"
WEAPONS = [
    "Dussak - Women",
    "Katana",
    "Longsword - Women",
    "Longsword & Rondel",
    "Rapier - Women",
    "Rapier & Dagger",
    "Rapier & Dagger - Women",
    "Saber - Woman",
    "Spear",
    "Sword & Buckler - Women",
]


def close_cookie_banner(page):
    """Close cookie consent banner if present."""
    btn = page.query_selector('button:has-text("YES, I AGREE")')
    if btn:
        btn.click()
        page.wait_for_timeout(1000)


def scrape_summary(page, profile_url: str) -> dict:
    """Parse fighter summary from profile page."""
    page.goto(profile_url, wait_until="domcontentloaded", timeout=60000)
    page.wait_for_timeout(3000)
    close_cookie_banner(page)

    result = {}

    name_el = page.query_selector('h1')
    result["name"] = name_el.inner_text() if name_el else ""

    text = page.inner_text("body")

    if "HEMA TEAM" in text:
        result["club"] = "HEMA TEAM"

    import re

    country_match = re.search(
        r"(Russia|Ukraine|Belarus|Poland|Germany|USA|UK|Italy|France|Sweden)", text
    )
    if country_match:
        result["country"] = country_match.group(1)

    lines = [l.strip() for l in text.split("\n") if l.strip()]
    if lines:
        location_candidate = lines[-1]
        if location_candidate and len(location_candidate) < 50:
            result["location"] = location_candidate

    summary = {}
    events_match = re.search(r"EVENTS\s*(\d+)", text)
    if events_match:
        summary["events"] = int(events_match.group(1))

    cats_match = re.search(r"CATEGORIES\s*(\d+)", text)
    if cats_match:
        summary["categories"] = int(cats_match.group(1))

    fights_match = re.search(r"FIGHTS\s*(\d+)", text)
    if fights_match:
        summary["fights"] = int(fights_match.group(1))

    result["summary"] = summary

    weapons = []
    table = page.query_selector("table")
    if table:
        rows = table.query_selector_all("tr")
        for row in rows[1:]:
            cols = row.query_selector_all("td")
            if len(cols) >= 5:
                weapon = {
                    "name": cols[0].inner_text().strip(),
                    "rank": int(cols[1].inner_text().strip())
                    if cols[1].inner_text().strip().isdigit()
                    else None,
                    "rating": int(cols[2].inner_text().strip())
                    if cols[2].inner_text().strip().isdigit()
                    else None,
                    "events": int(cols[3].inner_text().strip())
                    if cols[3].inner_text().strip().isdigit()
                    else None,
                    "fights": int(cols[4].inner_text().strip())
                    if cols[4].inner_text().strip().isdigit()
                    else None,
                }
                if len(cols) > 5:
                    win_pct = cols[5].inner_text().strip().replace("%", "")
                    weapon["win_percent"] = int(win_pct) if win_pct.isdigit() else None
                weapons.append(weapon)
    result["weapons"] = weapons

    return result


def scrape_fights_for_weapon(page, fighter_name: str, weapon_name: str) -> list:
    """Scrape all fights for a specific weapon (50 per page)."""
    page.wait_for_timeout(2000)

    show_fights_btn = page.query_selector("button:has-text('SHOW FIGHTS WITH ME')")
    if show_fights_btn:
        show_fights_btn.click()
        page.wait_for_timeout(2000)

    per_page_btns = page.query_selector_all("button")
    for pbtn in per_page_btns:
        if pbtn.inner_text().strip() == "50":
            pbtn.click()
            page.wait_for_timeout(2000)
            break

    page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
    page.wait_for_timeout(1000)

    fights = []
    fight_rows = page.query_selector_all("div[class*='row'], tr")

    for row in fight_rows:
        row_text = row.inner_text()
        if fighter_name not in row_text or len(row_text) < 10:
            continue

        cols = row.query_selector_all("div[class*='col'], td")
        if len(cols) < 4:
            continue

        try:
            user_score = int(cols[0].inner_text().strip())
            opponent = cols[1].inner_text().strip()
            opponent_score = int(cols[2].inner_text().strip())
        except (ValueError, IndexError):
            continue

        if not opponent or opponent == fighter_name or len(opponent) < 3:
            continue
        if "Round" in opponent or "Pool" in opponent:
            continue

        result_val = "win" if user_score > opponent_score else "lose"

        club = None
        club_el = cols[1].query_selector("span") if len(cols) > 1 else None
        if club_el:
            club = club_el.inner_text().strip()

        extra = ""
        if len(cols) > 3:
            extra = cols[3].inner_text().strip()

        tournament = ""
        stage = ""
        date = ""
        if extra:
            parts = extra.split("\t")
            if len(parts) >= 1:
                tournament = parts[0].strip()
            if len(parts) >= 2:
                stage = parts[1].strip()
            if len(parts) >= 3:
                date = parts[2].strip()

        fights.append(
            {
                "opponent": opponent,
                "club": club,
                "user_score": user_score,
                "opponent_score": opponent_score,
                "result": result_val,
                "tournament": tournament,
                "stage": stage,
                "date": date,
            }
        )

    return fights


def scrape_all_fights(profile_url: str, fighter_name: str) -> list:
    """Scrape fights for all weapons."""
    all_fights = []
    base_url = profile_url.rstrip("/") + "/stats"

    with sync_playwright() as p:
        browser = p.chromium.launch(
            headless=True,
            args=["--disable-blink-features=AutomationControlled"],
        )
        context = browser.new_context(
            user_agent="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
            viewport={"width": 1920, "height": 1080},
        )
        page = context.new_page()

        for weapon_name in WEAPONS:
            print(f"=== Processing: {weapon_name} ===")

            page.goto(base_url, wait_until="domcontentloaded", timeout=60000)
            page.wait_for_timeout(2000)
            close_cookie_banner(page)

            weapon_link = page.query_selector(f'a.flex:has-text("{weapon_name}")')
            if not weapon_link:
                short_name = weapon_name.split(" - ")[0].split(" & ")[0]
                weapon_link = page.query_selector(f'a.flex:has-text("{short_name}")')

            if weapon_link:
                weapon_link.click()
                page.wait_for_timeout(1500)

            fights = scrape_fights_for_weapon(page, fighter_name, weapon_name)
            print(f"  Found {len(fights)} fights")
            all_fights.extend(fights)

        browser.close()

    return all_fights


def scrape_tournaments(profile_url: str, context) -> list:
    """Scrape tournament list with VK links."""
    tournaments = []
    stats_url = profile_url.rstrip("/") + "/stats"

    page = context.new_page()
    page.goto(stats_url, wait_until="domcontentloaded", timeout=60000)
    page.wait_for_timeout(3000)
    close_cookie_banner(page)

    seen = set()
    tournament_links = page.query_selector_all('a[href*="/tournament/"]')

    for link in tournament_links:
        href = link.get_attribute("href") or ""
        text = link.inner_text().strip()

        if not href or "/nomination/" in href or href in seen:
            continue
        seen.add(href)

        if not text or len(text) < 2:
            continue

        import re

        nomination_match = re.search(r"(.+?)\s*\t\s*(.+)", text)
        category = ""
        if nomination_match:
            category = nomination_match.group(2).strip()

        vk_link = None
        try:
            tourney_page = context.new_page()
            tourney_page.goto(HEMAGON_BASE + href, wait_until="domcontentloaded", timeout=30000)
            tourney_page.wait_for_timeout(2000)

            vk_anchors = tourney_page.query_selector_all('a[href*="vk.com"]')
            for vk_anchor in vk_anchors:
                vk_link = vk_anchor.get_attribute("href")
                if vk_link:
                    break
            tourney_page.close()
        except Exception as e:
            print(f"Error fetching VK for {text}: {e}")

        tournaments.append(
            {
                "name": text.split("\t")[0].strip(),
                "url": href,
                "slug": href.split("/")[-1] if href else "",
                "category": category,
                "vk_link": vk_link,
            }
        )

    page.close()
    return tournaments


def scrape_profile(profile_url: str) -> dict:
    """
    Scrape a fighter's profile from hemagon.com

    Args:
        profile_url: URL to the fighter's profile (e.g., https://hemagon.com/users/eugeniya.shumakova)

    Returns:
        Dictionary with fighter data including name, club, location, summary, weapons, fights, and tournaments
    """
    result = {}

    with sync_playwright() as p:
        browser = p.chromium.launch(
            headless=True,
            args=["--disable-blink-features=AutomationControlled"],
        )

        context = browser.new_context(
            user_agent="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
            viewport={"width": 1920, "height": 1080},
        )

        page = context.new_page()

        page.goto(HEMAGON_BASE + "/", wait_until="domcontentloaded", timeout=60000)
        page.wait_for_timeout(1000)
        close_cookie_banner(page)

        summary_data = scrape_summary(page, profile_url)
        result.update(summary_data)

        fighter_name = summary_data.get("name", "")

        all_fights = []
        for weapon_name in WEAPONS:
            print(f"=== Processing: {weapon_name} ===")

            stats_url = profile_url.rstrip("/") + "/stats"
            page.goto(stats_url, wait_until="domcontentloaded", timeout=60000)
            page.wait_for_timeout(2000)
            close_cookie_banner(page)

            weapon_link = page.query_selector(f'a.flex:has-text("{weapon_name}")')
            if not weapon_link:
                short_name = weapon_name.split(" - ")[0].split(" & ")[0]
                weapon_link = page.query_selector(f'a.flex:has-text("{short_name}")')

            if weapon_link:
                weapon_link.click()
                page.wait_for_timeout(1500)

            show_fights_btn = page.query_selector("button:has-text('SHOW FIGHTS WITH ME')")
            if show_fights_btn:
                show_fights_btn.click()
                page.wait_for_timeout(2000)

            # Debug: check what's visible
            page_text = page.inner_text("body")
            print(f"  Page contains 'SHOW FIGHTS': {'SHOW FIGHTS' in page_text}")
            print(f"  Page contains 'Евгения': {'Евгения' in page_text}")

            # Debug: list all buttons
            all_buttons = page.query_selector_all("button")
            print(f"  All buttons: {[b.inner_text().strip()[:30] for b in all_buttons[:15]]}")

            # Debug: list tables
            tables = page.query_selector_all("table")
            print(f"  Tables count: {len(tables)}")
            if tables:
                for i, t in enumerate(tables):
                    rows = t.query_selector_all("tr")
                    print(f"    Table {i}: {len(rows)} rows")

            per_page_btns = page.query_selector_all("button")
            for pbtn in per_page_btns:
                if pbtn.inner_text().strip() == "50":
                    pbtn.click()
                    page.wait_for_timeout(2000)
                    break

            page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
            page.wait_for_timeout(1000)

            # Try table rows first, fallback to div rows
            rows = page.query_selector_all("table tbody tr")
            if not rows:
                rows = page.query_selector_all("div[class*='row']")
            seen_opponents = set()

            for row in rows:
                # Try table cells first, then divs
                cols = row.query_selector_all("td")
                if not cols or len(cols) < 3:
                    cols = row.query_selector_all("div")
                if len(cols) < 3:
                    continue

                row_text = row.inner_text()
                if not fighter_name or fighter_name not in row_text or len(row_text) < 10:
                    continue

                try:
                    user_score_text = cols[0].inner_text().strip()
                    opponent = cols[1].inner_text().strip()
                    opponent_score_text = cols[2].inner_text().strip()

                    if not user_score_text.isdigit() or not opponent_score_text.isdigit():
                        continue

                    user_score = int(user_score_text)
                    opponent_score = int(opponent_score_text)
                except (ValueError, IndexError):
                    continue

                if (
                    not opponent
                    or opponent == fighter_name
                    or len(opponent) < 3
                    or "Round" in opponent
                    or "Pool" in opponent
                ):
                    continue

                if opponent in seen_opponents:
                    continue
                seen_opponents.add(opponent)

                result_val = "win" if user_score > opponent_score else "lose"

                extra = ""
                if len(cols) > 3:
                    extra = cols[3].inner_text().strip()

                tournament = ""
                stage = ""
                date = ""
                if extra:
                    parts = extra.split("\t")
                    if len(parts) >= 1:
                        tournament = parts[0].strip()
                    if len(parts) >= 2:
                        stage = parts[1].strip()
                    if len(parts) >= 3:
                        date = parts[2].strip()

                all_fights.append(
                    {
                        "opponent": opponent,
                        "club": None,
                        "user_score": user_score,
                        "opponent_score": opponent_score,
                        "result": result_val,
                        "tournament": tournament,
                        "stage": stage,
                        "date": date,
                    }
                )

            print(f"  Found {len(all_fights)} fights so far")

        result["fights"] = all_fights

        tournaments = scrape_tournaments(profile_url, context)
        result["tournaments"] = tournaments

        browser.close()

    return result


if __name__ == "__main__":
    import sys
    import json

    url = sys.argv[1] if len(sys.argv) > 1 else "https://hemagon.com/users/eugeniya.shumakova"
    result = scrape_profile(url)
    print(json.dumps(result, indent=2, ensure_ascii=False))