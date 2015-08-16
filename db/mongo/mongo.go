package mongo

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// Session default db connection
	Session *mgo.Session
)

type (
	// ID type abstraction of mgo's bson.ObjectId
	ID bson.ObjectId

	// Model defines the interface for mongo.Models
	Model interface {
		ID() bson.ObjectId
		DBName() string
		CName() string
	}

	// SessionError for mongo session failures
	SessionError struct{}

	// ErrorList collection of multiple errors
	ErrorList []error
)

// Error implements error interface
func (e SessionError) Error() string {
	return "Mongo session is not valid or nil"
}

// Error implements error interface
func (e ErrorList) Error() string {
	err := ""
	for _, one := range e {
		err += one.Error() + "\n"
	}
	return err
}

// IndexKey ensures an index on the given keys
func IndexKey(m Model, key ...string) error {
	s := session()
	defer s.Close()

	return c(s, m).EnsureIndexKey(key...)
}

// Load a db model matching the given conditions
func Load(m Model, query bson.M) error {
	s := session()
	defer s.Close()

	q := c(s, m).Find(query)
	n, err := q.Count()
	if err != nil {
		return err
	}
	if n > 0 {
		q.One(m)
	}
	return nil
}

// LoadAll db models matching the given conditions
func LoadAll(m Model, query bson.M, into []Model) error {
	s := session()
	defer s.Close()

	return c(s, m).Find(query).All(into)
}

// Save the given db models
func Save(m ...Model) error {
	var errl ErrorList

	s := session()
	defer s.Close()

	for _, model := range m {
		id := bson.ObjectId(model.ID())
		if !id.Valid() {
			id = bson.NewObjectId()
		}
		_, err := c(s, model).Upsert(bson.M{"_id": id}, model)
		if err != nil {
			errl = append(errl, err)
		}
	}
	if len(errl) > 0 {
		return errl
	}
	return nil
}

// session returns a *mgo.Session
func session() *mgo.Session {
	return Session.Copy()
}

// c returns a *mgo.Collection
func c(s *mgo.Session, m Model) *mgo.Collection {
	return s.DB(m.DBName()).C(m.CName())
}
