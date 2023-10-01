package leaderboard

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type LeaderboardRepository interface {
	Insert(data UserLeaderboardData) error
	GetTopK(k int64) ([]UserLeaderboardData, error)
}

type RedisLeaderboardRepository struct {
	Client *redis.Client
	Ctx    context.Context
	Key    string
}

type PostgresLeaderboardRepository struct {
	Client *sql.DB
	Table  string
}

func (plr *PostgresLeaderboardRepository) DropTable() error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", plr.Table)
	_, err := plr.Client.Exec(query)

	return err
}

func (plr *PostgresLeaderboardRepository) CreateTable() error {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(ID UUID PRIMARY KEY, SCORE INTEGER)", plr.Table)
	_, err := plr.Client.Exec(query)

	return err
}

func (plr *PostgresLeaderboardRepository) CreateIndex() error {
		query := fmt.Sprintf("CREATE INDEX ON %s USING BTREE (SCORE)", plr.Table)
	_, err := plr.Client.Exec(query)

	return err
}
func (plr *PostgresLeaderboardRepository) exist(id string) (bool, error) {
	var exist bool

	if err := plr.Client.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE ID='%s')", plr.Table, id)).Scan(&exist); err != nil {
		return false, err
	}
	return exist, nil
}

func (rlr *RedisLeaderboardRepository) Insert(data UserLeaderboardData) error {
	_, err := rlr.Client.ZAdd(rlr.Ctx, rlr.Key, &redis.Z{Score: float64(data.Score), Member: data.Id}).Result()
	return err
}

func (plr *PostgresLeaderboardRepository) Insert(data UserLeaderboardData) error {

	exist, err := plr.exist(data.Id)
	if err != nil {
		return err
	}

	if !exist {
		row, err := plr.Client.Query(fmt.Sprintf("INSERT INTO %s VALUES('%s', %f)", plr.Table, data.Id, data.Score))
		row.Close()
		return err
	} else {
		row, err := plr.Client.Query(fmt.Sprintf("UPDATE %s SET SCORE = %f WHERE ID = '%s'", plr.Table, data.Score, data.Id))
		row.Close()
		return err

	}

}

func (rlr *RedisLeaderboardRepository) GetTopK(k int64) ([]UserLeaderboardData, error) {
	result, err := rlr.Client.ZRevRangeWithScores(rlr.Ctx, rlr.Key, 0, k).Result()
	response := make([]UserLeaderboardData, 0)

	if err == nil {
		for i := 0; i < len(result); i++ {
			id := result[i]
			response = append(response, UserLeaderboardData{Score: id.Score, Id: id.Member.(string)})
		}
	}
	
	return response, err
}

func (prl *PostgresLeaderboardRepository) GetTopK(k int64) ([]UserLeaderboardData, error) {

	rows, err := prl.Client.Query(fmt.Sprintf("SELECT ID, SCORE FROM %s ORDER BY SCORE DESC LIMIT %d", prl.Table, k))

	if err != nil {
		return nil, err
	}
	response := make([]UserLeaderboardData, 0)

	for rows.Next() {
		var data UserLeaderboardData

		if err := rows.Scan(&data.Id, &data.Score); err != nil {
			return response, err
		}

		response = append(response, data)

	}
	rows.Close()
	return response, nil
}
