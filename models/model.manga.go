package models

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MangaMagnetModel struct {
	DB *mongo.Database
}

type MangaListFields struct {
	ID          primitive.ObjectID `bson:"_id, omitempty"`
	Name        string
	UriName     string
	Link        string
	CreatedDate time.Time
	UpdatedDate time.Time
}

type Chapter struct {
	MangaID     primitive.ObjectID `bson:"mangaId"`
	ChapterName string             `bson:"chapterName"`
	Link        string             `bson:"originalUrl"`
	ImagePath   []string           `bson:"imagePath"`
	Date        time.Time          `bson:"createdDate"`
}

func (m MangaMagnetModel) GetMangaList() []MangaListFields {
	cur, err := m.DB.Collection("mangaList").Find(context.TODO(), bson.M{})
	if err != nil {
		log.Panic(err)
	}

	var fields []MangaListFields
	if err = cur.All(context.TODO(), &fields); err != nil {
		log.Fatal(err)
	}

	return fields
}

func (m MangaMagnetModel) GetLatestChapter(mangaId primitive.ObjectID, chapterName string) (Chapter, error) {
	var result Chapter
	err := m.DB.Collection("mangaChapter").FindOne(context.TODO(), bson.M{"mangaId": mangaId, "chapterName": chapterName}).Decode(&result)
	return result, err

}

func (m MangaMagnetModel) SetMangaChapter(doc interface{}) error {
	_, err := m.DB.Collection("mangaChapter").InsertOne(context.TODO(), doc)
	if err != nil {
		return err
	}

	// fmt.Printf("inserted document with ID %v\n", res.InsertedID)
	return nil
}
