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
	urlPrefix = "mongodb"

	// Keeps track of highest version
	countersCollection = "counters"

	// Keeps track of ItemId -> PublishedVersionId
	publishedCollection = "publish"
)

type mongoEndpoint struct {
	session  *mgo.Session
	db       *mgo.Database
	settings *Settings
}

// Connect to mongo db via an ssl connection
func mongoSslConnect(settings *Settings) (*mgo.Session, error) {
	url := formMongoUrl(settings)

	roots := x509.NewCertPool()
	ca, err := ioutil.ReadFile(settings.PemFile)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	roots.AppendCertsFromPEM(ca)
	tlsConfig := &tls.Config{
		// ToDo: Fix .. once I get a verifiable cert
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

// Create and return db wrapper & call connect
func mongoConnect(settings *Settings) (Database, error) {
	sess := &mgo.Session{}
	var err error

	if settings.PemFile != "" {
		sess, err = mongoSslConnect(settings)
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

// Set the version with the given Id as published
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
	// version marked as PublishedVersion (where the Version is a child of the given Item)
	version_obj := vers[0]
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"Item": version_obj.Parent, "PublishedVersion": version_id}},
		ReturnNew: true,
		Upsert:    true,
	}

	var doc bson.M
	col := e.getCollection(publishedCollection)

	// Atomic findAndModify call
	_, err = col.Find(bson.M{"Item": version_obj.Parent}).Apply(change, &doc)
	return err
}

// Get the published versino for the given item Id
func (e *mongoEndpoint) Published(item_id string) (*wyc.Version, error) {
	// Look up the published version id for the given item
	var doc bson.M
	col := e.getCollection(publishedCollection)
	err := col.Find(bson.M{"Item": item_id}).One(&doc)
	if err != nil {
		return nil, err
	}

	// Assuming we got one, pull out the PublishedVersion version id
	version_id := doc["PublishedVersion"].(string)

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

// Insert a collection into the db with the given Id
func (e *mongoEndpoint) InsertCollection(id string, d *wyc.Collection) error {
	collection := e.getCollection(tableCollection)

	var res []interface{}
	err := collection.Find(bson.M{"name": d.Name}).All(&res)
	if err != nil {
		return err
	}
	if len(res) > 0 {
		return errors.New("Unable to create: Would cause duplicate Collection")
	}

	return e.insert(tableCollection, id, d)
}

// Insert a item into the db with the given Id
func (e *mongoEndpoint) InsertItem(id string, d *wyc.Item) error {
	key := fmt.Sprintf("%s:%s:%s:%s", tableItem, d.Parent, d.ItemType, d.Variant)

	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"CounterFor": key}},
		ReturnNew: true,
		Upsert:    true,
	}

	var doc counter
	col := e.getCollection(countersCollection)
	details, err := col.Find(bson.M{"CounterFor": key}).Apply(change, &doc)
	if err != nil {
		return err
	}

	if details.Updated > 0 {
		return errors.New(fmt.Sprintf("Unable to insert Item %s %s it exists in collection already", d.ItemType, d.Variant))
	}

	return e.insert(tableItem, id, d)
}

// Insert a version into the db with the given Id & set version number
func (e *mongoEndpoint) InsertNextVersion(id string, d *wyc.Version) (int32, error) {
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"Count": 1}, "$set": bson.M{"CounterFor": d.Parent}},
		ReturnNew: true,
		Upsert:    true,
	}

	// bson.M works here, using 'counter' struct doesn't -> ??
	// http://www.mrwaggel.be/post/golang-mgo-findandmodify-auto-increment-id/
	var doc bson.M
	col := e.getCollection(countersCollection)

	// Atomic findAndModify call
	_, err := col.Find(bson.M{"CounterFor": d.Parent}).Apply(change, &doc)
	if err != nil {
		return -1, err
	}

	d.Number = int32(doc["Count"].(int))
	return d.Number, e.insert(tableVersion, id, d)
}

// Insert a resource into the db with the given Id
func (e *mongoEndpoint) InsertResource(id string, d *wyc.Resource) error {
	return e.insert(tableFileresource, id, d)
}

// Insert a link into the db with the given Id
func (e *mongoEndpoint) InsertLink(id string, d *wyc.Link) error {
	return e.insert(tableLink, id, d)
}


// Retrieve the collections indicated by the given Id(s)
func (e *mongoEndpoint) RetrieveCollection(ids ...string) (res []*wyc.Collection, err error) {
	err = e.retrieve(tableCollection, &res, ids...)
	return
}

// Retrieve the items indicated by the given Id(s)
func (e *mongoEndpoint) RetrieveItem(ids ...string) (res []*wyc.Item, err error) {
	err = e.retrieve(tableItem, &res, ids...)
	return
}

