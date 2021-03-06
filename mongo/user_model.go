package mongo

import (
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/Maekes/planer/mongo/role"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type userModel struct {
	Id         bson.ObjectId `bson:"_id,omitempty"`
	UUID       uuid.UUID
	Username   string
	Password   []byte
	Mail       string
	Role       role.Role
	Created    time.Time
	Active     bool
	PublicID   uuid.UUID
	Planfooter []string
}

type userSettings struct {
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
