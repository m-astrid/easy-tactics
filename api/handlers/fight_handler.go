package handlers

import (
	"context"

	"github.com/easy-tactics/api/domain"
	"github.com/easy-tactics/api/storage"
)

type FightHandler struct {
	fighterStore *storage.FighterStorage
}

func NewFightHandler(fighterStore *storage.FighterStorage) *FightHandler {
	return &FightHandler{fighterStore: fighterStore}
}

type GetFighterFightsRequest struct {
	FighterUUID    string `json:"fighter_uuid"`
	TournamentUUID string `json:"tournament_uuid"`
	Limit          int32  `json:"limit"`
	Offset         int32  `json:"offset"`
}

type GetFightRequest struct {
	UUID string `json:"uuid"`
}

type FightResponse struct {
	UUID           string `json:"uuid"`
	FighterUUID    string `json:"fighter_uuid"`
	TournamentUUID string `json:"tournament_uuid"`
	OpponentUUID   string `json:"opponent_uuid"`
	OpponentName   string `json:"opponent_name"`
	ScoreWin       int32  `json:"score_win"`
	ScoreLose      int32  `json:"score_lose"`
	Round          string `json:"round"`
	FightDate      string `json:"fight_date"`
}

type FightListResponse struct {
	Fights []*FightResponse `json:"fights"`
	Total  int32            `json:"total"`
}

func (h *FightHandler) GetFighterFights(ctx context.Context, req *GetFighterFightsRequest) (*FightListResponse, error) {
	var fights []*domain.Fight
	var err error

	if req.TournamentUUID != "" {
		fights, err = h.fighterStore.GetFightsByTournamentUUID(req.TournamentUUID)
		if err != nil {
			return nil, err
		}
		var filtered []*domain.Fight
		for _, f := range fights {
			if f.FighterUUID == req.FighterUUID {
				filtered = append(filtered, f)
			}
		}
		fights = filtered
	} else {
		fights, err = h.fighterStore.GetFightsByFighterUUID(req.FighterUUID)
		if err != nil {
			return nil, err
		}
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}

	if offset > len(fights) {
		offset = len(fights)
	}

	end := offset + limit
	if end > len(fights) {
		end = len(fights)
	}

	result := fights[offset:end]

	responses := make([]*FightResponse, len(result))
	for i, f := range result {
		responses[i] = &FightResponse{
			UUID:           f.UUID,
			FighterUUID:    f.FighterUUID,
			TournamentUUID: f.TournamentUUID,
			OpponentUUID:   f.OpponentUUID,
			OpponentName:   f.OpponentName,
			ScoreWin:       int32(f.ScoreWin),
			ScoreLose:      int32(f.ScoreLose),
			Round:          f.Round,
			FightDate:      f.FightDate,
		}
	}

	return &FightListResponse{
		Fights: responses,
		Total:  int32(len(fights)),
	}, nil
}

func (h *FightHandler) GetFight(ctx context.Context, uuid string) (*FightResponse, error) {
	fights, err := h.fighterStore.GetAllFights()
	if err != nil {
		return nil, err
	}

	for _, f := range fights {
		if f.UUID == uuid {
			return &FightResponse{
				UUID:           f.UUID,
				FighterUUID:    f.FighterUUID,
				TournamentUUID: f.TournamentUUID,
				OpponentUUID:   f.OpponentUUID,
				OpponentName:   f.OpponentName,
				ScoreWin:       int32(f.ScoreWin),
				ScoreLose:      int32(f.ScoreLose),
				Round:          f.Round,
				FightDate:      f.FightDate,
			}, nil
		}
	}

	return nil, nil
}
