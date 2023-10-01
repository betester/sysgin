package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
	"github.com/sysygn/cache/factory"
	"github.com/sysygn/cache/leaderboard"
)

const (
	TABLE_NAME        = "LEADERBOARD"
	READ_BUFFER_SIZE  = 1024
	WRITE_BUFFER_SIZE = 1024
)

func main() {

	dbArg := flag.String("db", "postgres", "database that server will use either redis or postges")
	portArg := flag.Int("port", 8080, "port the server will host")
	userArg := flag.Int("user", 10, "concurrent number of user that tries to modify the database")
	kArg := flag.Int("k", 5, "top K user from the database that will fetched")

	flag.Parse()

	var database leaderboard.LeaderboardRepository
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  READ_BUFFER_SIZE,
		WriteBufferSize: WRITE_BUFFER_SIZE,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	leaderboardFactory := factory.LeaderboardDatabaseFactory{Ctx: context.Background(), Table: TABLE_NAME}
	database = leaderboardFactory.CreateDb(*dbArg)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

		if err != nil {
			return
		}

		defer conn.Close()

		for {
			startTime := time.Now() // Record the start time before making the database call

			response, err := database.GetTopK(int64(*kArg))

			if err != nil {
				conn.WriteJSON(err)
			} else {
				// Calculate the elapsed time in milliseconds
				elapsedTimeMs := time.Since(startTime).Milliseconds()

				// Include the elapsed time in the response
				responseWithTime := map[string]interface{}{
					"concurrent_user": *userArg,
					"used_database":   *dbArg,
					"response":        response,
					"elapsed_time_ms": elapsedTimeMs,
				}

				conn.WriteJSON(responseWithTime)
			}

			time.Sleep(time.Second)
		}
	})

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Hello world",
		})
	})
	done := make(chan interface{})
	defer close(done)
	mockUserRequest(database, done, *userArg)
	r.Run(fmt.Sprintf(":%d", *portArg))
}

func mockUserRequest(database leaderboard.LeaderboardRepository, done <-chan interface{}, user int) {

	mockInsert := func(id string) {
		for {
			select {
			case <-done:
				return
			default:
				randomScore := rand.Intn(1000000)
				err := database.Insert(leaderboard.UserLeaderboardData{Score: float64(randomScore), Id: id})
				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Printf("SUCCESS: Score %d input from user %s\n", randomScore, id)
				}
				time.Sleep(6 * time.Second)
			}
		}
	}

	for i := 0; i < user; i++ {
		newUuid := uuid.New().String()
		go mockInsert(newUuid)
	}

}
