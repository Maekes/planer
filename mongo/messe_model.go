package mongo

import (
	"time"

	uuid "github.com/satori/go.uuid"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MesseModel struct {
	Id                 bson.ObjectId `bson:"_id,omitempty"`
	UUID               uuid.UUID
	Datum              time.Time
	KaplanID           string
	Ort                string
	Gottesdienst       string
	LiturgischerTag    string
	Bemerkung          string
	IsRelevant         bool
	ErforderlicheMinis int
	InfoForPlan        string
	MinisForPlan       []uuid.UUID
	Rueckmeldungen     []string
	UserUUID           uuid.UUID
}

func messeModelIndex() mgo.Index {
	return mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
}
