package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	ctx := context.Background()
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?pool_max_conns=10",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		getEnvDefault("DB_PORT", "5432"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = pgxpool.New(ctx, connStr)
	if err != nil {
		LogError("PostgreSQL Connection Error: " + err.Error())
		os.Exit(1)
	}

	if err := DB.Ping(ctx); err != nil {
		LogError("PostgreSQL Ping Error: " + err.Error())
		os.Exit(1)
	}

	LogSuccess("PostgreSQL Database connected!")
}

func getEnvDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

type Account struct {
	Pid            int64  `json:"pid"`
	UserId         string `json:"user_id"`
	ScreenName     string `json:"screen_name"`
	Country        string `json:"country"`
	ProfileComment string `json:"profile_comment"`
	GameSkill      int    `json:"game_skill"`
}

type Community struct {
	CommunityId      int64  `json:"community_id"`
	OliveTitleId     string `json:"olive_title_id"`
	OliveCommunityId string `json:"olive_community_id"`
	Name             string `json:"name"`
	Icon             string `json:"icon"`
	Category         string `json:"category"`
	SortOrder        int    `json:"sort_order"`
	Hidden           bool   `json:"hidden"`
	FavoriteCount    int    `json:"favorite_count"`
}

func GetUserById(userId string) (*Account, error) {
	ctx := context.Background()
	query := `SELECT pid, user_id, screen_name, country, profile_comment, game_skill FROM accounts WHERE user_id = $1`

	var acc Account
	err := DB.QueryRow(ctx, query, userId).Scan(
		&acc.Pid, &acc.UserId, &acc.ScreenName, &acc.Country, &acc.ProfileComment, &acc.GameSkill,
	)
	if err != nil {
		return nil, nil
	}
	return &acc, nil
}

func GetAllCommunities() ([]Community, error) {
	ctx := context.Background()
	query := `
		SELECT c.*, 
		(SELECT COUNT(*) FROM community_favorites f WHERE f.community_id = c.community_id)::int AS favorite_count
		FROM communities c 
		WHERE (c.hidden IS NULL OR c.hidden = false) 
		ORDER BY c.sort_order, c.name
	`
	rows, err := DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var communities []Community
	for rows.Next() {
		var c Community
		if err := rows.Scan(
			&c.CommunityId, &c.OliveTitleId, &c.OliveCommunityId, &c.Name,
			&c.Icon, &c.Category, &c.SortOrder, &c.Hidden, &c.FavoriteCount,
		); err != nil {
			continue
		}
		communities = append(communities, c)
	}
	return communities, nil
}

func GetCommunityById(oliveCommunityId string) (*Community, error) {
	ctx := context.Background()
	query := `
		SELECT c.*, 
		(SELECT COUNT(*) FROM community_favorites f WHERE f.community_id = c.community_id)::int AS favorite_count
		FROM communities c 
		WHERE c.olive_community_id = $1
	`
	var c Community
	err := DB.QueryRow(ctx, query, oliveCommunityId).Scan(
		&c.CommunityId, &c.OliveTitleId, &c.OliveCommunityId, &c.Name,
		&c.Icon, &c.Category, &c.SortOrder, &c.Hidden, &c.FavoriteCount,
	)
	if err != nil {
		return nil, nil
	}
	return &c, nil
}
