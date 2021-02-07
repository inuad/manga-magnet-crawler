package models

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MangaListModel struct {
	DB *mongo.Database
}

type MangaRepository interface {
	GetMangaList() []Fields
}

type Fields struct {
	ID          primitive.ObjectID `bson:"_id, omitempty"`
	Name        string
	UriName     string
	Link        string
	CreatedDate time.Time
	UpdatedDate time.Time
}

func (m MangaListModel) GetMangaList() []Fields {
	cur, err := m.DB.Collection("mangaList").Find(context.TODO(), bson.M{})
	if err != nil {
		log.Panic(err)
	}

	var fields []Fields
	if err = cur.All(context.TODO(), &fields); err != nil {
		log.Fatal(err)
	}

	return fields
}
