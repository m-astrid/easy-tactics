import os
import re
import json

script_dir = os.path.dirname(os.path.abspath(__file__))
data_dir = os.path.join(script_dir, "hemagon_data")

profile_text = open(os.path.join(data_dir, "profile.txt")).read()
stats_text = open(os.path.join(data_dir, "stats_fights_50.txt")).read()
lines = [l for l in profile_text.split('\n') if l.strip()]
stats_lines = [l for l in stats_text.split('\n') if l.strip()]

result = {}

name = None
for line in lines:
    if line and line[0].isupper() and not any(x in line for x in ['HEMAGON', 'MENU', 'SIGN IN', 'CIS', 'Golden', 'UHFC', 'World', 'CSEN', 'Danish', 'HHF', 'Nordic', 'En', 'Ru']):
        name_candidate = line.strip()
        if len(name_candidate) > 2 and len(name_candidate) < 50:
            name = name_candidate
            break

result['name'] = name

for i, line in enumerate(lines):
    if line.strip() == 'HEMA TEAM':
        result['club'] = 'HEMA TEAM'
        break

if 'Russia' in profile_text:
    result['country'] = 'Russia'

location = None
for i, line in enumerate(lines):
    if line.strip() == '/':
        if i + 1 < len(lines):
            location = lines[i + 1].strip()
            break

result['location'] = location

summary = {}
events_match = re.search(r'EVENTS\s+(\d+)', profile_text)
if events_match:
    summary['events'] = int(events_match.group(1))

cats_match = re.search(r'CATEGORIES\s+(\d+)', profile_text)
if cats_match:
    summary['categories'] = int(cats_match.group(1))

fights_match = re.search(r'FIGHTS\s+(\d+)', profile_text)
if fights_match:
    summary['fights'] = int(fights_match.group(1))

result['summary'] = summary

weapons = []
weapon_line = None
for line in lines:
    if 'Longsword' in line and '\t' in line:
        weapon_line = line
        break

if weapon_line:
    parts = weapon_line.split('\t')
    if len(parts) >= 6:
        weapon = {
            'name': parts[0].strip(),
            'rank': int(parts[1].strip()) if parts[1].strip().isdigit() else None,
            'rating': int(parts[2].strip()) if parts[2].strip().isdigit() else None,
            'events': int(parts[3].strip()) if parts[3].strip().isdigit() else None,
            'fights': int(parts[4].strip()) if parts[4].strip().isdigit() else None,
            'win_percent': int(parts[5].strip().replace('%', '')) if parts[5].strip().replace('%', '').isdigit() else None
        }
        weapons.append(weapon)

result['weapons'] = weapons

fights = []
date_pattern = re.compile(r'([A-Za-z]{3}\s+\d{1,2},\s+\d{4})')

for i, line in enumerate(stats_lines):
    if date_pattern.search(line) and line.startswith('\t'):
        parts = line.split('\t')
        if len(parts) >= 5:
            date = parts[-1].strip()
            tournament = parts[1].strip() if parts[1].strip() else ''
            
            stage = ''
            if len(parts) >= 4:
                stage_candidate = parts[3].strip()
                if 'Round' in stage_candidate or 'Pool' in stage_candidate:
                    stage = stage_candidate
            
            if i >= 6:
                try:
                    opponent_score = int(stats_lines[i - 1].strip())
                except:
                    continue
                
                opp_club_idx = i - 2
                opp_name_idx = i - 3
                user_score_idx = i - 4
                
                opponent_club = stats_lines[opp_club_idx].strip() if opp_club_idx >= 0 else ''
                opponent_name = stats_lines[opp_name_idx].strip() if opp_name_idx >= 0 else ''
                user_score_str = stats_lines[user_score_idx].strip() if user_score_idx >= 0 else ''
                
                if opponent_name == name:
                    opp_name_idx2 = i - 6
                    opponent_name = stats_lines[opp_name_idx2].strip() if opp_name_idx2 >= 0 else ''
                    opp_club_idx2 = i - 5
                    opponent_club = stats_lines[opp_club_idx2].strip() if opp_club_idx2 >= 0 else ''
                
                if not opponent_name or len(opponent_name) < 2:
                    continue
                
                try:
                    user_score = int(user_score_str)
                except:
                    continue
                
                result_val = 'win' if user_score > opponent_score else 'lose' if user_score < opponent_score else 'draw'
                
                fights.append({
                    'opponent': opponent_name,
                    'club': opponent_club if opponent_club and not opponent_club[0].isupper() else None,
                    'user_score': user_score,
                    'opponent_score': opponent_score,
                    'result': result_val,
                    'tournament': tournament,
                    'stage': stage,
                    'date': date
                })

result['fights'] = fights

tournaments = []
for filename in os.listdir(data_dir):
    if filename.startswith('tournament_') and filename.endswith('.json'):
        with open(os.path.join(data_dir, filename)) as f:
            data = json.load(f)
            text = data.get('text', '')
            
            tourney_name = None
            for line in text.split('\n'):
                if line and not any(x in line for x in ['CIS', 'Golden Falcon', 'UHFC', 'World', 'CSEN', 'Danish', 'HHF', 'Nordic', 'En', 'Ru', 'HEMAGON', 'MENU', 'SIGN IN', 'LOCATION', 'DATE', 'REGISTRATION', 'Tournament', 'Categories', 'Applications', 'Summary']):
                    if len(line) > 3 and len(line) < 100:
                        tourney_name = line.strip()
                        break
            
            slug = filename.replace('tournament_', '').replace('.json', '')
            
            category = None
            for link in data.get('links', []):
                if '/categories' in link.get('href', ''):
                    category = link.get('text', '')
                    break
            
            if tourney_name:
                tournaments.append({
                    'name': tourney_name,
                    'url': f'/tournament/{slug}',
                    'slug': slug,
                    'category': category
                })

result['tournaments'] = tournaments

print(json.dumps(result, indent=2, ensure_ascii=False))