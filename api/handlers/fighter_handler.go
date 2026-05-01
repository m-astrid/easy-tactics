package handlers

import (
	"context"
	"strings"

	"github.com/easy-tactics/api/domain"
	"github.com/easy-tactics/api/storage"
)

type FighterHandler struct {
	fighterStore *storage.FighterStorage
}

func NewFighterHandler(fighterStore *storage.FighterStorage) *FighterHandler {
	return &FighterHandler{fighterStore: fighterStore}
}

type SearchFighterRequest struct {
	Query         string `json:"query"`
	SearchHemagon bool   `json:"search_hemagon"`
}

type SearchFighterResponse struct {
	Source   Source          `json:"source"`
	Matches  []*FighterMatch `json:"matches"`
	Selected *FighterMatch   `json:"selected"`
}

type Source int

const (
	Source_SOURCE_UNKNOWN Source = iota
	Source_SOURCE_LOCAL
	Source_SOURCE_HEMAGON
	Source_SOURCE_NOT_FOUND
)

type FighterMatch struct {
	Uuid        string   `json:"uuid"`
	Slug        string   `json:"slug"`
	FullName    string   `json:"full_name"`
	City        string   `json:"city"`
	Club        string   `json:"club"`
	HemagonURL  string   `json:"hemagon_url"`
	Photos      []string `json:"photos"`
	LastUpdated string   `json:"last_updated"`
	ExistsInDB  bool     `json:"exists_in_db"`
}

type GetFighterRequest struct {
	Uuid string `json:"uuid"`
	Slug string `json:"slug"`
}

type FighterResponse struct {
	Uuid          string           `json:"uuid"`
	Slug          string           `json:"slug"`
	FullName      string           `json:"full_name"`
	City          string           `json:"city"`
	Club          string           `json:"club"`
	HemagonURL    string           `json:"hemagon_url"`
	Photos        []*FighterPhoto  `json:"photos"`
	Tags          []string         `json:"tags"`
	CreatedAt     string           `json:"created_at"`
	UpdatedAt     string           `json:"updated_at"`
	LatestSummary *SummaryResponse `json:"latest_summary"`
}

type FighterPhoto struct {
	URL     string `json:"url"`
	Source  string `json:"source"`
	TakenAt string `json:"taken_at"`
}

type SummaryResponse struct {
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
	Version   int32  `json:"version"`
}

type CreateFighterRequest struct {
	FullName   string   `json:"full_name"`
	City       string   `json:"city"`
	Club       string   `json:"club"`
	HemagonURL string   `json:"hemagon_url"`
	Photos     []string `json:"photos"`
}

type UpdateFighterRequest struct {
	Uuid        string   `json:"uuid"`
	FullName    string   `json:"full_name"`
	City        string   `json:"city"`
	Club        string   `json:"club"`
	HemagonURL  string   `json:"hemagon_url"`
	PhotosToAdd []string `json:"photos_to_add"`
	TagsToAdd   []string `json:"tags_to_add"`
}

type ListFightersRequest struct {
	Limit     int32  `json:"limit"`
	Offset    int32  `json:"offset"`
	TagFilter string `json:"tag_filter"`
}

type ListFightersResponse struct {
	Fighters []*FighterResponse `json:"fighters"`
	Total    int32              `json:"total"`
}

func (h *FighterHandler) SearchFighter(ctx context.Context, req *SearchFighterRequest) (*SearchFighterResponse, error) {
	if req.Query == "" {
		return &SearchFighterResponse{
			Source:  Source_SOURCE_NOT_FOUND,
			Matches: []*FighterMatch{},
		}, nil
	}

	fighters, err := h.fighterStore.Search(req.Query)
	if err != nil {
		return nil, err
	}

	if len(fighters) == 0 {
		if req.SearchHemagon {
			return &SearchFighterResponse{
				Source:  Source_SOURCE_HEMAGON,
				Matches: []*FighterMatch{},
			}, nil
		}
		return &SearchFighterResponse{
			Source:  Source_SOURCE_NOT_FOUND,
			Matches: []*FighterMatch{},
		}, nil
	}

	matches := make([]*FighterMatch, len(fighters))
	for i, f := range fighters {
		matches[i] = &FighterMatch{
			Uuid:       f.UUID,
			Slug:       f.Slug,
			FullName:   f.FullName,
			City:       f.City,
			Club:       f.Club,
			HemagonURL: f.HemagonURL,
			ExistsInDB: true,
		}
	}

	return &SearchFighterResponse{
		Source:  Source_SOURCE_LOCAL,
		Matches: matches,
	}, nil
}

