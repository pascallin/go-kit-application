package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/go-kit/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestUserServiceLogin(t *testing.T) {
	logger := log.NewLogfmtLogger(os.Stderr)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("login succeed", func(mt *mtest.T) {
		db := mt.DB
		svc := NewUserService(db, logger)

		docs := bson.D{
			{Key: "_id", Value: "123"},
			{Key: "username", Value: "pascal"},
			{Key: "password", Value: "3858f62230ac3c915f300c664312c63f"},
			{Key: "Nickname", Value: "lin"},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.users", mt.DB.Name()), mtest.FirstBatch, docs))

		token, err := svc.Login(context.Background(), "pascal", "foobar")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(token)
	})

	mt.Run("login with wrong password", func(mt *mtest.T) {
		db := mt.DB
		svc := NewUserService(db, logger)

		docs := bson.D{
			{Key: "_id", Value: "123"},
			{Key: "username", Value: "pascal"},
			{Key: "password", Value: "fake"},
			{Key: "Nickname", Value: "lin"},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.users", mt.DB.Name()), mtest.FirstBatch, docs))

		_, err := svc.Login(context.Background(), "pascal", "foobar")
		if err != nil && !errors.Is(err, ErrWrongPassword) {
			t.Fatalf("expected ErrWrongPassword")
		}
	})

	mt.Run("register succeed", func(mt *mtest.T) {
		db := mt.DB
		svc := NewUserService(db, logger)

		find := mtest.CreateCursorResponse(1, fmt.Sprintf("%s.users", mt.DB.Name()), mtest.FirstBatch)
		killCursors := mtest.CreateCursorResponse(
			0,
			fmt.Sprintf("%s.users", mt.DB.Name()),
			mtest.NextBatch)
		mt.AddMockResponses(
			find,
			killCursors,
		)
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		id, err := svc.Register(context.Background(), "pascal", "foobar", "lin")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(id.String())
	})

	mt.Run("register error with existed user", func(mt *mtest.T) {
		db := mt.DB
		svc := NewUserService(db, logger)

		docs := bson.D{
			{Key: "_id", Value: "123"},
			{Key: "username", Value: "pascal"},
			{Key: "password", Value: "fake"},
			{Key: "Nickname", Value: "lin"},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, fmt.Sprintf("%s.users", mt.DB.Name()), mtest.FirstBatch, docs))
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		_, err := svc.Register(context.Background(), "pascal", "foobar", "lin")
		if err != nil && !errors.Is(err, ErrExistedUsername) {
			t.Fatalf("expected ErrExistedUsername")
		}
	})
}
