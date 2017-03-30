package database

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net"
	"strconv"
)

const (
	url_prefix = "mongodb"

	// Keeps track of highest version
	counters_collection = "counters"

	// Keeps track of ItemId -> PublishedVersionId
	published_collection = "publish"
)

type mongoEndpoint struct {
	session  *mgo.Session
	db       *mgo.Database
	settings *DatabaseSettings
}

// Connection funcs
func mongo_ssl_connect(settings *DatabaseSettings) (*mgo.Session, error) {
	url := formMongoUrl(settings)

	roots := x509.NewCertPool()
	ca, err := ioutil.ReadFile(settings.PemFile)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	roots.AppendCertsFromPEM(ca)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tlsConfig.RootCAs = roots

	dialInfo, err := mgo.ParseURL(url)
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}

	return mgo.DialWithInfo(dialInfo)
}

// Connect function to be used in func Connect (database.go)
//
func mongo_connect(settings *DatabaseSettings) (Database, error) {
	sess := &mgo.Session{}
	var err error

	if settings.PemFile != "" {
		sess, err = mongo_ssl_connect(settings)
	} else {
		url := formMongoUrl(settings)
		sess, err = mgo.Dial(url)
	}
	if err != nil {
		return nil, err
	}
	sess.SetMode(mgo.Monotonic, true)

	ep := mongoEndpoint{
		settings: settings,
		session:  sess,
	}
	ep.db = ep.session.DB(settings.Database)

	return &ep, nil
}

// Kill connection to db
//
func (e *mongoEndpoint) Close() error {
	e.session.Close()
	return nil
}

// Internal counter
// Used to,
//   Ensure items are unique in a collection
//   Keep count of the highest version for an item
type counter struct {
	// Internal counter obj
	CounterFor string `json:"CounterFor"`
	Count      int32  `json:"Count"`
}

func (e *mongoEndpoint) SetPublished(version_id string) error {
	// First, look up the version obj, because we need the parent Item ID
	vers, err := e.RetrieveVersion(version_id)
	if err != nil {
		return err
	}

	// Check we got the expected number of versions
	if len(vers) != 1 {
		return errors.New(fmt.Sprintf("Expected one matching version, got %d", len(vers)))
	}

	// We use a document with two values, the Item Id we're talking about and the ID of the current
	// version marked as Published (where the Version is a child of the given Item)
	version_obj := vers[0]
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"Item": version_obj.Parent, "Published": version_id}},
		ReturnNew: true,
		Upsert:    true,
	}

	var doc bson.M
	col := e.getCollection(published_collection)

	// Atomic findAndModify call
	_, err = col.Find(bson.M{"Item": version_obj.Parent}).Apply(change, &doc)
	return err
}

func (e *mongoEndpoint) GetPublished(item_id string) (*wyc.Version, error) {
	// Look up the published version id for the given item
	var doc bson.M
	col := e.getCollection(published_collection)
	err := col.Find(bson.M{"Item": item_id}).One(&doc)
	if err != nil {
		return nil, err
	}

	// Assuming we got one, pull out the Published version id
	version_id := doc["Published"].(string)

	// Retrieve said version
	vers, err := e.RetrieveVersion(version_id)
	if err != nil {
		return nil, err
	}

	// Check we got the expected number of versions
	if len(vers) != 1 {
		return nil, errors.New(fmt.Sprintf("Expected one matching version, got %d", len(vers)))
	}

	return vers[0], nil
}

func (e *mongoEndpoint) InsertCollection(id string, d *wyc.Collection) error {
	collection := e.getCollection(table_collection)

	var res []interface{}
	err := collection.Find(bson.M{"name": d.Name}).All(&res)
	if err != nil {
		return err
	}
	if len(res) > 0 {
		return errors.New("Unable to create: Would cause duplicate Collection")
	}

	return e.insert(table_collection, id, d)
}

func (e *mongoEndpoint) InsertItem(id string, d *wyc.Item) error {
	key := fmt.Sprintf("%s:%s:%s:%s", table_item, d.Parent, d.ItemType, d.Variant)

	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"CounterFor": key}},
		ReturnNew: true,
		Upsert:    true,
	}

	var doc counter
	col := e.getCollection(counters_collection)
	details, err := col.Find(bson.M{"CounterFor": key}).Apply(change, &doc)
	if err != nil {
		return err
	}

	if details.Updated > 0 {
		return errors.New(fmt.Sprintf("Unable to insert Item %s %s it exists in collection already", d.ItemType, d.Variant))
	}

	return e.insert(table_item, id, d)
}

