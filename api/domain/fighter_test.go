package domain

import "testing"

func TestFighter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fighter Fighter
		wantErr bool
	}{
		{
			name: "valid fighter with all fields",
			fighter: Fighter{
				UUID:       "abc-123",
				Slug:       "john-doe",
				FullName:   "John Doe",
				HemagonURL: "https://hemagon.com/fencer/abc-123",
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			fighter: Fighter{
				Slug:     "john-doe",
				FullName: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "missing FullName",
			fighter: Fighter{
				UUID: "abc-123",
				Slug: "john-doe",
			},
			wantErr: true,
		},
		{
			name: "missing Slug",
			fighter: Fighter{
				UUID:     "abc-123",
				FullName: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "empty UUID",
			fighter: Fighter{
				UUID:     "",
				Slug:     "john-doe",
				FullName: "John Doe",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fighter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Fighter.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTournament_Validate(t *testing.T) {
	tests := []struct {
		name       string
		tournament Tournament
		wantErr    bool
	}{
		{
			name: "valid tournament with all fields",
			tournament: Tournament{
				UUID:        "tourn-123",
				FighterUUID: "fighter-456",
				Name:        "Spring Championship 2024",
				City:        "Prague",
				Country:     "Czech Republic",
				StartDate:   "2024-03-15",
				HemagonURL:  "https://hemagon.com/tournament/tourn-123",
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			tournament: Tournament{
				FighterUUID: "fighter-456",
				Name:        "Spring Championship 2024",
			},
			wantErr: true,
		},
		{
			name: "missing Name",
			tournament: Tournament{
				UUID:        "tourn-123",
				FighterUUID: "fighter-456",
			},
			wantErr: true,
		},
		{
			name: "missing FighterUUID",
			tournament: Tournament{
				UUID: "tourn-123",
				Name: "Spring Championship 2024",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tournament.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Tournament.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFight_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fight   Fight
		wantErr bool
	}{
		{
			name: "valid fight with all fields",
			fight: Fight{
				UUID:           "fight-123",
				FighterUUID:    "fighter-456",
				TournamentUUID: "tourn-789",
				OpponentUUID:   "opponent-111",
				OpponentName:   "Jane Smith",
				ScoreWin:       15,
				ScoreLose:      12,
				Round:          "Quarter Finals",
				FightDate:      "2024-03-16",
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			fight: Fight{
				FighterUUID:    "fighter-456",
				TournamentUUID: "tourn-789",
				OpponentUUID:   "opponent-111",
				OpponentName:   "Jane Smith",
			},
			wantErr: true,
		},
		{
			name: "missing FighterUUID",
			fight: Fight{
				UUID:           "fight-123",
				TournamentUUID: "tourn-789",
				OpponentUUID:   "opponent-111",
				OpponentName:   "Jane Smith",
			},
			wantErr: true,
		},
		{
			name: "missing TournamentUUID",
			fight: Fight{
				UUID:         "fight-123",
				FighterUUID:  "fighter-456",
				OpponentUUID: "opponent-111",
				OpponentName: "Jane Smith",
			},
			wantErr: true,
		},
		{
			name: "missing OpponentName",
			fight: Fight{
				UUID:           "fight-123",
				FighterUUID:    "fighter-456",
				TournamentUUID: "tourn-789",
				OpponentUUID:   "opponent-111",
			},
			wantErr: true,
		},
		{
			name: "valid fight with zero scores",
			fight: Fight{
				UUID:           "fight-123",
				FighterUUID:    "fighter-456",
				TournamentUUID: "tourn-789",
				OpponentUUID:   "opponent-111",
				OpponentName:   "Jane Smith",
				ScoreWin:       0,
				ScoreLose:      0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fight.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Fight.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
