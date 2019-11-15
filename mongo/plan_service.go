package mongo

import (
	"time"

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

func (p *PlanService) GetNewestPlanFromUser(UId uuid.UUID) (*PlanModel, error) {
	var result PlanModel
	err := p.collection.Find(bson.M{"useruuid": UId}).Sort("-erstellt").One(&result)
	return &result, err
}

func (p *PlanService) GetPlanByUUIDPublic(UId uuid.UUID) (*PlanModel, error) {
	var results PlanModel
	err := p.collection.Find(bson.M{"uuid": UId}).One(&results)

	return &results, err
}

func (p *PlanService) DeletePlanById(uid uuid.UUID) error {
	err := p.collection.Remove(bson.M{"uuid": uid, "useruuid": p.aktUser})
	return err
}

func (p *PlanService) NewRueckmeldungPublic(name string, messen []string, hinweis string, planid uuid.UUID) {
	var plan PlanModel
	var r Rueckmeldung

	r.Name = name
	r.UUID = uuid.NewV4()
	r.Hinweis = hinweis
	r.Zeit = time.Now()

	for _, m := range messen {
		uid, err := uuid.FromString(m)
		if err != nil {
			//TODO
		}
		r.Messen = append(r.Messen, uid)
	}

	err := p.collection.Find(bson.M{"uuid": planid}).One(&plan)
	plan.Rueckmeldungen = append(plan.Rueckmeldungen, r)
	err = p.collection.Update(bson.M{"uuid": planid}, &plan)
	if err != nil {
		//TODO
	}

}
