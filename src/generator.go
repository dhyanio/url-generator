package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"

	"net/http"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// urls -> URL database structure
type urls struct {
	gorm.Model
	Tinyurl string `gorm:"unique;not null"`
	Longurl string
}

// PostgresClient -> Provides a connection to the postgres database server
func PostgresClient() *gorm.DB {
	dbClient, err := gorm.Open("postgres", "host=127.0.0.1 port=5432 user=postgres dbname=tiny_scale_go password=<db password> sslmode=disable")
	if err != nil {
		panic(err)
	}
	return dbClient
}

// RedisClient -> Provides a connection to the Redis server
func RedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return client
}

// StoreTinyURL -> puts the urls into cache and DB
func StoreTinyURL(dbURLData urls, longURL string, tinyURL string, dbClient *gorm.DB, redisClient *redis.Client) {
	dbClient.Create(&dbURLData)
	redisClient.HSet("urls", tinyURL, longURL)
}

// GenerateHashAndInsert -> Genarates a unique tiny URL and inserts it to DB
func GenerateHashAndInsert(longURL string, startIndex int, dbClient *gorm.DB, redisClient *redis.Client) string {
	byteURLData := []byte(longURL)
	hashedURLData := fmt.Sprintf("%x", md5.Sum(byteURLData))
	tinyURLRegex, err := regexp.Compile("[/+]")
	if err != nil {
		return "Unable to generate tiny URL"
	}
	tinyURLData := tinyURLRegex.ReplaceAllString(base64.URLEncoding.EncodeToString([]byte(hashedURLData)), "_")
	if len(tinyURLData) < (startIndex + 6) {
		return "Unable to generate tiny URL"
	}
	tinyURL := tinyURLData[startIndex : startIndex+6]
	var dbURLData urls
	dbClient.Where("tinyurl = ?", tinyURL).Find(&dbURLData)
	if dbURLData.Tinyurl == "" {
		fmt.Println(dbURLData, "in not found")
		go StoreTinyURL(urls{Tinyurl: tinyURL, Longurl: longURL}, longURL, tinyURL, dbClient, redisClient)
		return tinyURL
	} else if (dbURLData.Tinyurl == tinyURL) && (dbURLData.Longurl == longURL) {
		fmt.Println(dbURLData, "in found and equal")
		return tinyURL
	} else {
		return GenerateHashAndInsert(longURL, startIndex+1, dbClient, redisClient)
	}
}

// IndexHandler -> Handles requests coming to / route
func IndexHandler(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "Welcome!\n")
}

// GetTinyHandler -> Generates tiny URL and returns it
func GetTinyHandler(res http.ResponseWriter, req *http.Request, dbClient *gorm.DB, redisClient *redis.Client) {
	requestParams, err := req.URL.Query()["longUrl"]
	if !err || len(requestParams[0]) < 1 {
		io.WriteString(res, "URL parameter longUrl is missing")
	} else {
		longURL := requestParams[0]
		tinyURL := GenerateHashAndInsert(longURL, 0, dbClient, redisClient)
		io.WriteString(res, tinyURL)
	}
}

// GetLongHandler -> Fetches long URL and returns it
func GetLongHandler(res http.ResponseWriter, req *http.Request, dbClient *gorm.DB, redisClient *redis.Client) {
	requestParams, err := req.URL.Query()["tinyUrl"]
	if !err || len(requestParams[0]) < 1 {
		io.WriteString(res, "URL parameter tinyUrl is missing")
	}
	tinyURL := requestParams[0]
	redisSearchResult := redisClient.HGet("urls", tinyURL)
	if redisSearchResult.Val() != "" {
		io.WriteString(res, redisSearchResult.Val())
	} else {
		var url urls
		dbClient.Where("tinyurl = ?", tinyURL).Select("longurl").Find(&url)
		if url.Longurl != "" {
			redisClient.HSet("urls", tinyURL, url.Longurl)
			io.WriteString(res, url.Longurl)
		} else {
			io.WriteString(res, "Unable to find long URL")
		}
	}
}

// StopHandler -> Stops the server on request to /stop route
func StopHandler(res http.ResponseWriter, req *http.Request, dbClient *gorm.DB, redisClient *redis.Client, serverInstance *http.Server) {
	fmt.Println("Stopping server...\n")
	dbClient.Close()
	redisClient.Close()
	serverInstance.Shutdown(context.TODO())
}

func main() {
	redisClient := RedisClient()

	pong, err := redisClient.Ping().Result()
	fmt.Println("Redis ping", pong, err)

	dbClient := PostgresClient()
	defer dbClient.Close()

	dbClient.AutoMigrate(&urls{})

	serverInstance := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/long/", func(w http.ResponseWriter, r *http.Request) {
		GetLongHandler(w, r, dbClient, redisClient)
	})

	http.HandleFunc("/tiny/", func(w http.ResponseWriter, r *http.Request) {
		GetTinyHandler(w, r, dbClient, redisClient)
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		StopHandler(w, r, dbClient, redisClient, serverInstance)
	})

	http.HandleFunc("/", IndexHandler)

	serverInstance.ListenAndServe()
}