// Retrieve the versions indicated by the given Id(s)
func (e *mongoEndpoint) RetrieveVersion(ids ...string) (res []*wyc.Version, err error) {
	err = e.retrieve(tableVersion, &res, ids...)
	return
}

// Retrieve the resources indicated by the given Id(s)
func (e *mongoEndpoint) RetrieveResource(ids ...string) (res []*wyc.Resource, err error) {
	err = e.retrieve(tableFileresource, &res, ids...)
	return
}

// Retrieve the links indicated by the given Id(s)
func (e *mongoEndpoint) RetrieveLink(ids ...string) (res []*wyc.Link, err error) {
	err = e.retrieve(tableLink, &res, ids...)
	return
}


// Update the facets of the item with the given Id
func (e *mongoEndpoint) UpdateItem(id string, d *wyc.Item) error {
	return e.update(tableItem, id, d)
}

// Update the facets of the version with the given Id
func (e *mongoEndpoint) UpdateVersion(id string, d *wyc.Version) error {
	return e.update(tableVersion, id, d)
}


// Delete collections matching given Id(s)
func (e *mongoEndpoint) DeleteCollection(ids ...string) error {
	return e.deleteById(tableCollection, ids...)
}

// Delete items matching given Id(s)
//  Also delete unique constraint & publish data
func (e *mongoEndpoint) DeleteItem(ids ...string) error {
	items, err := e.RetrieveItem(ids...)
	if err != nil {
		return err
	}

	counter_col := e.getCollection(countersCollection)
	publish_col := e.getCollection(publishedCollection)
	for _, item := range items {
		key := fmt.Sprintf("%s:%s:%s:%s", tableItem, item.Parent, item.ItemType, item.Variant)

		counter_col.RemoveAll(bson.M{"CounterFor": key})
		counter_col.RemoveAll(bson.M{"CounterFor": item.Id})
		publish_col.RemoveAll(bson.M{"Item": item.Id})
	}

	return e.deleteById(tableItem, ids...)
}

// Delete versions matching given Id(s)
func (e *mongoEndpoint) DeleteVersion(ids ...string) error {
	return e.deleteById(tableVersion, ids...)
}

// Delete resources matching given Id(s)
func (e *mongoEndpoint) DeleteResource(ids ...string) error {
	return e.deleteById(tableFileresource, ids...)
}

// Delete links matching given Id(s)
func (e *mongoEndpoint) DeleteLink(ids ...string) error {
	return e.deleteById(tableLink, ids...)
}


// Generic insert of some document into a named column with the given id
func (e *mongoEndpoint) insert(col string, sid string, data interface{}) error {
	// First, we need to format the Id as a bson
	// ToDo: Look at generic-izing the setting / creation of Ids
	err := bsonIdCheck(sid)
	if err != nil {
		return err
	}

	// get the given collection
	collection := e.getCollection(col)

	// upsert our document into mongo, setting the id as desired
	_, err = collection.Upsert(bson.M{"_id": bson.ObjectIdHex(sid)}, data)
	return err
}

// Generic retrieve doc(s) by Id(s) from the named database collection
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

// Generic delete all from the named collection by Id(s)
func (e *mongoEndpoint) deleteById(col string, sids ...string) (err error) {
	ids, err := toBsonIds(sids...)
	if err != nil {
		return err
	}

	collection := e.getCollection(col)
	_, err = collection.RemoveAll(bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	})
	return err
}

// Generic update document from the named collection with the given Id
func (e *mongoEndpoint) update(col, sid string, data interface{}) (err error) {
	err = bsonIdCheck(sid)
	if err != nil {
		return
	}

	collection := e.getCollection(col)
	return collection.UpdateId(bson.ObjectIdHex(sid), data)
}

// Convert Ids of type string (default) into our mongo format (bson.ObjectId)
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

// Form the mongo connection url given the database settings
func formMongoUrl(settings *Settings) string {
	url := fmt.Sprintf("%s://", urlPrefix)
	if settings.User != "" {
		url += settings.User
		if settings.Pass != "" {
			url += ":" + settings.Pass
		}
		url += "@"
	}
	return url + settings.Host + ":" + strconv.Itoa(settings.Port) + "/" + settings.Database
}

// Check if a given Id in it's default string format is a valid mongo bson.ObjectId
func bsonIdCheck(sid string) error {
	if bson.IsObjectIdHex(sid) {
		return nil
	}

	return errors.New(fmt.Sprintf("Invalid ID: %s", sid))
}

// Get named collection from the mongo database
func (e *mongoEndpoint) getCollection(name string) *mgo.Collection {
	return e.db.C(name)
}
