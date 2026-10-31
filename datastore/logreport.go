package datastore

import (
	"go.etcd.io/bbolt"
)

var (
	bucketLogReport = []byte("logReport")
)

// GetLogReportData returns the raw JSON of one report entity, or (nil, nil) if absent.
func GetLogReportData(kind, id string) ([]byte, error) {
	if kind == "" || id == "" {
		return nil, ErrInvalidParams
	}
	if db == nil {
		return nil, ErrDBNotOpen
	}
	var ret []byte
	err := db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket(bucketLogReport)
		if root == nil {
			return nil
		}
		b := root.Bucket([]byte(kind))
		if b == nil {
			return nil
		}
		if v := b.Get([]byte(id)); v != nil {
			ret = append([]byte(nil), v...)
		}
		return nil
	})
	return ret, err
}

// ListLogReportData returns every entity of a report kind keyed by ID.
func ListLogReportData(kind string) (map[string][]byte, error) {
	if kind == "" {
		return nil, ErrInvalidParams
	}
	if db == nil {
		return nil, ErrDBNotOpen
	}
	ret := make(map[string][]byte)
	err := db.View(func(tx *bbolt.Tx) error {
		root := tx.Bucket(bucketLogReport)
		if root == nil {
			return nil
		}
		b := root.Bucket([]byte(kind))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			ret[string(k)] = append([]byte(nil), v...)
			return nil
		})
	})
	return ret, err
}

// SaveLogReportData upserts entities of a report kind in a single transaction.
func SaveLogReportData(kind string, items map[string][]byte) error {
	if kind == "" {
		return ErrInvalidParams
	}
	if len(items) == 0 {
		return nil
	}
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists(bucketLogReport)
		if err != nil {
			return err
		}
		b, err := root.CreateBucketIfNotExists([]byte(kind))
		if err != nil {
			return err
		}
		for id, v := range items {
			if id == "" {
				continue
			}
			if err := b.Put([]byte(id), v); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteLogReportData removes the given entity IDs of a report kind.
func DeleteLogReportData(kind string, ids []string) error {
	if kind == "" {
		return ErrInvalidParams
	}
	if len(ids) == 0 {
		return nil
	}
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		root := tx.Bucket(bucketLogReport)
		if root == nil {
			return nil
		}
		b := root.Bucket([]byte(kind))
		if b == nil {
			return nil
		}
		for _, id := range ids {
			if err := b.Delete([]byte(id)); err != nil {
				return err
			}
		}
		return nil
	})
}

// ResetLogReportData removes all entities of a kind (or all kinds when kind is empty).
func ResetLogReportData(kind string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		if kind == "" {
			if err := tx.DeleteBucket(bucketLogReport); err != nil && err != bbolt.ErrBucketNotFound {
				return err
			}
			return nil
		}
		root := tx.Bucket(bucketLogReport)
		if root == nil {
			return nil
		}
		if err := root.DeleteBucket([]byte(kind)); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}
		return nil
	})
}
