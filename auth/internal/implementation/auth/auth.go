package auth

import (
	"context"
	"database/sql"
	"errors"
	pb "github.com/JoseDguez/go-microservices/proto"
	jwt "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"time"
)

type Implementation struct {
	db *sql.DB
	pb.UnimplementedAuthServiceServer
}

func NewAuthImplementation(db *sql.DB) *Implementation {
	return &Implementation{db: db}
}

func (this *Implementation) GetToken(ctx context.Context, credentials *pb.Credentials) (*pb.Token, error) {
	type user struct {
		userID   string
		password string
	}

	var u user

	stmt, err := this.db.Prepare("SELECT user_id, password FROM users WHERE user_id = ? AND password = ?")
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = stmt.QueryRow(credentials.GetUsername(), credentials.GetPassword()).Scan(&u.userID, &u.password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	token, err := createJWT(u.userID)
	if err != nil {
		return nil, err
	}

	return &pb.Token{Jwt: token}, nil
}

func (this *Implementation) ValidateToken(ctx context.Context, token *pb.Token) (*pb.User, error) {
	return &pb.User{}, nil
}

func createJWT(userID string) (string, error) {
	key := []byte(os.Getenv("SIGNING_KEY"))
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "auth-service",
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
	})

	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	return signedToken, nil
}
