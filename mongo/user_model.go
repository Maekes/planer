package mongo

import (
	uuid "github.com/satori/go.uuid"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type userModel struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	UUID     uuid.UUID
	Username string
	Password []byte
	Mail     string
}

func userModelIndex() mgo.Index {
	return mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
}
