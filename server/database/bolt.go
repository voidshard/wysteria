/*
The BoltDB implementation is suitable for testing or single server small scale use.

It's wicked fast and highly efficient but offers nothing in the way of replication or tolerance
against large scale node failure(s).
*/

package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	wyc "github.com/voidshard/wysteria/common"
)

var (
	// The boltdb collections we're going to be writing to
	bucketCollection = []byte(tableCollection)
	bucketItem       = []byte(tableItem)
	bucketVersion    = []byte(tableVersion)
	bucketResource   = []byte(tableFileresource)
	bucketLink       = []byte(tableLink)

	// additional bucket that we use to keep track of unique constraints we need to keep in mind
	bucketCollisions = []byte("collisions")
)

type boltDb struct {
	db *bolt.DB
}

// Called on start up, this opens / creates all of our boltdb buckets defined above for writing
func (b *boltDb) createBuckets() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{bucketCollection, bucketItem, bucketVersion, bucketResource, bucketLink, bucketCollisions} {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return nil
			}
		}
		return nil
	})
}

// Standard function to kick off a new boltdb 'connection'
// It's a bit of misnomer since we don't really talk to another host, but still.
func boltConnect(settings *Settings) (Database, error) {
	db, err := bolt.Open(settings.Database, 0600, nil)
	if err != nil {
		return nil, err
	}

	bolt_endpoint := &boltDb{
		db: db,
	}
	return bolt_endpoint, bolt_endpoint.createBuckets()
}

// Set the version with the given Id as the published version
// ToDo: We assume here the ParentId is valid.
func (b *boltDb) SetPublished(in string) error {
	version_id := []byte(in)
	return b.db.Update(func(tx *bolt.Tx) error {
		version_data := tx.Bucket(bucketVersion).Get(version_id)
		if version_data == nil {
			return errors.New(fmt.Sprintf("Version with id %s doesn't exist", in))
		}
		version := &wyc.Version{}
		err := version.UnmarshalJSON(version_data)
		if err != nil {
			return err
		}

		collision_key := []byte(fmt.Sprintf("published:%s", version.Parent))
		return tx.Bucket(bucketCollisions).Put(collision_key, version_id)
	})
}

// Given the Id of some item, return the version we've got marked as published (if any)
func (b *boltDb) Published(in string) (*wyc.Version, error) {
	collision_key := []byte(fmt.Sprintf("published:%s", in))

	val := &wyc.Version{}
	err := b.db.Update(func(tx *bolt.Tx) error {
		version_id := tx.Bucket(bucketCollisions).Get(collision_key)
		if version_id == nil {
			return errors.New(fmt.Sprintf("No published version for item with id %s", in))
		}

		version_data := tx.Bucket(bucketVersion).Get(version_id)
		if version_data == nil {
			return errors.New(fmt.Sprintf("Version with id %s doesn't exist", in))
		}

		return val.UnmarshalJSON(version_data)
	})
	return val, err
}

// Generic Util func for inserting something into the db with the given Id.
// Note that we're using the fact that wysteria base objects all fit the Marshalable interface
// defined in the common/ package.
func (b *boltDb) genericInsert(id string, in wyc.Marshalable, bucket []byte) error {
	data, err := in.MarshalJSON()
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put([]byte(id), data)
	})
}

// Insert a collection into the database given the Id and the Collection itself
func (b *boltDb) InsertCollection(id string, in *wyc.Collection) error {
	// From collision key for a collection name
	collision_key := collision_key_collection(in)

	// Marshal our given collection to []byte
	// This is done here so we don't do anything to unnecessary inside our transaction
	data, err := in.MarshalJSON()
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		// See if our collision key exists already (if so, we've made a collection of this name before)
		val := tx.Bucket(bucketCollisions).Get(collision_key)
		if val != nil {
			return errors.New("Unable to create: Would cause duplicate Collection")
		}

		// If not, we should add the collision key so we can't make it
		// At first glance there appears to be a race here, but recall that "Update" on a bucket
		// locks it for a 'read write' transaction.
		// Other routines that reach this Update func will also need to lock the db for the whole
		// transaction. So we should be good ..
		err := tx.Bucket(bucketCollisions).Put(collision_key, []byte("x")) // Nb the value here is not used
		if err != nil {
			return err
		}

		// Finally, save our collection to the db
		return tx.Bucket(bucketCollection).Put([]byte(id), data)
	})
}

