package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"go.etcd.io/bbolt"
)

// BboltLogStats provides statistics about log records stored in bbolt.
type BboltLogStats struct {
	SyslogCount       int   `json:"syslogCount"`
	TrapCount         int   `json:"trapCount"`
	NetflowCount      int   `json:"netflowCount"`
	SFlowCount        int   `json:"sFlowCount"`
	SFlowCounterCount int   `json:"sFlowCounterCount"`
	ArpLogCount       int   `json:"arpLogCount"`
	PollingLogCount   int   `json:"pollingLogCount"`
	TotalCount        int   `json:"totalCount"`
	BboltSize         int64 `json:"bboltSize"`
	ParquetSize       int64 `json:"parquetSize"`
}

// GetBboltLogStats counts the log records stored in bbolt.
// For databases larger than 50MB, it avoids expensive recursive page traversals.
func GetBboltLogStats() BboltLogStats {
	stats := BboltLogStats{}
	if db == nil {
		return stats
	}

	_ = db.View(func(tx *bbolt.Tx) error {
		stats.BboltSize = tx.Size()

		// For large databases (>50MB), skip b.Stats() to prevent millions of page faults
		if stats.BboltSize > 50*1024*1024 {
			return nil
		}

		countBucket := func(name string) int {
			b := tx.Bucket([]byte(name))
			if b == nil {
				return 0
			}
			c := b.Cursor()
			if k, _ := c.First(); k == nil {
				return 0
			}
			return b.Stats().KeyN
		}

		stats.SyslogCount = countBucket("syslog")
		stats.TrapCount = countBucket("trap")
		stats.NetflowCount = countBucket("netflow")
		stats.SFlowCount = countBucket("sflow")
		stats.SFlowCounterCount = countBucket("sflowCounter")
		stats.ArpLogCount = countBucket("arplog")

		pb := tx.Bucket([]byte("pollingLogs"))
		if pb != nil {
			_ = pb.ForEachBucket(func(k []byte) error {
				sub := pb.Bucket(k)
				if sub != nil {
					sc := sub.Cursor()
					if sk, _ := sc.First(); sk != nil {
						stats.PollingLogCount += sub.Stats().KeyN
					}
				}
				return nil
			})
		}
		return nil
	})

	stats.TotalCount = stats.SyslogCount + stats.TrapCount + stats.NetflowCount +
		stats.SFlowCount + stats.SFlowCounterCount + stats.ArpLogCount + stats.PollingLogCount

	return stats
}

// MigrateBboltToParquet migrates all log data from bbolt to the Parquet log store using streaming batches.
func MigrateBboltToParquet(progressCb func(bucket string, current, total int)) error {
	if db == nil {
		return ErrDBNotOpen
	}
	if logStore == nil {
		if dspath != "" {
			logStore = NewParquetLogDataStore()
			if err := logStore.Open(filepath.Join(dspath, "logs")); err != nil {
				return fmt.Errorf("open parquet log store: %w", err)
			}
		} else {
			return fmt.Errorf("parquet log store is not initialized")
		}
	}

	total := 0
	processed := 0
	batchSize := 2000

	generalBuckets := []string{"syslog", "trap", "netflow", "sflow", "sflowCounter", "arplog"}

	for _, bucketName := range generalBuckets {
		if progressCb != nil {
			progressCb(bucketName, processed, total)
		}
		for {
			var batch []*LogEnt
			var keys [][]byte

			err := db.View(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte(bucketName))
				if b == nil {
					return nil
				}
				c := b.Cursor()
				for k, v := c.First(); k != nil && len(batch) < batchSize; k, v = c.Next() {
					v = deCompressLog(v)
					var l LogEnt
					if err := json.Unmarshal(v, &l); err != nil {
						continue
					}
					batch = append(batch, &l)
					keys = append(keys, append([]byte(nil), k...))
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("read bucket %s: %w", bucketName, err)
			}

			if len(batch) == 0 {
				break
			}

			if err := logStore.SaveLogs(bucketName, batch); err != nil {
				return fmt.Errorf("save parquet logs %s: %w", bucketName, err)
			}

			err = db.Update(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte(bucketName))
				if b == nil {
					return nil
				}
				for _, k := range keys {
					_ = b.Delete(k)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("delete migrated keys %s: %w", bucketName, err)
			}

			processed += len(batch)
			if progressCb != nil {
				progressCb(bucketName, processed, total)
			}
		}

		// Ensure bucket is recreated empty in bbolt
		_ = db.Update(func(tx *bbolt.Tx) error {
			_ = tx.DeleteBucket([]byte(bucketName))
			_, _ = tx.CreateBucketIfNotExists([]byte(bucketName))
			return nil
		})
	}

	// Migrate pollingLogs
	for {
		var subBucketName []byte
		var pBatch []*PollingLogEnt
		var pKeys [][]byte

		err := db.View(func(tx *bbolt.Tx) error {
			pb := tx.Bucket([]byte("pollingLogs"))
			if pb == nil {
				return nil
			}
			c := pb.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if v == nil { // sub-bucket
					sub := pb.Bucket(k)
					if sub == nil {
						continue
					}
					sc := sub.Cursor()
					fk, _ := sc.First()
					if fk == nil {
						subBucketName = append([]byte(nil), k...)
						return nil
					}
					subBucketName = append([]byte(nil), k...)
					for sk, sv := sc.First(); sk != nil && len(pBatch) < batchSize; sk, sv = sc.Next() {
						var pe PollingLogEnt
						if err := json.Unmarshal(sv, &pe); err != nil {
							continue
						}
						pBatch = append(pBatch, &pe)
						pKeys = append(pKeys, append([]byte(nil), sk...))
					}
					return nil
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("read pollingLogs: %w", err)
		}

		if subBucketName == nil {
			break
		}

		if len(pBatch) > 0 {
			if err := logStore.SavePollingLogs(pBatch); err != nil {
				return fmt.Errorf("save parquet polling logs: %w", err)
			}
			processed += len(pBatch)
			if progressCb != nil {
				progressCb("pollingLogs", processed, total)
			}

			err = db.Update(func(tx *bbolt.Tx) error {
				pb := tx.Bucket([]byte("pollingLogs"))
				if pb == nil {
					return nil
				}
				sub := pb.Bucket(subBucketName)
				if sub == nil {
					return nil
				}
				for _, k := range pKeys {
					_ = sub.Delete(k)
				}
				sc := sub.Cursor()
				if k, _ := sc.First(); k == nil {
					_ = pb.DeleteBucket(subBucketName)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("delete migrated polling keys: %w", err)
			}
		} else {
			// Sub-bucket is empty, delete it
			_ = db.Update(func(tx *bbolt.Tx) error {
				pb := tx.Bucket([]byte("pollingLogs"))
				if pb != nil {
					_ = pb.DeleteBucket(subBucketName)
				}
				return nil
			})
		}
	}

	// Recreate empty pollingLogs bucket in bbolt
	_ = db.Update(func(tx *bbolt.Tx) error {
		_ = tx.DeleteBucket([]byte("pollingLogs"))
		_, _ = tx.CreateBucketIfNotExists([]byte("pollingLogs"))
		return nil
	})

	// Flush parquet logs to disk
	if err := logStore.Flush(); err != nil {
		return fmt.Errorf("flush parquet log store: %w", err)
	}

	// Set and persist MapConf.LogFormat = "parquet"
	MapConf.LogFormat = "parquet"
	if err := SaveMapConf(); err != nil {
		log.Printf("save map conf err=%v", err)
	}

	if progressCb != nil {
		progressCb("done", total, total)
	}

	return nil
}
