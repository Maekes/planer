package mongo

import (
	"time"

	uuid "github.com/satori/go.uuid"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PlanModel struct {
	Id                  bson.ObjectId `bson:"_id,omitempty"`
	UUID                uuid.UUID
	Erstellt            time.Time
	Von                 time.Time
	Bis                 time.Time
	Titel               string
	UserUUID            uuid.UUID
	Rueckmeldungen      []Rueckmeldung
	RueckmeldungHinweis string
}

type Rueckmeldung struct {
	Name    string
	UUID    uuid.UUID
	Hinweis string
	Zeit    time.Time
	Messen  []uuid.UUID
}

func planModelIndex() mgo.Index {
	return mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
}