// Get a unique collision key for the given item.
//  - We require it to match on any item from the same parent, with the same type and variant
func collision_key_item(in *wyc.Item) []byte {
	return []byte(fmt.Sprintf("item:%s:%s:%s", in.Parent, in.ItemType, in.Variant))
}

func (b *boltDb) InsertItem(id string, in *wyc.Item) error {
	// Form unique collision for items with the same parent, type & variant
	collision_key := collision_key_item(in)

	// Marshal to []byte before transaction
	data, err := in.MarshalJSON()
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		// Check if we've made this before in the collision bucket
		val := tx.Bucket(bucketCollisions).Get(collision_key)
		if val != nil {
			return errors.New(fmt.Sprintf("Unable to insert Item %s %s it exists in collection already", in.ItemType, in.Variant))
		}

		// if not, set it in the collision bucket to say we've made it
		err := tx.Bucket(bucketCollisions).Put(collision_key, []byte("x")) // Nb the value here is not used
		if err != nil {
			return err
		}

		// finally, save the actual item data
		return tx.Bucket(bucketItem).Put([]byte(id), data)
	})
}

// Struct to handle parsing Int32 to and from []byte
// There must be a better way, but at the time of googling not much of note seemed apparent.
// Need to research more.
// This struct isn't exposed, it's simply used in the functions
// int32ToByte (convert int32 -> []byte)
// byteToInt32 (convert []byte -> int32)
// ToDo: Use encoding/binary for this?
type bint32 struct {
	I int32 `json:"I"`
}

// Change int32 -> []byte
func int32ToByte(in int32) ([]byte, error) {
	tmp := &bint32{in}
	return json.Marshal(tmp)
}

// Change []byte -> int32
func byteToInt32(in []byte) (int32, error) {
	tmp := &bint32{}
	err := json.Unmarshal(in, tmp)
	if err != nil {
		return 0, err
	}
	return tmp.I, nil
}

// Get a unique collision key for the given version.
//  - We require it to match on any child version of the parent item
func collision_key_version(parent_id string) []byte {
	return []byte(fmt.Sprintf("version:%s", parent_id))
}

// Insert version into the db & set version number
func (b *boltDb) InsertNextVersion(id string, in *wyc.Version) (int32, error) {
	// Create unique key to track the greatest version number yet created for a given Item
	collision_key := collision_key_version(in.Parent)

	// Initialise version number to 0
	new_version := int32(0)

	return new_version, b.db.Update(func(tx *bolt.Tx) error {
		// Get the saved value for our unique key
		raw_val := tx.Bucket(bucketCollisions).Get(collision_key)
		if raw_val == nil {
			// If it doesn't exist, we're the first created version .. lucky us!
			new_version = 1
		} else {
			// If it does exist, we need to add 1 to it
			val, err := byteToInt32(raw_val)
			if err != nil {
				return err
			}
			new_version = int32(val) + 1
		}

		// Now we need the version number as a []byte again to store it
		raw_result, err := int32ToByte(new_version)
		if err != nil {
			return err
		}

		// Save our new version number with our unique key
		err = tx.Bucket(bucketCollisions).Put(collision_key, raw_result)
		if err != nil {
			return err
		}

		// Set the version number of our version
		in.Number = new_version

		// marshal our version to []byte
		data, err := in.MarshalJSON()
		if err != nil {
			return err
		}

		// Save the data to our version bucket
		return tx.Bucket(bucketVersion).Put([]byte(id), data)
	})
}

// Save the given resource with given Id to the database
func (b *boltDb) InsertResource(id string, in *wyc.Resource) error {
	return b.genericInsert(id, in, bucketResource)
}

// Save the given link with given Id to the database
func (b *boltDb) InsertLink(id string, in *wyc.Link) error {
	return b.genericInsert(id, in, bucketLink)
}

// Retrieve all collections that match the given Id(s)
func (b *boltDb) RetrieveCollection(ids ...string) (result []*wyc.Collection, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketCollection)
		for _, id := range ids {
			value := bucket.Get([]byte(id))
			if value == nil {
				continue
			}

			obj := &wyc.Collection{}
			err := obj.UnmarshalJSON(value)
			if err != nil {
				return err
			}
			result = append(result, obj)
		}
		return nil
	})
	return
}

