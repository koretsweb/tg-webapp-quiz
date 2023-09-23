package player

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage interface {
	Insert(ctx context.Context, p Player) (Player, error)
	Replace(ctx context.Context, oldP, newP Player) (Player, error)
	GetByID(ctx context.Context, id xid.ID) (Player, error)
	Delete(ctx context.Context, id xid.ID) error
	All(ctx context.Context) ([]Player, error)
	Filter(ctx context.Context, req FilterRequest, offset, limit uint) (total uint, pp []Player, err error)
}

type StorageMongo struct {
	collection *mongo.Collection
}

func NewStorageMongo() *StorageMongo {
	return &StorageMongo{}
}

func (s *StorageMongo) Insert(ctx context.Context, p Player) (Player, error) {
	_, err := s.collection.InsertOne(ctx, p)
	if err != nil {
		return Player{}, s.convertErr(err)
	}

	return p, nil
}

func (s *StorageMongo) Replace(ctx context.Context, oldP, newP Player) (Player, error) {
	if oldP.ID != newP.ID {
		return Player{}, ErrIDMismatch
	}

	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":     oldP.ID,
			"version": oldP.Version,
		},
		bson.M{"$set": newP},
	)
	if err != nil {
		return Player{}, s.convertErr(err)
	}

	if res.ModifiedCount == 0 {
		return Player{}, ErrVersionMismatch
	}

	return newP, nil
}

func (s *StorageMongo) GetByID(ctx context.Context, id xid.ID) (Player, error) {
	var p Player

	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		return Player{}, s.convertErr(err)
	}

	return p, err
}

func (s *StorageMongo) Delete(ctx context.Context, id xid.ID) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return s.convertErr(err)
	}

	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *StorageMongo) All(ctx context.Context) ([]Player, error) {
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, s.convertErr(err)
	}

	defer cursor.Close(ctx) // nolint

	return s.cursorToPlayers(ctx, cursor)
}

func (s *StorageMongo) cursorToPlayers(ctx context.Context, cursor *mongo.Cursor) ([]Player, error) {
	players := make([]Player, 0)

	if err := cursor.All(ctx, &players); err != nil {
		return nil, fmt.Errorf("cursor convert all: %w", err)
	}

	return players, nil
}

func (s *StorageMongo) Filter(
	ctx context.Context,
	req FilterRequest,
	offset,
	limit uint,
) (total uint, pp []Player, err error) {
	if req.Name == "" && req.Email == "" {
		return 0, nil, ErrEmptyRequest
	}

	filter := bson.M{}

	if req.Name != "" {
		filter["name"] = req.Name
	}

	if req.Email != "" {
		filter["email"] = req.Email
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, nil, fmt.Errorf("count players: %w", err)
	}

	if count == 0 {
		return 0, []Player{}, nil
	}

	cursor, err := s.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSkip(int64(offset)).
			SetLimit(int64(limit)),
	)
	if err != nil {
		return 0, nil, fmt.Errorf("find players: %w", err)
	}

	defer cursor.Close(ctx) // nolint

	pp, err = s.cursorToPlayers(ctx, cursor)
	if err != nil {
		return 0, nil, err
	}

	return uint(count), pp, nil
}

func (s *StorageMongo) Setup(ctx context.Context) error {
	// todo
	//_, err := s.collection.Indexes().CreateOne(
	//	ctx,
	//	mongo.IndexModel{
	//		Keys:    bson.M{"email": 1},
	//		Options: options.Index().SetUnique(true).SetName("email_idx"),
	//	},
	//)

	return errors.New("")
}

func (s *StorageMongo) convertErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}

	if mongo.IsDuplicateKeyError(err) {
		return ErrConflict
	}

	return err
}
