package internal

import (
	_ "encoding/json"
	"fmt"
	"log"
	"os"

	"log_reader/configs"
	"log_reader/internal/database"

	_ "github.com/go-sql-driver/mysql"

	"database/sql"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

type Bot struct {
	db *sql.DB
}

func NewBot() (*Bot, error) {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbName := os.Getenv("MYSQL_DB")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	if err := database.ConnectDB(dsn); err != nil {
		log.Printf("Error connecting to database: %v", err)
	}

	db := database.GetDB()

	return &Bot{
		db: db,
	}, nil
}

func (b Bot) InitData(phoneNumber string, cfg *configs.Config) {

	var authKey, client_system string

	userId, err := database.GetUserByPhoneNumber(phoneNumber, b.db)
	if err != nil {
		log.Fatalf("Failed to get user ID: %v", err)
	}

	authKeys, err := database.GetAuthkeysByUserId(userId, b.db)
	if err != nil {
		log.Fatalf("Failed to get user ID: %v", err)
	}

	sortedAuthKeys, err := database.SortAuthkeys(authKeys, b.db)

	if err != nil {
		log.Fatal("failed to sort the authkeys : ", err)
	}

	clients, err := database.GetDevicesByAuthKeyId(sortedAuthKeys, b.db)
	if err != nil {
		log.Fatal("error getting the client : ", err)
	}

	fmt.Println("choose a op system :")

	for index, client := range clients {
		fmt.Printf("%d)%s_%s_%s\n", index+1, client.DeviceModel, client.SystemVersion, client.ClientIp)
	}
	var opIndex int
	fmt.Scanln(&opIndex)

	for index, client := range clients {
		if index+1 == opIndex {
			authKeyint := client.AuthKeyId
			authKey = strconv.Itoa(authKeyint)
			client_system = strings.Replace(client.SystemVersion, " ", "_", -1)
		}
	}
	if authKey == "" {
		fmt.Println("invalid number")
		return
	}
	cfg.AuthKey = authKey
	cfg.ClientSystem = client_system
	cfg.PhoneNumber = phoneNumber
}