// Retrieve all items that match the given Id(s)
func (b *boltDb) RetrieveItem(ids ...string) (result []*wyc.Item, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketItem)
		for _, id := range ids {
			value := bucket.Get([]byte(id))
			if value == nil {
				continue
			}

			obj := &wyc.Item{}
			err := obj.UnmarshalJSON(value)
			if err != nil {
				return err
			}
			result = append(result, obj)
		}
		return nil
	})
	return
}

// Retrieve all versions that match the given Id(s)
func (b *boltDb) RetrieveVersion(ids ...string) (result []*wyc.Version, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketVersion)
		for _, id := range ids {
			value := bucket.Get([]byte(id))
			if value == nil {
				continue
			}

			obj := &wyc.Version{}
			err := obj.UnmarshalJSON(value)
			if err != nil {
				return err
			}
			result = append(result, obj)
		}
		return nil
	})
	return
}

// Retrieve all resources that match the given Id(s)
func (b *boltDb) RetrieveResource(ids ...string) (result []*wyc.Resource, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketResource)
		for _, id := range ids {
			value := bucket.Get([]byte(id))
			if value == nil {
				continue
			}

			obj := &wyc.Resource{}
			err := obj.UnmarshalJSON(value)
			if err != nil {
				return err
			}
			result = append(result, obj)
		}
		return nil
	})
	return
}

// Retrieve all links that match the given Id(s)
func (b *boltDb) RetrieveLink(ids ...string) (result []*wyc.Link, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketLink)
		for _, id := range ids {
			value := bucket.Get([]byte(id))
			if value == nil {
				continue
			}

			obj := &wyc.Link{}
			err := obj.UnmarshalJSON(value)
			if err != nil {
				return err
			}
			result = append(result, obj)
		}
		return nil
	})
	return
}

// Update all facets of the version matching the given Id
func (b *boltDb) UpdateItem(id string, in *wyc.Item) error {
	return b.genericInsert(id, in, bucketItem)
}

// Update all facets of the item matching the given Id
func (b *boltDb) UpdateVersion(id string, in *wyc.Version) error {
	return b.genericInsert(id, in, bucketVersion)
}

// Generic util func to call delete by Id(s) on some bucket
func (b *boltDb) genericDelete(bucket []byte, ids ...string) error {
	for _, id := range ids {
		err := b.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(bucket).Delete([]byte(id))
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Get a unique collision key for the given collection.
//  - We require it to match on any collection of the same name
func collision_key_collection(collection *wyc.Collection) []byte {
	return []byte(fmt.Sprintf("collection:%s", collection.Name))
}

// Delete collection(s) by Id(s).
// Deletes relevant collection collision data too.
func (b *boltDb) DeleteCollection(ids ...string) error {
	// Retrieve the collections (we need the name fields to remove collisions)
	collections, err := b.RetrieveCollection(ids...)
	if err != nil {
		return err
	}

	// Remove any recorded collisions for these collections
	for _, collection := range collections {
		collision_key := collision_key_collection(collection)
		err = b.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(bucketCollisions).Delete(collision_key)
		})
		if err != nil {
			return err
		}
	}

	// Remove collections themselves
	return b.genericDelete(bucketCollection, ids...)
}

// Delete item(s) by Id(s)
// Deletes relevant item collision data too.
func (b *boltDb) DeleteItem(ids ...string) error {
	// Retrieve the items (we need the parent, type & variant fields to remove collision data)
	items, err := b.RetrieveItem(ids...)
	if err != nil {
		return err
	}

	// Remove all matching collision data for our items
	for _, item := range items {
		published_version_collision_key := collision_key_version(item.Id)
		item_collision_key := collision_key_item(item)

		// remove unique item type / variant information (item collision key)
		err = b.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(bucketCollisions).Delete(item_collision_key)
		})
		if err != nil {
			return err
		}

		// remove publish information (version collision key)
		err = b.db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(bucketCollisions).Delete(published_version_collision_key)
		})
		if err != nil {
			return err
		}
	}

	return b.genericDelete(bucketItem, ids...)
}

// Delete version(s) by Id(s)
func (b *boltDb) DeleteVersion(ids ...string) error {
	return b.genericDelete(bucketVersion, ids...)
}

// Delete resource(s) by Id(s)
func (b *boltDb) DeleteResource(ids ...string) error {
	return b.genericDelete(bucketResource, ids...)
}

// Delete link(s) by Id(s)
func (b *boltDb) DeleteLink(ids ...string) error {
	return b.genericDelete(bucketLink, ids...)
}

// Close connection to db
func (b *boltDb) Close() error {
	return b.db.Close()
}
