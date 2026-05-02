package handlers

import (
	"context"

	"github.com/easy-tactics/api/storage"
)

type TournamentHandler struct {
	fighterStore *storage.FighterStorage
}

func NewTournamentHandler(fighterStore *storage.FighterStorage) *TournamentHandler {
	return &TournamentHandler{fighterStore: fighterStore}
}

type GetFighterTournamentsRequest struct {
	FighterUUID string `json:"fighter_uuid"`
	Limit       int32  `json:"limit"`
	Offset      int32  `json:"offset"`
}

type TournamentResponse struct {
	UUID        string `json:"uuid"`
	FighterUUID string `json:"fighter_uuid"`
	Name        string `json:"name"`
	City        string `json:"city"`
	Country     string `json:"country"`
	StartDate   string `json:"start_date"`
	HemagonURL  string `json:"hemagon_url"`
}

type TournamentListResponse struct {
	Tournaments []*TournamentResponse `json:"tournaments"`
	Total       int32                 `json:"total"`
}

func (h *TournamentHandler) GetFighterTournaments(ctx context.Context, req *GetFighterTournamentsRequest) (*TournamentListResponse, error) {
	tournaments, err := h.fighterStore.GetTournamentsByFighterUUID(req.FighterUUID)
	if err != nil {
		return nil, err
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}

	if offset > len(tournaments) {
		offset = len(tournaments)
	}

	end := offset + limit
	if end > len(tournaments) {
		end = len(tournaments)
	}

	result := tournaments[offset:end]

	responses := make([]*TournamentResponse, len(result))
	for i, t := range result {
		responses[i] = &TournamentResponse{
			UUID:        t.UUID,
			FighterUUID: t.FighterUUID,
			Name:        t.Name,
			City:        t.City,
			Country:     t.Country,
			StartDate:   t.StartDate,
			HemagonURL:  t.HemagonURL,
		}
	}

	return &TournamentListResponse{
		Tournaments: responses,
		Total:       int32(len(tournaments)),
	}, nil
}

func (h *TournamentHandler) GetTournament(ctx context.Context, uuid string) (*TournamentResponse, error) {
	tournaments, err := h.fighterStore.GetAllTournaments()
	if err != nil {
		return nil, err
	}

	for _, t := range tournaments {
		if t.UUID == uuid {
			return &TournamentResponse{
				UUID:        t.UUID,
				FighterUUID: t.FighterUUID,
				Name:        t.Name,
				City:        t.City,
				Country:     t.Country,
				StartDate:   t.StartDate,
				HemagonURL:  t.HemagonURL,
			}, nil
		}
	}

	return nil, nil
}
