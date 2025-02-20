package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const (
	PORT = 5500
	ADDR = "127.0.0.1"

	DBPORT = 3306
	DBADDR = "192.168.1.19"
	DBUSER = "admin"
)

var db *sql.DB

type AliveRequest struct {
	JobId       string `json:"jobId"`
	PlayerCount int    `json:"playerCount"`
}

func checkUserWhitelist(UserId int) bool {
	var count int

	query := "SELECT COUNT(*) FROM Users WHERE UserId = ?"
	err := db.QueryRow(query, UserId).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	return count > 0
}

func pong(context *gin.Context) {
	context.String(200, "Pong")
}

func landing(context *gin.Context) {
	context.Redirect(302, "/ping")
}

func whitelistCheckHandler(context *gin.Context) {
	idStr := context.DefaultQuery("id", "i_will_error")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		context.String(400, "Invalid ID parameter")
		return
	}

	if checkUserWhitelist(id) {
		context.String(200, "True")
	} else {
		context.String(200, "False")
	}
}

func ServerPostHandler(context *gin.Context) {
	var request AliveRequest

	if err := context.ShouldBindJSON(&request); err != nil {
		context.String(400, "Bad request: "+err.Error())
		return
	}

	fmt.Println("Received JobId: " + request.JobId + " with " + strconv.Itoa(request.PlayerCount) + " players!")
	context.String(200, "Received!")
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env: ", err)
	}

	SQL_PASS := os.Getenv("MYSQL_PASSWORD")
	if SQL_PASS == "" {
		log.Fatal("Failed to obtain MYSQL_PASSWORD from environment variables")
	}

	dburl := fmt.Sprintf("%s:%s@tcp(%s:%d)/WhitelistDB", DBUSER, SQL_PASS, DBADDR, DBPORT)
	db, err = sql.Open("mysql", dburl)
	if err != nil {
		log.Fatal("Failed to open database: ", err)
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	router := gin.Default()
	api := router.Group("/api")
	router.SetTrustedProxies([]string{
		"127.0.0.1",
	})

	api.GET("/ping", pong)
	api.GET("/", landing)
	api.GET("/whitelistcheck", whitelistCheckHandler)
	api.POST("/roblox_server_response", ServerPostHandler)

	address := ADDR + ":" + strconv.Itoa(PORT)
	router.Run(address)
}
