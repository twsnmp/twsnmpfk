// Package datastore : データ保存
package datastore

import (
	"log"
	"os"

	"go.etcd.io/bbolt"
)

var dstTx *bbolt.Tx

func BackupDB(file string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	dstDB, err := bbolt.Open(file, 0600, nil)
	if err != nil {
		return err
	}
	defer func() {
		dstDB.Close()
	}()
	dstTx, err = dstDB.Begin(true)
	if err != nil {
		return err
	}
	err = db.View(func(srcTx *bbolt.Tx) error {
		return srcTx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			return walkBucket(b, nil, name, nil, b.Sequence())
		})
	})
	if err != nil {
		_ = dstTx.Rollback()
		return err
	}
	return dstTx.Commit()
}

var configBuckets = map[string]bool{
	"config":   true,
	"nodes":    true,
	"lines":    true,
	"items":    true,
	"pollings": true,
	"networks": true,
	"grok":     true,
	"images":   true,
	"certs":    true,
	"memo":     true,
}

func walkBucket(b *bbolt.Bucket, keypath [][]byte, k, v []byte, seq uint64) error {
	if v == nil {
		if _, ok := configBuckets[string(k)]; !ok {
			log.Printf("skip backup %s", string(k))
			return nil
		}
		log.Printf("do backup %s", string(k))
	}
	// Execute callback.
	if err := walkFunc(keypath, k, v, seq); err != nil {
		return err
	}

	// If this is not a bucket then stop.
	if v != nil {
		return nil
	}
	// Iterate over each child key/value.
	keypath = append(keypath, k)
	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			bkt := b.Bucket(k)
			return walkBucket(bkt, keypath, k, nil, bkt.Sequence())
		}
		return walkBucket(b, keypath, k, v, b.Sequence())
	})
}

func walkFunc(keys [][]byte, k, v []byte, seq uint64) error {
	// Create bucket on the root transaction if this is the first level.
	nk := len(keys)
	if nk == 0 {
		bkt, err := dstTx.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := bkt.SetSequence(seq); err != nil {
			return err
		}
		return nil
	}
	// Create buckets on subsequent levels, if necessary.
	b := dstTx.Bucket(keys[0])
	if nk > 1 {
		for _, k := range keys[1:] {
			b = b.Bucket(k)
		}
	}
	// Fill the entire page for best compaction.
	b.FillPercent = 1.0
	// If there is no value then this is a bucket call.
	if v == nil {
		bkt, err := b.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := bkt.SetSequence(seq); err != nil {
			return err
		}
		return nil
	}
	// Otherwise treat it as a key/value pair.
	return b.Put(k, v)
}

func CompactDB(srcPath, dstPath string) error {
	fi, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	src, err := bbolt.Open(srcPath, 0444, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := bbolt.Open(dstPath, fi.Mode(), nil)
	if err != nil {
		return err
	}
	defer dst.Close()
	return bbolt.Compact(dst, src, 1024*1024*64)
}
