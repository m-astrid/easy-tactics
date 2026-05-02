package ai

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
)

type Client struct {
	conn *grpc.ClientConn
	ai   fapiv1.AIServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		ai:   fapiv1.NewAIServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) AnalyzeFight(ctx context.Context, fighterUUID, fightUUID, videoURL, opponentName string, scoreWin, scoreLose int32, round string) (*fapiv1.AnalysisResult, error) {
	req := &fapiv1.AnalyzeFightRequest{
		FighterUuid:  fighterUUID,
		FightUuid:    fightUUID,
		VideoUrl:     videoURL,
		OpponentName: opponentName,
		ScoreWin:     scoreWin,
		ScoreLose:    scoreLose,
		Round:        round,
	}
	return c.ai.AnalyzeFight(ctx, req)
}

func (c *Client) GenerateSummary(ctx context.Context, fighterUUID string, fightUUIDs []string, includeTournaments bool) (*fapiv1.SummaryResult, error) {
	req := &fapiv1.GenerateSummaryRequest{
		FighterUuid:        fighterUUID,
		FightUuids:         fightUUIDs,
		IncludeTournaments: includeTournaments,
	}
	return c.ai.GenerateSummary(ctx, req)
}

func (c *Client) SearchVideos(ctx context.Context, fighterName, tournament string, fightUUIDs []string, searchVK, searchYouTube bool) (*fapiv1.VideosResult, error) {
	req := &fapiv1.SearchVideosRequest{
		FighterName:   fighterName,
		Tournament:    tournament,
		FightUuids:    fightUUIDs,
		SearchVk:      searchVK,
		SearchYoutube: searchYouTube,
	}
	return c.ai.SearchVideos(ctx, req)
}

func (c *Client) ParseHemagon(ctx context.Context, fighterSlug string, includeTournaments, includeFights bool) (*fapiv1.ParseResult, error) {
	req := &fapiv1.ParseHemagonRequest{
		FighterSlug:        fighterSlug,
		IncludeTournaments: includeTournaments,
		IncludeFights:      includeFights,
	}
	return c.ai.ParseHemagon(ctx, req)
}
