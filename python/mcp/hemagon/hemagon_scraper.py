"""
Hemagon profile scraper - saves rendered text content (like browser user would copy/paste)
"""
import os
import json
import tempfile
from playwright.sync_api import sync_playwright
from urllib.parse import urljoin

HEMAGON_BASE = "https://hemagon.com"
TEMP_DIR = tempfile.mkdtemp(prefix="hemagon_")

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


def save_text_content(page, filename: str):
    """Save rendered text content to file (like browser user copy/paste)."""
    filepath = os.path.join(TEMP_DIR, filename)
    content = page.inner_text("body")
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(content)
    print(f"  Saved: {filepath}")
    return filepath


def scrape_profile_to_files(profile_url: str):
    """Navigate and save all pages text content."""
    print(f"Temp directory: {TEMP_DIR}")

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

        print("=== Saving profile page ===")
        save_text_content(page, "profile.txt")

        stats_url = profile_url.rstrip("/") + "/stats"
        print("=== Saving stats page ===")
        page.goto(stats_url, wait_until="domcontentloaded", timeout=60000)
        page.wait_for_timeout(3000)
        close_cookie_banner(page)
        save_text_content(page, "stats.txt")

        for weapon_name in WEAPONS:
            print(f"=== Processing weapon: {weapon_name} ===")

            page.goto(stats_url, wait_until="domcontentloaded", timeout=60000)
            page.wait_for_timeout(2000)
            close_cookie_banner(page)

            weapon_link = page.query_selector(f'a.flex:has-text("{weapon_name}")')
            if not weapon_link:
                short_name = weapon_name.split(" - ")[0].split(" & ")[0]
                weapon_link = page.query_selector(f'a.flex:has-text("{short_name}")')

            if weapon_link:
                weapon_link.click()
                page.wait_for_timeout(2000)

            page.evaluate("window.scrollTo(0, 500)")
            page.wait_for_timeout(1000)

            show_fights_btn = page.query_selector("button:has-text('SHOW FIGHTS WITH ME')")
            if show_fights_btn:
                try:
                    show_fights_btn.click(timeout=5000)
                except:
                    page.evaluate("document.querySelector('button:has-text(\"SHOW FIGHTS WITH ME\")').click()")
                page.wait_for_timeout(6000)
            else:
                print(f"  Button not found: {weapon_name}")

            per_page_btns = page.query_selector_all("button")

            per_page_btns = page.query_selector_all("button")
            for pbtn in per_page_btns:
                if pbtn.inner_text().strip() == "50":
                    pbtn.click()
                    page.wait_for_timeout(3000)
                    break

            safe_name = weapon_name.replace(" ", "_").replace("&", "and").replace("-", "_")
            save_text_content(page, f"fights_{safe_name}.txt")

        print("=== Saving tournament pages ===")
        page.goto(stats_url, wait_until="domcontentloaded", timeout=60000)
        page.wait_for_timeout(2000)

        tournament_links = page.query_selector_all('a[href*="/tournament/"]')
        seen = set()
        for link in tournament_links:
            href = link.get_attribute("href") or ""
            if not href or "/nomination/" in href or href in seen:
                continue
            seen.add(href)

            text = link.inner_text().strip()[:30]
            print(f"  Tournament: {text}")

            tpage = context.new_page()
            try:
                tpage.goto(HEMAGON_BASE + href, wait_until="domcontentloaded", timeout=30000)
                tpage.wait_for_timeout(2000)

                slug = href.split("/")[-1]
                save_text_content(tpage, f"tournament_{slug}.txt")
            except Exception as e:
                print(f"    Error: {e}")
            finally:
                tpage.close()

        browser.close()

    print(f"\nAll files saved to: {TEMP_DIR}")
    return TEMP_DIR


def parse_profile_page(text: str) -> dict:
    """Parse profile text."""
    import re

    result = {}
    lines = [l.strip() for l in text.split("\n") if l.strip()]

    for i, line in enumerate(lines):
        if "HEMA TEAM" in line:
            result["club"] = "HEMA TEAM"
        if re.match(r"^[A-Z][a-z]+ [A-Z][a-z]+", line) and "club" not in result:
            if i < 5:
                result["name"] = line

    text_content = text
    if "HEMA TEAM" in text_content:
        result["club"] = "HEMA TEAM"

    country_match = re.search(
        r"\b(Russia|Ukraine|Belarus|Poland|Germany|USA|UK|Italy|France|Sweden)\b",
        text_content,
    )
    if country_match:
        result["country"] = country_match.group(1)

    events_match = re.search(r"EVENTS\s*(\d+)", text_content)
    cats_match = re.search(r"CATEGORIES\s*(\d+)", text_content)
    fights_match = re.search(r"FIGHTS\s*(\d+)", text_content)

    summary = {}
    if events_match:
        summary["events"] = int(events_match.group(1))
    if cats_match:
        summary["categories"] = int(cats_match.group(1))
    if fights_match:
        summary["fights"] = int(fights_match.group(1))
    result["summary"] = summary

    return result