func (h *FighterHandler) GetFighter(ctx context.Context, req *GetFighterRequest) (*FighterResponse, error) {
	var fighter *domain.Fighter
	var err error

	if req.Uuid != "" {
		fighter, err = h.fighterStore.GetByUUID(req.Uuid)
	} else if req.Slug != "" {
		fighter, err = h.fighterStore.GetBySlug(req.Slug)
	}

	if err != nil {
		return nil, err
	}

	if fighter == nil {
		return nil, nil
	}

	return &FighterResponse{
		Uuid:       fighter.UUID,
		Slug:       fighter.Slug,
		FullName:   fighter.FullName,
		City:       fighter.City,
		Club:       fighter.Club,
		HemagonURL: fighter.HemagonURL,
	}, nil
}

func (h *FighterHandler) CreateFighter(ctx context.Context, req *CreateFighterRequest) (*FighterResponse, error) {
	uuid := generateUUID()
	slug := generateSlug(req.FullName, req.City)

	fighter := &domain.Fighter{
		UUID:       uuid,
		Slug:       slug,
		FullName:   req.FullName,
		City:       req.City,
		Club:       req.Club,
		HemagonURL: req.HemagonURL,
	}

	if err := h.fighterStore.Create(fighter); err != nil {
		return nil, err
	}

	return &FighterResponse{
		Uuid:       fighter.UUID,
		Slug:       fighter.Slug,
		FullName:   fighter.FullName,
		City:       fighter.City,
		Club:       fighter.Club,
		HemagonURL: fighter.HemagonURL,
	}, nil
}

func (h *FighterHandler) UpdateFighter(ctx context.Context, req *UpdateFighterRequest) (*FighterResponse, error) {
	fighter, err := h.fighterStore.GetByUUID(req.Uuid)
	if err != nil {
		return nil, err
	}

	if fighter == nil {
		return nil, nil
	}

	if req.FullName != "" {
		fighter.FullName = req.FullName
	}
	if req.City != "" {
		fighter.City = req.City
	}
	if req.Club != "" {
		fighter.Club = req.Club
	}
	if req.HemagonURL != "" {
		fighter.HemagonURL = req.HemagonURL
	}

	if err := h.fighterStore.Update(fighter); err != nil {
		return nil, err
	}

	return &FighterResponse{
		Uuid:       fighter.UUID,
		Slug:       fighter.Slug,
		FullName:   fighter.FullName,
		City:       fighter.City,
		Club:       fighter.Club,
		HemagonURL: fighter.HemagonURL,
	}, nil
}

func (h *FighterHandler) ListFighters(ctx context.Context, req *ListFightersRequest) (*ListFightersResponse, error) {
	fighters, err := h.fighterStore.List()
	if err != nil {
		return nil, err
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}

	if offset > len(fighters) {
		offset = len(fighters)
	}

	end := offset + limit
	if end > len(fighters) {
		end = len(fighters)
	}

	result := fighters[offset:end]

	responses := make([]*FighterResponse, len(result))
	for i, f := range result {
		responses[i] = &FighterResponse{
			Uuid:       f.UUID,
			Slug:       f.Slug,
			FullName:   f.FullName,
			City:       f.City,
			Club:       f.Club,
			HemagonURL: f.HemagonURL,
		}
	}

	return &ListFightersResponse{
		Fighters: responses,
		Total:    int32(len(fighters)),
	}, nil
}

func generateUUID() string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
}

func generateSlug(fullName, city string) string {
	namePart := strings.ToLower(strings.ReplaceAll(fullName, " ", "-"))
	if city != "" {
		cityPart := strings.ToLower(strings.ReplaceAll(city, " ", "-"))
		return namePart + "-" + cityPart
	}
	return namePart
}
