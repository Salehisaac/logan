package main

import (
	_ "encoding/json"
	"fmt"
	"log"
	"os"

	
	_ "github.com/go-sql-driver/mysql"
	
	"github.com/joho/godotenv"

)




func NewBot() (*Bot , error) {

    
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

	if err := ConnectDB(dsn); err != nil {
        log.Printf("Error connecting to database: %v", err)
    }

	db := GetDB()
    
    return &Bot{
		db: db,
        } , nil
}









