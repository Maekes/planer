package mongo

import (
	uuid "github.com/satori/go.uuid"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MiniService struct {
	collection *mgo.Collection
	AktUser    uuid.UUID
}

func NewMiniService(session *Session, dbName string, collectionName string) *MiniService {
	collection := session.GetCollection(dbName, collectionName)
	collection.EnsureIndex(userModelIndex())
	return &MiniService{collection, uuid.Nil}
}

func (p *MiniService) ForUser(u uuid.UUID) {
	p.AktUser = u
}

func (p *MiniService) Create(m *MiniModel) error {
	m.UserUUID = p.AktUser
	return p.collection.Insert(&m)
}

func (p *MiniService) GetAllMinis() (*[]MiniModel, error) {
	var results []MiniModel
	err := p.collection.Find(bson.M{"useruuid": p.AktUser}).Sort("nachname", "vorname").All(&results)

	return &results, err
}

func (p *MiniService) GetAllMinisFromGroup(g string) (*[]MiniModel, error) {
	var results []MiniModel
	err := p.collection.Find(bson.M{"useruuid": p.AktUser, "gruppe": g}).Sort("nachname", "vorname").All(&results)

	return &results, err
}

func (p *MiniService) GetMiniByUUID(UId uuid.UUID) (*MiniModel, error) {
	var results MiniModel
	err := p.collection.Find(bson.M{"uuid": UId, "useruuid": p.AktUser}).One(&results)

	return &results, err
}

func (p *MiniService) UpdateMini(m *MiniModel) error {

	err := p.collection.Update(bson.M{"uuid": m.UUID}, &m)
	return err
}

func (p *MiniService) DeleteMiniById(Id string) error {
	uid, err := uuid.FromString(Id)
	err = p.collection.Remove(bson.M{"uuid": uid, "useruuid": p.AktUser})
	return err
}
