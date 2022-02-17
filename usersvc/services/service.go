package services

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/pascallin/go-kit-application/conn"
)

// Service describes a service that adds things together.
type Service interface {
	Register(ctx context.Context, username, password, nickname string) (error, primitive.ObjectID)
}

// New returns a basic Service with all of the expected middlewares wired in.
func NewService(logger log.Logger) Service {
	var svc Service
	{
		svc = NewUserService()
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

type userService struct{}

func NewUserService() Service {
	return userService{}
}

type User struct {
	Username string `bson:"username" json:"username"`
	Nickname string `bson:"nickname" json:"nickname"`
	Password string `bson:"password" json:"password"`
}

func (s userService) findUserByUserName(username string) (error, *User) {
	user := &User{}
	err := db.MongoDB.DB.Collection("users").
		FindOne(context.Background(), bson.M{"username": username}).Decode(user)
	if err != nil {
		return err, nil
	}
	return nil, user
}

func (s userService) login(username string, password string) (err error, token string) {
	err, user := s.findUserByUserName(username)
	if err != nil {
		return err, ""
	}
	p := md5.Sum([]byte(password))
	if user.Password != fmt.Sprintf("%x", p) {
		return errors.New("wrong password"), ""
	}
	gentoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		//"nbf": time.Date(2020, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})
	tokenString, err := gentoken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return errors.New("generate token error: " + err.Error()), ""
	}
	return nil, tokenString
}

func (s userService) Register(ctx context.Context, username, password, nickname string) (error, primitive.ObjectID) {
	_, existUser := s.findUserByUserName(username)
	fmt.Println(existUser)
	if existUser != nil {
		return errors.New("username existed"), primitive.NilObjectID
	}
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	p := md5.Sum([]byte(password))
	insertResult, err := db.MongoDB.DB.Collection("users").InsertOne(dbCtx, User{
		username,
		nickname,
		fmt.Sprintf("%x", p),
	})
	if err != nil {
		return err, primitive.NilObjectID
	}
	return nil, insertResult.InsertedID.(primitive.ObjectID)
}

func (s userService) updatePassword(username, password, newPassword string) error {
	var user User
	p := md5.Sum([]byte(password))
	matchUser := db.MongoDB.DB.Collection("users").
		FindOne(context.Background(), bson.M{
			"username": username,
			"password": fmt.Sprintf("%x", p),
		})
	if matchUser == nil {
		return errors.New("username and old password not match")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	np := md5.Sum([]byte(newPassword))
	after := options.After
	err := db.MongoDB.DB.Collection("users").
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
