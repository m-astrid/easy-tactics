package handlers

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	fapiv1 "github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
)

type APIClient struct {
	conn          *grpc.ClientConn
	authClient    fapiv1.AuthServiceClient
	fighterClient fapiv1.FighterServiceClient
}

func NewAPIClient(addr string) (*APIClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &APIClient{
		conn:          conn,
		authClient:    fapiv1.NewAuthServiceClient(conn),
		fighterClient: fapiv1.NewFighterServiceClient(conn),
	}, nil
}

func (c *APIClient) Close() error {
	return c.conn.Close()
}

func (c *APIClient) CheckAccess(ctx context.Context, telegramID int64, requiredRole string) (bool, string, error) {
	resp, err := c.authClient.CheckAccess(ctx, &fapiv1.CheckAccessRequest{
		TelegramId:   telegramID,
		RequiredRole: requiredRole,
	})
	if err != nil {
		return false, "", err
	}
	return resp.Allowed, resp.Reason, nil
}

func (c *APIClient) GetUser(ctx context.Context, telegramID int64) (*UserInfo, error) {
	resp, err := c.authClient.GetUser(ctx, &fapiv1.GetUserRequest{
		TelegramId: telegramID,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	return &UserInfo{
		TelegramID: resp.TelegramId,
		Username:   resp.Username,
		FullName:   resp.FullName,
		Role:       resp.Role,
	}, nil
}

func (c *APIClient) AddUser(ctx context.Context, telegramID int64, username, fullName, role string) error {
	_, err := c.authClient.AddUser(ctx, &fapiv1.AddUserRequest{
		TelegramId: telegramID,
		Username:   username,
		FullName:   fullName,
		Role:       role,
	})
	return err
}

func (c *APIClient) ListUsers(ctx context.Context) ([]*UserInfo, error) {
	resp, err := c.authClient.ListUsers(ctx, &fapiv1.ListUsersRequest{
		Limit: 100,
	})
	if err != nil {
		return nil, err
	}

	users := make([]*UserInfo, len(resp.Users))
	for i, u := range resp.Users {
		users[i] = &UserInfo{
			TelegramID: u.TelegramId,
			Username:   u.Username,
			FullName:   u.FullName,
			Role:       u.Role,
		}
	}
	return users, nil
}

func (c *APIClient) SearchFighter(ctx context.Context, query string) (*SearchResult, error) {
	resp, err := c.fighterClient.SearchFighter(ctx, &fapiv1.SearchFighterRequest{
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	matches := make([]*FighterMatch, len(resp.Matches))
	for i, m := range resp.Matches {
		matches[i] = &FighterMatch{
			UUID:     m.Uuid,
			Slug:     m.Slug,
			FullName: m.FullName,
			City:     m.City,
			Club:     m.Club,
		}
	}

	return &SearchResult{
		Source:  resp.Source.String(),
		Matches: matches,
	}, nil
}

func (c *APIClient) GetFighterBySlug(ctx context.Context, slug string) (*FighterInfo, error) {
	resp, err := c.fighterClient.GetFighter(ctx, &fapiv1.GetFighterRequest{
		Id: &fapiv1.GetFighterRequest_Slug{Slug: slug},
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	return &FighterInfo{
		UUID:       resp.Uuid,
		Slug:       resp.Slug,
		FullName:   resp.FullName,
		City:       resp.City,
		Club:       resp.Club,
		HemagonURL: resp.HemagonUrl,
	}, nil
}

func (c *APIClient) GetTournaments(ctx context.Context, fighterUUID string) ([]*TournamentInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *APIClient) GetFights(ctx context.Context, fighterUUID string) ([]*FightInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

type UserInfo struct {
	TelegramID int64
	Username   string
	FullName   string
	Role       string
}

type FighterMatch struct {
	UUID     string
	Slug     string
	FullName string
	City     string
	Club     string
}

type SearchResult struct {
	Source  string
	Matches []*FighterMatch
}

type FighterInfo struct {
	UUID       string
	Slug       string
	FullName   string
	City       string
	Club       string
	HemagonURL string
}

type TournamentInfo struct {
	UUID      string
	Name      string
	City      string
	Country   string
	StartDate string
}

type FightInfo struct {
	UUID         string
	OpponentName string
	ScoreWin     int32
	ScoreLose    int32
	Round        string
}
