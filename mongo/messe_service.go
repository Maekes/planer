package mongo

import (
	"errors"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MesseService struct {
	collection *mgo.Collection
	aktUser    uuid.UUID
}

func NewMesseService(session *Session, dbName string, collectionName string) *MesseService {
	collection := session.GetCollection(dbName, collectionName)
	collection.EnsureIndex(userModelIndex())
	return &MesseService{collection, uuid.Nil}
}

func (p *MesseService) ForUser(u uuid.UUID) {
	p.aktUser = u
}

func (p *MesseService) Create(m *MesseModel) error {
	m.UserUUID = p.aktUser
	return p.collection.Insert(&m)
}

func (p *MesseService) GetAllMessen() (*[]MesseModel, error) {
	var results []MesseModel
	err := p.collection.Find(bson.M{"useruuid": p.aktUser}).Sort("datum").All(&results)

	return &results, err
}

func (p *MesseService) GetMesseByUUID(uid uuid.UUID) (*MesseModel, error) {
	var m MesseModel
	err := p.collection.Find(bson.M{"uuid": uid, "useruuid": p.aktUser}).One(&m)

	return &m, err
}

func (p *MesseService) GetAllMessenWithUUIDsPublic(uids []string) (*[]MesseModel, error) {
	var m []MesseModel
	var uidss []uuid.UUID
	for _, u := range uids {
		uid, _ := uuid.FromString(u)
		uidss = append(uidss, uid)
	}

	err := p.collection.Find(bson.M{"uuid": bson.M{"$in": uidss}}).All(&m)

	return &m, err
}

func (p *MesseService) UpdateMesse(m *MesseModel) error {
	err := p.collection.Update(bson.M{"uuid": m.UUID}, &m)
	return err
}

func (p *MesseService) AddNameToMessePublic(n string, uid uuid.UUID) error {
	var m MesseModel
	err := p.collection.Find(bson.M{"uuid": uid}).One(&m)
	m.Rueckmeldungen = append(m.Rueckmeldungen, n)
	err = p.collection.Update(bson.M{"uuid": m.UUID}, &m)
	return err
}

func (p *MesseService) GetAllMessenFromDate(fromDate time.Time) (*[]MesseModel, error) {
	var results []MesseModel
	err := p.collection.Find(
		bson.M{
			"datum": bson.M{
				"$gt": fromDate,
			},
			"useruuid": p.aktUser,
		}).Sort("datum").All(&results)
	return &results, err

}

func (p *MesseService) GetAllMessenFromToDate(fromDate, toDate time.Time) (*[]MesseModel, error) {
	var results []MesseModel
	err := p.collection.Find(
		bson.M{
			"datum": bson.M{
				"$gt": fromDate,
				"$lt": toDate,
			},
			"useruuid": p.aktUser,
		}).Sort("datum").All(&results)
	return &results, err

}

func (p *MesseService) DeleteAllMessenToDate(toDate time.Time) error {
	err := p.collection.Remove(
		bson.M{
			"datum": bson.M{
				"$lt": toDate,
			},
			"useruuid": p.aktUser,
		})
	return err

}

func (p *MesseService) GetAllMessenThatAreRelevant() (*[]MesseModel, error) {
	var results []MesseModel
	err := p.collection.Find(bson.M{"isrelevant": true, "useruuid": p.aktUser}).Sort("datum").All(&results)

	return &results, err
}

func (p *MesseService) GetAllMessenThatAreRelevantFromToDate(fromDate, toDate time.Time) (*[]MesseModel, error) {
	var results []MesseModel
	err := p.collection.Find(
		bson.M{
			"datum": bson.M{
				"$gt": fromDate,
				"$lt": toDate,
			},
			"isrelevant": true,
			"useruuid":   p.aktUser,
		}).Sort("datum").All(&results)
	return &results, err

}

func (p *MesseService) GetAllMessenThatAreRelevantFromToDatePublic(fromDate, toDate time.Time, user uuid.UUID) (*[]MesseModel, error) {
	var results []MesseModel
	err := p.collection.Find(
		bson.M{
			"datum": bson.M{
				"$gt": fromDate,
				"$lt": toDate,
			},
			"useruuid":   user,
			"isrelevant": true,
		}).Sort("datum").All(&results)
	return &results, err

}

func (p *MesseService) GetMaxDate() (time.Time, error) {
	var results MesseModel
	err := p.collection.Find(
		bson.M{
			"useruuid":   p.aktUser,
			"isrelevant": true,
		}).Sort("-datum").One(&results)
	if err != nil {
		return time.Now(), err
	}
	return results.Datum, nil
}

func (p *MesseService) AddMiniToMesse(UId uuid.UUID, m MiniModel) error {

	// Update the necessary 'Company' document
	i, err := p.collection.Find(bson.M{"minisforplan": bson.M{"$in": []uuid.UUID{m.UUID}}, "uuid": UId, "useruuid": p.aktUser}).Count()
	if i == 0 {
		err = nil
		err = p.collection.Update(bson.M{"uuid": UId, "useruuid": p.aktUser}, bson.M{"$addToSet": bson.M{"minisforplan": m.UUID}})
	} else {
		err = errors.New("Mini already in field")
	}
	return err
}

func (p *MesseService) DeleteMiniFromMesse(miniUId uuid.UUID, messeUId uuid.UUID) error {

	err := p.collection.Update(bson.M{"uuid": messeUId, "useruuid": p.aktUser}, bson.M{"$pull": bson.M{"minisforplan": bson.M{"$in": []uuid.UUID{miniUId}}}})

	return err
}

func (p *MesseService) DeleteMesseByUId(UId uuid.UUID) error {
	err := p.collection.Remove(bson.M{"uuid": UId, "useruuid": p.aktUser})
	return err
}

func (p *MesseService) DeleteAllMessen() error {
	_, err := p.collection.RemoveAll(bson.M{"useruuid": p.aktUser})
	return err
}

func (p *MesseService) UpdateRelevantMesseByUId(UId uuid.UUID, state bool) error {
	err := p.collection.Update(bson.M{"uuid": UId, "useruuid": p.aktUser}, bson.M{"$set": bson.M{"isrelevant": state}})
	return err
}

func (p *MesseService) ChangeInfoForPlan(UId uuid.UUID, value string) error {
	err := p.collection.Update(bson.M{"uuid": UId, "useruuid": p.aktUser}, bson.M{"$set": bson.M{"infoforplan": value}})
	return err
}

func (p *MesseService) CountMiniInMessen(from time.Time, to time.Time, mini uuid.UUID) int {

	c, err := p.collection.Find(bson.M{
		"$and": []bson.M{
			bson.M{"datum": bson.M{
				"$gte": from,
				"$lte": to,
			}},
			bson.M{"minisforplan": mini},
		},
		"useruuid": p.aktUser,
	},
	).Count()

	if err != nil {
		fmt.Println(err)
		//TODO
	}

	return c
}