func (e *mongoEndpoint) InsertNextVersion(id string, d *wyc.Version) (int32, error) {
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"Count": 1}, "$set": bson.M{"CounterFor": d.Parent}},
		ReturnNew: true,
		Upsert:    true,
	}

	// bson.M works here, using 'counter' struct doesn't -> ??
	// http://www.mrwaggel.be/post/golang-mgo-findandmodify-auto-increment-id/
	var doc bson.M
	col := e.getCollection(counters_collection)

	// Atomic findAndModify call
	_, err := col.Find(bson.M{"CounterFor": d.Parent}).Apply(change, &doc)
	if err != nil {
		return -1, err
	}

	d.Number = int32(doc["Count"].(int))
	return d.Number, e.insert(table_version, id, d)
}
func (e *mongoEndpoint) InsertResource(id string, d *wyc.Resource) error {
	return e.insert(table_fileresource, id, d)
}

func (e *mongoEndpoint) InsertLink(id string, d *wyc.Link) error {
	return e.insert(table_link, id, d)
}

// Retrieve impl
func (e *mongoEndpoint) RetrieveCollection(ids ...string) (res []*wyc.Collection, err error) {
	err = e.retrieve(table_collection, &res, ids...)
	return
}

func (e *mongoEndpoint) RetrieveItem(ids ...string) (res []*wyc.Item, err error) {
	err = e.retrieve(table_item, &res, ids...)
	return
}
func (e *mongoEndpoint) RetrieveVersion(ids ...string) (res []*wyc.Version, err error) {
	err = e.retrieve(table_version, &res, ids...)
	return
}
func (e *mongoEndpoint) RetrieveResource(ids ...string) (res []*wyc.Resource, err error) {
	err = e.retrieve(table_fileresource, &res, ids...)
	return
}
func (e *mongoEndpoint) RetrieveLink(ids ...string) (res []*wyc.Link, err error) {
	err = e.retrieve(table_link, &res, ids...)
	return
}

// Update impl
func (e *mongoEndpoint) UpdateItem(id string, d *wyc.Item) error {
	return e.update(table_item, id, d)
}
func (e *mongoEndpoint) UpdateVersion(id string, d *wyc.Version) error {
	return e.update(table_version, id, d)
}

func (e *mongoEndpoint) DeleteCollection(ids ...string) error {
	return e.delete_by_id(table_collection, ids...)
}

func (e *mongoEndpoint) DeleteItem(ids ...string) error {
	return e.delete_by_id(table_item, ids...)
}

func (e *mongoEndpoint) DeleteVersion(ids ...string) error {
	return e.delete_by_id(table_version, ids...)
}

func (e *mongoEndpoint) DeleteResource(ids ...string) error {
	return e.delete_by_id(table_fileresource, ids...)
}

func (e *mongoEndpoint) DeleteLink(ids ...string) error {
	return e.delete_by_id(table_link, ids...)
}

// Util funcs
func (e *mongoEndpoint) insert(col string, sid string, data interface{}) error {
	err := bsonIdCheck(sid)
	if err != nil {
		return err
	}

	collection := e.getCollection(col)
	_, err = collection.Upsert(bson.M{"_id": bson.ObjectIdHex(sid)}, data)
	return err
}

func (e *mongoEndpoint) retrieve(col string, out interface{}, sids ...string) (err error) {
	ids, err := toBsonIds(sids...)
	if err != nil {
		return err
	}

	collection := e.getCollection(col)
	err = collection.Find(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}).All(out)
	return
}

func (e *mongoEndpoint) delete_by_id(col string, sids ...string) (err error) {
	ids, err := toBsonIds(sids...)
	if err != nil {
		return err
	}

	collection := e.getCollection(col)
	return collection.Remove(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	})
}

func (e *mongoEndpoint) update(col, sid string, data interface{}) (err error) {
	err = bsonIdCheck(sid)
	if err != nil {
		return
	}

	collection := e.getCollection(col)
	return collection.UpdateId(bson.ObjectIdHex(sid), data)
}

func toBsonIds(sids ...string) ([]bson.ObjectId, error) {
	ids := []bson.ObjectId{}
	for _, sid := range sids {
		err := bsonIdCheck(sid)
		if err != nil {
			return nil, err
		}
		ids = append(ids, bson.ObjectIdHex(sid))
	}
	return ids, nil
}

func formMongoUrl(settings *DatabaseSettings) string {
	url := fmt.Sprintf("%s://", url_prefix)
	if settings.User != "" {
		url += settings.User
		if settings.Pass != "" {
			url += ":" + settings.Pass
		}
		url += "@"
	}
	return url + settings.Host + ":" + strconv.Itoa(settings.Port) + "/" + settings.Database
}

func bsonIdCheck(sid string) error {
	if bson.IsObjectIdHex(sid) {
		return nil
	}

	return errors.New(fmt.Sprintf("Invalid ID: %s", sid))
}

func (e *mongoEndpoint) getCollection(name string) *mgo.Collection {
	return e.db.C(name)
}
