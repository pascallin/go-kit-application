package services

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrWrongPassword   = errors.New("wrong password")
	ErrExistedUsername = errors.New("username existed")
)

// Service describes a service that adds things together.
type Service interface {
	Register(ctx context.Context, username, password, nickname string) (primitive.ObjectID, error)
	Login(ctx context.Context, username string, password string) (token string, err error)
	UpdatePassword(ctx context.Context, username, password, newPassword string) error
}

// New returns a basic Service with all of the expected middlewares wired in.
func NewService(db *mongo.Database, logger log.Logger) Service {
	var svc Service
	{
		svc = NewUserService(db, logger)
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

type userService struct {
	db     *mongo.Database
	logger log.Logger
}

func NewUserService(db *mongo.Database, logger log.Logger) Service {
	return &userService{
		db:     db,
		logger: logger,
	}
}

type User struct {
	Username string `bson:"username" json:"username"`
	Nickname string `bson:"nickname" json:"nickname"`
	Password string `bson:"password" json:"password"`
}

func (s *userService) findUserByUserName(ctx context.Context, username string) (user *User, err error) {
	user = &User{}
	err = s.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(user)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection, skip this error in this method
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) Login(ctx context.Context, username string, password string) (token string, err error) {
	user, err := s.findUserByUserName(ctx, username)
	if err != nil {
		return "", err
	}

	p := md5.Sum([]byte(password))

	if user.Password != fmt.Sprintf("%x", p) {
		return "", ErrWrongPassword
	}
	gentoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		//"nbf": time.Date(2020, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})
	tokenString, err := gentoken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", errors.New("generate token error: " + err.Error())
	}
	return tokenString, nil
}

func (s *userService) Register(ctx context.Context, username, password, nickname string) (id primitive.ObjectID, err error) {
	existUser, err := s.findUserByUserName(ctx, username)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if existUser != nil {
		return primitive.NilObjectID, ErrExistedUsername
	}
	p := md5.Sum([]byte(password))
	insertResult, err := s.db.Collection("users").InsertOne(ctx, User{
		username,
		nickname,
		fmt.Sprintf("%x", p),
	})
	if err != nil {
		return primitive.NilObjectID, err
	}

	id = insertResult.InsertedID.(primitive.ObjectID)
	return id, nil
}

func (s *userService) UpdatePassword(ctx context.Context, username, password, newPassword string) (err error) {
	var user User
	p := md5.Sum([]byte(password))
	matchUser := s.db.Collection("users").
		FindOne(ctx, bson.M{
			"username": username,
			"password": fmt.Sprintf("%x", p),
		})
	if matchUser == nil {
		return errors.New("username and old password not match")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	np := md5.Sum([]byte(newPassword))
	after := options.After
	err = s.db.Collection("users").
		FindOneAndUpdate(ctx,
			bson.M{"username": username},
			bson.M{"$set": bson.M{"password": fmt.Sprintf("%x", np)}},
			&options.FindOneAndUpdateOptions{
				ReturnDocument: &after,
			},
		).Decode(&user)
	if err != nil {
		return errors.New("update password fail: " + err.Error())
	}
	return nil
}
