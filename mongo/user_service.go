package mongo

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Maekes/planer/mongo/role"
	uuid "github.com/satori/go.uuid"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserService struct {
	collection *mgo.Collection
	aktUser    uuid.UUID
}

const (
	mongoUrl = "localhost:27017"
	dbName   = "test_db"
)

func (p *UserService) ForUser(u uuid.UUID) {
	p.aktUser = u
}

func NewUserService(session *Session, dbName string, collectionName string) *UserService {
	collection := session.GetCollection(dbName, collectionName)
	collection.EnsureIndex(userModelIndex())
	return &UserService{collection, uuid.Nil}
}

func (p *UserService) ExistsAdmin() bool {
	i, err := p.collection.Find(bson.M{"role": role.Admin}).Count()
	if err != nil {
		//TODO
	}
	if i == 0 {
		return false
	} else {
		return true
	}

}

func (p *UserService) GetUsernameByID(u uuid.UUID) string {
	model := userModel{}
	err := p.collection.Find(bson.M{"uuid": u}).One(&model)
	if err != nil {
		//TODO
	}
	return model.Username
}

func (p *UserService) GetAktUser() *userModel {
	model := userModel{}
	err := p.collection.Find(bson.M{"uuid": p.aktUser}).One(&model)
	if err != nil {
		//TODO
	}
	return &model
}

func (p *UserService) GetRoleByID(u uuid.UUID) role.Role {
	model := userModel{}
	err := p.collection.Find(bson.M{"uuid": u}).One(&model)
	if err != nil {
		//TODO
	}
	return model.Role

}

func (p *UserService) CreateNewUser(name, mail, password string, role role.Role) error {
	i, _ := p.collection.Find(bson.M{"username": name}).Count()
	if i > 0 {
		return errors.New("Der Nutzername ist bereits vergeben")
	}

	i, _ = p.collection.Find(bson.M{"mail": mail}).Count()
	if i > 0 {
		return errors.New("Diese E-Mail Adresse wird bereits verwendet")
	}

	uid := uuid.NewV4()
	hp, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return errors.New("Es ist ein Fehler aufgetreten. Bitte versuchen Sie es erneut.")
	}

	u := userModel{
		UUID:     uid,
		Username: name,
		Mail:     mail,
		Password: hp,
		Role:     role,
		Created:  time.Now(),
		Active:   true,
	}

	return p.collection.Insert(&u)
}

func (p *UserService) GetByUsername(username string) (*userModel, error) {
	model := userModel{}
	err := p.collection.Find(bson.M{"username": username}).One(&model)
	return &model, err
}

func (p *UserService) GetAllUser() (*[]userModel, error) {
	var results []userModel
	err := p.collection.Find(nil).All(&results)
	return &results, err
}
func (p *UserService) DeleteUserById(Id string) error {
	uid, err := uuid.FromString(Id)
	err = p.collection.Remove(bson.M{"uuid": uid})
	return err
}

func (p *UserService) AdminDeleteUserById(Id string) error {
	uid, err := uuid.FromString(Id)
	err = p.collection.Remove(bson.M{"uuid": uid})
	return err
}

func (p *UserService) AdminChangeUserPasswordById(Id, password string) error {
	uid, err := uuid.FromString(Id)
	hp, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return errors.New("Es ist ein Fehler aufgetreten. Bitte versuchen Sie es erneut.")
	}
	err = p.collection.Update(bson.M{"uuid": uid}, bson.M{"$set": bson.M{"password": hp}})
	return err
}

func (p *UserService) ChangeUserPasswordById(Id, password string) error {
	uid, err := uuid.FromString(Id)
	err = p.collection.Update(bson.M{"uuid": uid}, bson.M{"$set": bson.M{"password": password}})
	return err
}

func (p *UserService) GetNewestUser() (*userModel, error) {
	model := userModel{}
	err := p.collection.Find(nil).Sort("-_id").One(&model)
	return &model, err
}

func (p *UserService) ValidateUser(username, password string) error {
	model := userModel{}

	err := p.collection.Find(bson.M{"username": username}).One(&model)
	if err != nil {
		return errors.New("Der Benutzername konnte nicht gefunden werden oder das Passwort ist falsch.")
	}

	err = bcrypt.CompareHashAndPassword(model.Password, []byte(password))

	if err != nil {
		return errors.New("Der Benutzername konnte nicht gefunden werden oder das Passwort ist falsch.")
	} else {
		return nil
	}

}
