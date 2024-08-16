package main

import (
	"database/sql"
	"fmt"
	"google.golang.org/grpc"
	"log"
)

const (
	dbDriver   = "mysql"
	dbUser     = "money_movement_user"
	dbPassword = "Admin123"
	dbName     = "money_movement"
)

var db *sql.DB

func main() {
	var err error

	// Connect to the database
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPassword, dbName)

	db, err = sql.Open(dbDriver, dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = db.Close(); err != nil {
			log.Printf("Failed to close database connection: %s", err)
		}
	}()

	// Ping the database to ensure the connection is active
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// grpc server setup
	grpcServer := grpc.NewServer()
}
