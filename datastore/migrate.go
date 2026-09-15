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
func GetBboltLogStats() BboltLogStats {
	stats := BboltLogStats{}
	if db == nil {
		return stats
	}

	_ = db.View(func(tx *bbolt.Tx) error {
		stats.BboltSize = tx.Size()

		countBucket := func(name string) int {
			b := tx.Bucket([]byte(name))
			if b == nil {
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
					stats.PollingLogCount += sub.Stats().KeyN
				}
				return nil
			})
		}
		return nil
	})

	stats.TotalCount = stats.SyslogCount + stats.TrapCount + stats.NetflowCount +
		stats.SFlowCount + stats.SFlowCounterCount + stats.ArpLogCount + stats.PollingLogCount

	if logStore != nil {
		stats.ParquetSize = logStore.Size()
	}

	return stats
}

// MigrateBboltToParquet migrates all log data from bbolt to the Parquet log store.
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

	stats := GetBboltLogStats()
	total := stats.TotalCount
	processed := 0

	generalBuckets := []string{"syslog", "trap", "netflow", "sflow", "sflowCounter", "arplog"}

	for _, bucketName := range generalBuckets {
		var entries []*LogEnt
		err := db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				return nil
			}
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				v = deCompressLog(v)
				var l LogEnt
				if err := json.Unmarshal(v, &l); err != nil {
					continue
				}
				entries = append(entries, &l)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("read bucket %s: %w", bucketName, err)
		}

		if len(entries) > 0 {
			// Write to parquet in batches of 5000
			batchSize := 5000
			for i := 0; i < len(entries); i += batchSize {
				end := i + batchSize
				if end > len(entries) {
					end = len(entries)
				}
				batch := entries[i:end]
				if err := logStore.SaveLogs(bucketName, batch); err != nil {
					return fmt.Errorf("save parquet logs %s: %w", bucketName, err)
				}
				processed += len(batch)
				if progressCb != nil {
					progressCb(bucketName, processed, total)
				}
			}
		}

		// Empty bbolt bucket after successful parquet write
		_ = db.Batch(func(tx *bbolt.Tx) error {
			_ = tx.DeleteBucket([]byte(bucketName))
			_, _ = tx.CreateBucketIfNotExists([]byte(bucketName))
			return nil
		})
	}

	// Migrate pollingLogs
	var pEntries []*PollingLogEnt
	err := db.View(func(tx *bbolt.Tx) error {
		pb := tx.Bucket([]byte("pollingLogs"))
		if pb == nil {
			return nil
		}
		return pb.ForEachBucket(func(k []byte) error {
			sub := pb.Bucket(k)
			if sub == nil {
				return nil
			}
			c := sub.Cursor()
			for sk, sv := c.First(); sk != nil; sk, sv = c.Next() {
				var pe PollingLogEnt
				if err := json.Unmarshal(sv, &pe); err != nil {
					continue
				}
				pEntries = append(pEntries, &pe)
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("read pollingLogs: %w", err)
	}

	if len(pEntries) > 0 {
		batchSize := 5000
		for i := 0; i < len(pEntries); i += batchSize {
			end := i + batchSize
			if end > len(pEntries) {
				end = len(pEntries)
			}
			batch := pEntries[i:end]
			if err := logStore.SavePollingLogs(batch); err != nil {
				return fmt.Errorf("save parquet polling logs: %w", err)
			}
			processed += len(batch)
			if progressCb != nil {
				progressCb("pollingLogs", processed, total)
			}
		}
	}

	// Empty pollingLogs bucket in bbolt
	_ = db.Batch(func(tx *bbolt.Tx) error {
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
