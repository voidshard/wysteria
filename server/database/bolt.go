package database

import (
	"github.com/boltdb/bolt"
	wyc "github.com/voidshard/wysteria/common"
	"fmt"
	"errors"
	"encoding/json"
)

var (
	bucket_collection = []byte(table_collection)
	bucket_item = []byte(table_item)
	bucket_version = []byte(table_version)
	bucket_resource = []byte(table_fileresource)
	bucket_link = []byte(table_link)

	// bucket that we use to keep track of unique constraints
	bucket_collisions = []byte("collisions")
)

type boltDb struct {
	db *bolt.DB
}

func (b *boltDb) create_buckets() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{bucket_collection, bucket_item, bucket_version, bucket_resource, bucket_link, bucket_collisions} {
			_, err := tx.CreateBucket(bucket)
			if err != nil {
				return nil
			}
		}
		return nil
	})
}

func bolt_connect(settings *DatabaseSettings) (Database, error) {
	db, err := bolt.Open(settings.Database, 0600, nil)
	if err != nil {
		return nil, err
	}

	bolt_endpoint := &boltDb{
		db: db,
	}
	return bolt_endpoint, bolt_endpoint.create_buckets()
}

func (b *boltDb) SetPublished(in string) error {
	version_id := []byte(in)
	return b.db.Update(func(tx *bolt.Tx) error {
		version_data := tx.Bucket(bucket_version).Get(version_id)
		if version_data == nil {
			return errors.New(fmt.Sprintf("Version with id %s doesn't exist", in))
		}
		version := &wyc.Version{}
		err := version.UnmarshalJSON(version_data)
		if err != nil {
			return err
		}

		collision_key := []byte(fmt.Sprintf("published:%s", version.Parent))
		return tx.Bucket(bucket_collisions).Put(collision_key, version_id)
	})
}

func (b *boltDb) GetPublished(in string) (*wyc.Version, error) {
	collision_key := []byte(fmt.Sprintf("published:%s", in))

	val := &wyc.Version{}
	err := b.db.Update(func(tx *bolt.Tx) error {
		version_id := tx.Bucket(bucket_collisions).Get(collision_key)
		if version_id == nil {
			return errors.New(fmt.Sprintf("No published version for item with id %s", in))
		}

		version_data := tx.Bucket(bucket_version).Get(version_id)
		if version_data == nil {
			return errors.New(fmt.Sprintf("Version with id %s doesn't exist", in))
		}

		return val.UnmarshalJSON(version_data)
	})
	return val, err
}

func (b *boltDb) generic_insert(id string, in wyc.Marshalable, bucket []byte) error {
	data, err := in.MarshalJSON()
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put([]byte(id), data)
	})
}

func (b *boltDb) InsertCollection(id string, in *wyc.Collection) error { // Ensure only one collection with given Name
	collision_key := []byte(fmt.Sprintf("collection:%s", in.Name))
	data, err := in.MarshalJSON()
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		val := tx.Bucket(bucket_collisions).Get(collision_key)
		if val != nil {
			return errors.New("Unable to create: Would cause duplicate Collection")
		}

		err := tx.Bucket(bucket_collisions).Put(collision_key, []byte("x"))
		if err != nil {
			return err
		}

		return tx.Bucket(bucket_collection).Put([]byte(id), data)
	})
}

func (b *boltDb) InsertItem(id string, in *wyc.Item) error { // Ensure only one item with the same Collection, Type and Variant
	collision_key := []byte(fmt.Sprintf("item:%s:%s:%s", in.Parent, in.ItemType, in.Variant))
	data, err := in.MarshalJSON()
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		val := tx.Bucket(bucket_collisions).Get(collision_key)
		if val != nil {
			return errors.New(fmt.Sprintf("Unable to insert Item %s %s it exists in collection already", in.ItemType, in.Variant))
		}

		err := tx.Bucket(bucket_collisions).Put(collision_key, []byte("x"))
		if err != nil {
			return err
		}

		return tx.Bucket(bucket_item).Put([]byte(id), data)
	})
}

// ToDo: Use encoding/binary for this
type bint32 struct {
	I int32 `json:"I"`
}

func int32_to_byte(in int32) ([]byte, error) {
	tmp := &bint32{in}
	return json.Marshal(tmp)
}

func byte_to_int32(in []byte) (int32, error) {
	tmp := &bint32{}
	err := json.Unmarshal(in, tmp)
	if err != nil {
		return 0, err
	}
	return tmp.I, nil
}

func (b *boltDb) InsertNextVersion(id string, in *wyc.Version) (int32, error) { // Ensure only one version of an Item with a given Number
	collision_key := []byte(fmt.Sprintf("version:%s", in.Parent))

	new_version := int32(0)

	return new_version, b.db.Update(func(tx *bolt.Tx) error {
		raw_val := tx.Bucket(bucket_collisions).Get(collision_key)
		if raw_val == nil {
			new_version = 1
		} else {
			val, err := byte_to_int32(raw_val)
			if err != nil {
				return err
			}
			new_version = int32(val) + 1
		}

		raw_result, err := int32_to_byte(new_version)
		if err != nil {
			return err
		}

		err = tx.Bucket(bucket_collisions).Put(collision_key, raw_result)
		if err != nil {
			return err
		}

		in.Number = new_version
		data, err := in.MarshalJSON()
		if err != nil {
			return err
		}

		return tx.Bucket(bucket_version).Put([]byte(id), data)
	})
}

func (b *boltDb) InsertResource(id string, in *wyc.Resource) error {
	return b.generic_insert(id, in, bucket_resource)
}

func (b *boltDb) InsertLink(id string, in *wyc.Link) error {
	return b.generic_insert(id, in, bucket_link)
}

func (b *boltDb) RetrieveCollection(ids ...string) (result []*wyc.Collection, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket_collection)
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

func (b *boltDb) RetrieveItem(ids ...string) (result []*wyc.Item, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket_item)
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

func (b *boltDb) RetrieveVersion(ids ...string) (result []*wyc.Version, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket_version)
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

func (b *boltDb) RetrieveResource(ids ...string) (result []*wyc.Resource, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket_resource)
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

func (b *boltDb) RetrieveLink(ids ...string) (result []*wyc.Link, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket_link)
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

func (b *boltDb) UpdateItem(id string, in *wyc.Item) error {
	return b.generic_insert(id, in, bucket_item)
}

func (b *boltDb) UpdateVersion(id string, in *wyc.Version) error {
	return b.generic_insert(id, in, bucket_version)
}

func (b *boltDb) generic_delete(bucket []byte, ids ...string) error {
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

func (b *boltDb) DeleteCollection(ids ...string) error {
	return b.generic_delete(bucket_collection, ids...)
}

func (b *boltDb) DeleteItem(ids ...string) error {
	return b.generic_delete(bucket_item, ids...)
}

func (b *boltDb) DeleteVersion(ids ...string) error {
	return b.generic_delete(bucket_version, ids...)
}

func (b *boltDb) DeleteResource(ids ...string) error {
	return b.generic_delete(bucket_resource, ids...)
}

func (b *boltDb) DeleteLink(ids ...string) error {
	return b.generic_delete(bucket_link, ids...)
}

func (b *boltDb) Close() error {
	return b.db.Close()
}
