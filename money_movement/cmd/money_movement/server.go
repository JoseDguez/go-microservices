package main

import (
	"database/sql"
	"fmt"
	"github.com/JoseDguez/go-microservices/internal/implementation/money_movement"
	pb "github.com/JoseDguez/go-microservices/proto"
	"google.golang.org/grpc"
	"log"
	"net"
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
	moneyMovementServerImplementation := money_movement.NewMoneyMovementImplementation(db)
	pb.RegisterMoneyMovementServiceServer(grpcServer, moneyMovementServerImplementation)

	// listen & serve
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen on port 50052: %v", err)
	}

	log.Printf("Server listening at :%v", listener.Addr())
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 50052: %v", err)
	}
}
