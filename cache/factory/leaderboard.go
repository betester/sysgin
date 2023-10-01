package factory

import (
	"context"

	"github.com/sysygn/cache/db"
	"github.com/sysygn/cache/leaderboard"
)
type LeaderboardDatabaseFactory struct {
	Ctx   context.Context
	Table string
}

func (ldf *LeaderboardDatabaseFactory) CreateDb(dbType string) leaderboard.LeaderboardRepository {
	var database leaderboard.LeaderboardRepository
	if dbType == "postgres" {
		pgDb := db.ConnectPg()
		database = &leaderboard.PostgresLeaderboardRepository{Client: pgDb, Table: ldf.Table}
		err := database.(*leaderboard.PostgresLeaderboardRepository).DropTable()
		if err != nil {
			panic(err.Error())
		}
		createErr := database.(*leaderboard.PostgresLeaderboardRepository).CreateTable()
		if createErr != nil {
			panic(createErr.Error())
		}
		createIndexErr := database.(*leaderboard.PostgresLeaderboardRepository).CreateIndex()

		if createIndexErr != nil {
			panic(createIndexErr.Error())
		}
	} else if dbType == "redis" {
		ctx := context.Background()
		database = &leaderboard.RedisLeaderboardRepository{Ctx: ctx, Client: db.ConnectRedis(), Key: ldf.Table}
	}

	return database
}