package config

import (
	"database/sql"

	"os"

	_ "github.com/go-sql-driver/mysql"

	"log"

	"github.com/joho/godotenv"
)

var db_name string = ""
var db_port string = ""
var db_user string = ""
var db_pass string = ""

func getEnvDB() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	db_name = os.Getenv("SECURE_ENV_DB_NAME")
	db_port = os.Getenv("SECURE_ENV_DB_PORT")
	db_user = os.Getenv("SECURE_ENV_DB_USER")
	db_pass = os.Getenv("SECURE_ENV_DB_PASSWORD")
}

func ConnectDB() (*sql.DB, error) {
	getEnvDB()
	db, err := sql.Open("mysql", db_user+":"+db_pass+"@tcp(localhost:"+db_port+")/"+db_name)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
