package mongo

import (
	uuid "github.com/satori/go.uuid"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PlanService struct {
	collection *mgo.Collection
	aktUser    uuid.UUID
}

func NewPlanService(session *Session, dbName string, collectionName string) *PlanService {
	collection := session.GetCollection(dbName, collectionName)
	collection.EnsureIndex(planModelIndex())
	return &PlanService{collection, uuid.Nil}
}

func (p *PlanService) ForUser(u uuid.UUID) {
	p.aktUser = u
}

func (p *PlanService) Create(m *PlanModel) error {
	m.UserUUID = p.aktUser
	return p.collection.Insert(&m)
}

func (p *PlanService) GetAllPlan() (*[]PlanModel, error) {
	var results []PlanModel
	err := p.collection.Find(bson.M{"useruuid": p.aktUser}).All(&results)

	return &results, err
}

func (p *PlanService) GetPlanByUUID(UId uuid.UUID) (*PlanModel, error) {
	var results PlanModel
	err := p.collection.Find(bson.M{"uuid": UId, "useruuid": p.aktUser}).One(&results)

	return &results, err
}

func (p *PlanService) DeletePlanById(Id string) error {
	uid, err := uuid.FromString(Id)
	err = p.collection.Remove(bson.M{"uuid": uid, "useruuid": p.aktUser})
	return err
}