def parse_weapons_from_stats(text: str) -> list:
    """Parse weapons table from stats text."""
    import re

    weapons = []
    lines = text.split("\n")

    in_table = False
    for i, line in enumerate(lines):
        if "WEAPON" in line.upper() and "RATING" in line.upper():
            in_table = True
            continue
        if not in_table:
            continue
        if "EVENTS" not in line and "WIN%" not in line and not re.match(r"^\d", line.strip()):
            continue
        if re.match(r"^\d", line.strip()):
            parts = line.split()
            if len(parts) >= 5:
                weapon = {
                    "name": parts[0],
                    "rank": int(parts[1]) if parts[1].isdigit() else None,
                    "rating": int(parts[2]) if parts[2].isdigit() else None,
                    "events": int(parts[3]) if parts[3].isdigit() else None,
                    "fights": int(parts[4]) if parts[4].isdigit() else None,
                    "win_percent": int(parts[5].replace("%", "")) if len(parts) > 5 and parts[5].replace("%", "").isdigit() else None,
                }
                weapons.append(weapon)

    return weapons


def parse_fights_from_text(text: str, fighter_name: str) -> list:
    """Parse fights from text content."""
    import re

    fights = []
    lines = text.split("\n")

    for i, line in enumerate(lines):
        if fighter_name not in line:
            continue

        parts = line.split()
        if len(parts) < 3:
            continue

        try:
            user_score = int(parts[0])
            opponent_score = int(parts[2])
        except (ValueError, IndexError):
            continue

        opponent = parts[1]
        if not opponent or opponent == fighter_name or len(opponent) < 3:
            continue
        if "Round" in opponent or "Pool" in opponent:
            continue
        if not opponent[0].isupper():
            continue

        result_val = "win" if user_score > opponent_score else "lose"

        extra = ""
        if i + 1 < len(lines):
            extra = lines[i + 1]

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
                "club": None,
                "user_score": user_score,
                "opponent_score": opponent_score,
                "result": result_val,
                "tournament": tournament,
                "stage": stage,
                "date": date,
            }
        )

    return fights


def parse_tournaments_from_stats(text: str) -> list:
    """Parse tournaments from stats text."""
    import re

    tournaments = []
    lines = text.split("\n")

    seen = set()
    for line in lines:
        if "/tournament/" in line:
            match = re.search(r"/tournament/([^\s]+)", line)
            if match:
                slug = match.group(1)
                if slug not in seen and "/nomination/" not in slug:
                    seen.add(slug)
                    name = re.sub(r"[^\w\s]", "", line).strip().split()[0] if line.strip() else slug
                    tournaments.append(
                        {
                            "name": name,
                            "url": f"/tournament/{slug}",
                            "slug": slug,
                            "category": "",
                            "vk_link": None,
                        }
                    )

    return tournaments


def parse_tournament_page(text: str, slug: str) -> dict:
    """Parse tournament page text to get VK link."""
    import re

    result = {"url": f"/tournament/{slug}", "slug": slug, "vk_link": None}

    vk_match = re.search(r"vk\.com/(\w+)", text)
    if vk_match:
        result["vk_link"] = f"https://vk.com/{vk_match.group(1)}"

    lines = [l.strip() for l in text.split("\n") if l.strip()]
    if lines:
        result["name"] = lines[0]

    return result


def parse_saved_files(temp_dir: str) -> dict:
    """Parse all saved text files and build result JSON."""
    result = {"fights": [], "tournaments": []}

    profile_path = os.path.join(temp_dir, "profile.txt")
    if os.path.exists(profile_path):
        with open(profile_path, "r", encoding="utf-8") as f:
            text = f.read()
        profile_data = parse_profile_page(text)
        result.update(profile_data)

    stats_path = os.path.join(temp_dir, "stats.txt")
    if os.path.exists(stats_path):
        with open(stats_path, "r", encoding="utf-8") as f:
            text = f.read()
        result["weapons"] = parse_weapons_from_stats(text)

        tournament_data = parse_tournaments_from_stats(text)
        for t in tournament_data:
            result["tournaments"].append(t)

    fighter_name = result.get("name", "")

    for weapon_name in WEAPONS:
        safe_name = weapon_name.replace(" ", "_").replace("&", "and").replace("-", "_")
        fights_path = os.path.join(temp_dir, f"fights_{safe_name}.txt")

        if os.path.exists(fights_path):
            with open(fights_path, "r", encoding="utf-8") as f:
                text = f.read()
            fights = parse_fights_from_text(text, fighter_name)
            print(f"  {weapon_name}: {len(fights)} fights")
            result["fights"].extend(fights)

    for t in result.get("tournaments", []):
        slug = t.get("slug", "")
        if not slug:
            continue

        tourney_path = os.path.join(temp_dir, f"tournament_{slug}.txt")
        if os.path.exists(tourney_path):
            with open(tourney_path, "r", encoding="utf-8") as f:
                text = f.read()

            import re

            vk_match = re.search(r"(https?://vk\.com/[^\s]+)", text)
            if vk_match:
                t["vk_link"] = vk_match.group(1)

            cat_match = re.search(r"category.*?(\w+)", text)
            if cat_match:
                t["category"] = cat_match.group(1)

    return result


def scrape_profile(profile_url: str) -> dict:
    """
    Scrape a fighter's profile from hemagon.com
    """
    print(f"Scraping: {profile_url}")
    print(f"Temp dir: {TEMP_DIR}")

    temp_dir = scrape_profile_to_files(profile_url)

    result = parse_saved_files(temp_dir)

    return result


if __name__ == "__main__":
    import sys

    url = sys.argv[1] if len(sys.argv) > 1 else "https://hemagon.com/users/eugeniya.shumakova"
    result = scrape_profile(url)
    print(json.dumps(result, indent=2, ensure_ascii=False))