package mongo

import (
	uuid "github.com/satori/go.uuid"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MiniModel struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	UUID     uuid.UUID
	Vorname  string
	Nachname string
	Gruppe   string
	UserUUID uuid.UUID
}

func miniModelIndex() mgo.Index {
	return mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
}
