package datastore

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/parquet-go/parquet-go"
)

const maxCompactedRecordsPerFile = 500000

// ParquetLogRecord defines the schema for parquet log storage.
type ParquetLogRecord struct {
	Time      int64  `parquet:"time,snappy"`
	Timestamp int64  `parquet:"timestamp,timestamp(nanosecond),snappy"`
	Type      string `parquet:"type,dict,snappy"`
	Src       string `parquet:"src,dict,snappy"`
	Log       string `parquet:"log,zstd"`
}

// ParquetLogDataStore manages log persistence in Parquet files.
type ParquetLogDataStore struct {
	dirPath        string
	bufferSize     int
	bufferInterval time.Duration

	mu          sync.Mutex
	typeBuffers map[string][]*ParquetLogRecord
	bufferCount int

	stopCh chan struct{}
	doneCh chan struct{}
}

// NewParquetLogDataStore creates a new ParquetLogDataStore.
func NewParquetLogDataStore() *ParquetLogDataStore {
	return &ParquetLogDataStore{
		typeBuffers: make(map[string][]*ParquetLogRecord),
	}
}

// Open initializes the parquet store directory and background flush loop.
func (s *ParquetLogDataStore) Open(path string) error {
	if path == "" {
		return fmt.Errorf("empty parquet directory path")
	}
	s.dirPath = path
	if err := os.MkdirAll(s.dirPath, 0755); err != nil {
		return fmt.Errorf("create parquet directory: %w", err)
	}

	s.bufferSize = 5000
	s.bufferInterval = 10 * time.Second

	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	go s.flushLoop()

	return nil
}

func (s *ParquetLogDataStore) flushLoop() {
	defer close(s.doneCh)
	ticker := time.NewTicker(s.bufferInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = s.Flush()
		case <-s.stopCh:
			return
		}
	}
}

// Close flushes all buffered records and terminates background routines.
func (s *ParquetLogDataStore) Close() error {
	if s.stopCh != nil {
		close(s.stopCh)
		<-s.doneCh
		s.stopCh = nil
	}
	return s.Flush()
}

// extractSrc extracts source information for dictionary indexing.
func extractSrc(t, logStr string) string {
	if logStr == "" {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(logStr), &m); err != nil {
		return ""
	}
	switch t {
	case "syslog":
		if h, ok := m["hostname"].(string); ok && h != "" {
			return h
		}
		if h, ok := m["host"].(string); ok && h != "" {
			return h
		}
	case "trap":
		if fa, ok := m["FromAddress"].(string); ok && fa != "" {
			return fa
		}
	case "netflow":
		if sa, ok := m["SrcAddr"].(string); ok && sa != "" {
			return sa
		}
	case "sflow":
		if sa, ok := m["SrcAddr"].(string); ok && sa != "" {
			return sa
		}
	case "arplog":
		if ip, ok := m["IP"].(string); ok && ip != "" {
			return ip
		}
	}
	return ""
}

// SaveLogs queues general log entries for writing.
func (s *ParquetLogDataStore) SaveLogs(t string, logs []*LogEnt) error {
	if len(logs) == 0 {
		return nil
	}
	records := make([]*ParquetLogRecord, len(logs))
	for i, l := range logs {
		src := extractSrc(l.Type, l.Log)
		records[i] = &ParquetLogRecord{
			Time:      l.Time,
			Timestamp: l.Time,
			Type:      l.Type,
			Src:       src,
			Log:       l.Log,
		}
	}

	s.mu.Lock()
	s.typeBuffers[t] = append(s.typeBuffers[t], records...)
	s.bufferCount += len(records)
	needFlush := s.bufferCount >= s.bufferSize
	s.mu.Unlock()

	if needFlush {
		return s.Flush()
	}
	return nil
}

type pollingLogPayload struct {
	State  string                 `json:"State"`
	Result map[string]interface{} `json:"Result"`
}

// SavePollingLogs queues polling log entries for writing.
func (s *ParquetLogDataStore) SavePollingLogs(logs []*PollingLogEnt) error {
	if len(logs) == 0 {
		return nil
	}
	records := make([]*ParquetLogRecord, 0, len(logs))
	for _, l := range logs {
		b, err := json.Marshal(&pollingLogPayload{
			State:  l.State,
			Result: l.Result,
		})
		if err != nil {
			continue
		}
		records = append(records, &ParquetLogRecord{
			Time:      l.Time,
			Timestamp: l.Time,
			Type:      "polling",
			Src:       l.PollingID,
			Log:       string(b),
		})
	}

	s.mu.Lock()
	s.typeBuffers["polling"] = append(s.typeBuffers["polling"], records...)
	s.bufferCount += len(records)
	needFlush := s.bufferCount >= s.bufferSize
	s.mu.Unlock()

	if needFlush {
		return s.Flush()
	}
	return nil
}

// Flush writes all pending memory buffers to disk partition files.
func (s *ParquetLogDataStore) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.bufferCount == 0 {
		return nil
	}

	for t, records := range s.typeBuffers {
		if len(records) == 0 {
			continue
		}

		// Group records by date (YYYY-MM-DD)
		grouped := make(map[string][]*ParquetLogRecord)
		for _, r := range records {
			d := time.Unix(0, r.Time).Format("2006-01-02")
			grouped[d] = append(grouped[d], r)
		}

		for dateStr, dateRecords := range grouped {
			partDir := filepath.Join(s.dirPath, "type="+t, "date="+dateStr)
			if err := os.MkdirAll(partDir, 0755); err != nil {
				return fmt.Errorf("create parquet partition directory: %w", err)
			}

			recs := make([]ParquetLogRecord, len(dateRecords))
			for i, r := range dateRecords {
				recs[i] = *r
			}

			var randBytes [4]byte
			_, _ = rand.Read(randBytes[:])
			randVal := binary.BigEndian.Uint32(randBytes[:])

			fileName := fmt.Sprintf("part_%016x_%08x.parquet", time.Now().UnixNano(), randVal)
			filePath := filepath.Join(partDir, fileName)

			file, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("create parquet file: %w", err)
			}

			writer := parquet.NewGenericWriter[ParquetLogRecord](file)
			if _, err := writer.Write(recs); err != nil {
				_ = file.Close()
				return fmt.Errorf("write parquet records: %w", err)
			}
			if err := writer.Close(); err != nil {
				_ = file.Close()
				return fmt.Errorf("close parquet writer: %w", err)
			}
			_ = file.Close()
		}
	}

	s.typeBuffers = make(map[string][]*ParquetLogRecord)
	s.bufferCount = 0
	return nil
}

// ForEachLog iterates over log entries in chronological order within [st, et].
func (s *ParquetLogDataStore) ForEachLog(t string, st, et int64, callBack func(log *LogEnt) bool) {
	_ = s.Flush()

	if et == 0 {
		et = time.Now().UnixNano()
	}

	var typeDirs []string
	if t == "all" || t == "" {
		pattern := filepath.Join(s.dirPath, "type=*")
		matches, _ := filepath.Glob(pattern)
		typeDirs = matches
		if len(typeDirs) == 0 {
			typeDirs = []string{s.dirPath}
		}
	} else {
		typeDirs = []string{filepath.Join(s.dirPath, "type=" + t)}
	}

	for _, tDir := range typeDirs {
		dateDirs, _ := filepath.Glob(filepath.Join(tDir, "date=*"))
		if len(dateDirs) == 0 {
			s.scanParquetDir(tDir, st, et, callBack)
			continue
		}

		sort.Strings(dateDirs)
		startDateStr := time.Unix(0, st).Format("2006-01-02")
		endDateStr := time.Unix(0, et).Format("2006-01-02")

		for _, dDir := range dateDirs {
			base := filepath.Base(dDir)
			dStr := strings.TrimPrefix(base, "date=")
			if dStr < startDateStr || dStr > endDateStr {
				continue
			}
			if !s.scanParquetDir(dDir, st, et, callBack) {
				return
			}
		}
	}
}

func (s *ParquetLogDataStore) scanParquetDir(dir string, st, et int64, callBack func(log *LogEnt) bool) bool {
	files, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	if err != nil || len(files) == 0 {
		return true
	}
	sort.Strings(files)

	for _, filePath := range files {
		if strings.HasPrefix(filepath.Base(filePath), "compacting_") {
			continue
		}

		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		fi, err := f.Stat()
		if err != nil || fi.Size() == 0 {
			_ = f.Close()
			continue
		}

		pf, err := parquet.OpenFile(f, fi.Size())
		if err != nil {
			_ = f.Close()
			continue
		}

		reader := parquet.NewGenericReader[ParquetLogRecord](pf)
		buf := make([]ParquetLogRecord, 1024)
		stop := false

		for {
			n, err := reader.Read(buf)
			if n > 0 {
				for i := 0; i < n; i++ {
					rec := &buf[i]
					if rec.Time < st {
						continue
					}
					if rec.Time > et {
						continue
					}
					entry := &LogEnt{
						Time: rec.Time,
						Type: rec.Type,
						Log:  rec.Log,
					}
					if !callBack(entry) {
						stop = true
						break
					}
				}
			}
			if stop || err != nil {
				break
			}
		}
		_ = reader.Close()
		_ = f.Close()

		if stop {
			return false
		}
	}
	return true
}

// ForEachLastLog iterates over log entries in reverse chronological order (newest first).
func (s *ParquetLogDataStore) ForEachLastLog(t string, callBack func(log *LogEnt) bool) {
	_ = s.Flush()

	tDir := filepath.Join(s.dirPath, "type="+t)
	dateDirs, _ := filepath.Glob(filepath.Join(tDir, "date=*"))
	if len(dateDirs) == 0 {
		return
	}

	// Sort date directories in descending order (newest date first)
	sort.Sort(sort.Reverse(sort.StringSlice(dateDirs)))

	for _, dDir := range dateDirs {
		if !s.scanParquetDirReverse(dDir, callBack) {
			return
		}
	}
}

func (s *ParquetLogDataStore) scanParquetDirReverse(dir string, callBack func(log *LogEnt) bool) bool {
	files, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	if err != nil || len(files) == 0 {
		return true
	}
	// Sort files in descending order (newest file first)
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, filePath := range files {
		if strings.HasPrefix(filepath.Base(filePath), "compacting_") {
			continue
		}

		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		fi, err := f.Stat()
		if err != nil || fi.Size() == 0 {
			_ = f.Close()
			continue
		}

		pf, err := parquet.OpenFile(f, fi.Size())
		if err != nil {
			_ = f.Close()
			continue
		}

		reader := parquet.NewGenericReader[ParquetLogRecord](pf)
		var allRecords []ParquetLogRecord
		buf := make([]ParquetLogRecord, 1024)

		for {
			n, err := reader.Read(buf)
			if n > 0 {
				allRecords = append(allRecords, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		_ = reader.Close()
		_ = f.Close()

		// Traverse in reverse order (newest record first)
		for i := len(allRecords) - 1; i >= 0; i-- {
			rec := &allRecords[i]
			entry := &LogEnt{
				Time: rec.Time,
				Type: rec.Type,
				Log:  rec.Log,
			}
			if !callBack(entry) {
				return false
			}
		}
	}
	return true
}

// ForEachLastPollingLog retrieves polling logs in reverse order for a specific polling ID.
func (s *ParquetLogDataStore) ForEachLastPollingLog(pollingID string, callBack func(log *PollingLogEnt) bool) {
	_ = s.Flush()

	tDir := filepath.Join(s.dirPath, "type=polling")
	dateDirs, _ := filepath.Glob(filepath.Join(tDir, "date=*"))
	if len(dateDirs) == 0 {
		return
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dateDirs)))

	for _, dDir := range dateDirs {
		files, err := filepath.Glob(filepath.Join(dDir, "*.parquet"))
		if err != nil || len(files) == 0 {
			continue
		}
		sort.Sort(sort.Reverse(sort.StringSlice(files)))

		for _, filePath := range files {
			if strings.HasPrefix(filepath.Base(filePath), "compacting_") {
				continue
			}

			f, err := os.Open(filePath)
			if err != nil {
				continue
			}
			fi, err := f.Stat()
			if err != nil || fi.Size() == 0 {
				_ = f.Close()
				continue
			}
			pf, err := parquet.OpenFile(f, fi.Size())
			if err != nil {
				_ = f.Close()
				continue
			}

			reader := parquet.NewGenericReader[ParquetLogRecord](pf)
			var allRecords []ParquetLogRecord
			buf := make([]ParquetLogRecord, 1024)
			for {
				n, err := reader.Read(buf)
				if n > 0 {
					allRecords = append(allRecords, buf[:n]...)
				}
				if err != nil {
					break
				}
			}
			_ = reader.Close()
			_ = f.Close()

			for i := len(allRecords) - 1; i >= 0; i-- {
				rec := &allRecords[i]
				if rec.Src != pollingID {
					continue
				}
				var payload pollingLogPayload
				if err := json.Unmarshal([]byte(rec.Log), &payload); err != nil {
					continue
				}
				ent := &PollingLogEnt{
					Time:      rec.Time,
					PollingID: rec.Src,
					State:     payload.State,
					Result:    payload.Result,
				}
				if !callBack(ent) {
					return
				}
			}
		}
	}
}

// GetAllPollingLog retrieves all logs for a given polling ID in chronological order.
func (s *ParquetLogDataStore) GetAllPollingLog(pollingID string) []*PollingLogEnt {
	_ = s.Flush()

	ret := []*PollingLogEnt{}
	tDir := filepath.Join(s.dirPath, "type=polling")
	dateDirs, _ := filepath.Glob(filepath.Join(tDir, "date=*"))
	if len(dateDirs) == 0 {
		return ret
	}

	sort.Strings(dateDirs)

	for _, dDir := range dateDirs {
		files, err := filepath.Glob(filepath.Join(dDir, "*.parquet"))
		if err != nil || len(files) == 0 {
			continue
		}
		sort.Strings(files)

		for _, filePath := range files {
			if strings.HasPrefix(filepath.Base(filePath), "compacting_") {
				continue
			}

			f, err := os.Open(filePath)
			if err != nil {
				continue
			}
			fi, err := f.Stat()
			if err != nil || fi.Size() == 0 {
				_ = f.Close()
				continue
			}
			pf, err := parquet.OpenFile(f, fi.Size())
			if err != nil {
				_ = f.Close()
				continue
			}

			reader := parquet.NewGenericReader[ParquetLogRecord](pf)
			buf := make([]ParquetLogRecord, 1024)
			for {
				n, err := reader.Read(buf)
				if n > 0 {
					for i := 0; i < n; i++ {
						rec := &buf[i]
						if rec.Src != pollingID {
							continue
						}
						var payload pollingLogPayload
						if err := json.Unmarshal([]byte(rec.Log), &payload); err != nil {
							continue
						}
						ret = append(ret, &PollingLogEnt{
							Time:      rec.Time,
							PollingID: rec.Src,
							State:     payload.State,
							Result:    payload.Result,
						})
					}
				}
				if err != nil {
					break
				}
			}
			_ = reader.Close()
			_ = f.Close()
		}
	}
	return ret
}

// ClearLog clears logs of a specified type or all types.
func (s *ParquetLogDataStore) ClearLog(t string) {
	s.mu.Lock()
	if t == "all" {
		s.typeBuffers = make(map[string][]*ParquetLogRecord)
		s.bufferCount = 0
	} else {
		if len(s.typeBuffers[t]) > 0 {
			s.bufferCount -= len(s.typeBuffers[t])
			delete(s.typeBuffers, t)
		}
	}
	s.mu.Unlock()

	if t == "all" {
		typeDirs, _ := filepath.Glob(filepath.Join(s.dirPath, "type=*"))
		for _, d := range typeDirs {
			_ = os.RemoveAll(d)
		}
		files, _ := filepath.Glob(filepath.Join(s.dirPath, "*.parquet"))
		for _, f := range files {
			_ = os.Remove(f)
		}
		return
	}

	targetDir := filepath.Join(s.dirPath, "type="+t)
	_ = os.RemoveAll(targetDir)
}

// Cleanup removes expired log date folders where date < cutoffDate.
func (s *ParquetLogDataStore) Cleanup(retentionDays int) error {
	_ = s.Flush()

	if retentionDays < 1 {
		retentionDays = 1
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	cutoffDateStr := cutoffTime.Format("2006-01-02")

	typeDirs, _ := filepath.Glob(filepath.Join(s.dirPath, "type=*"))
	for _, tDir := range typeDirs {
		dateDirs, _ := filepath.Glob(filepath.Join(tDir, "date=*"))
		for _, dDir := range dateDirs {
			base := filepath.Base(dDir)
			dStr := strings.TrimPrefix(base, "date=")
			if dStr < cutoffDateStr {
				_ = os.RemoveAll(dDir)
			}
		}
	}
	return nil
}

// Compact merges multiple parquet files in past date directories into compacted files.
func (s *ParquetLogDataStore) Compact(currentDate string) error {
	_ = s.Flush()

	if currentDate == "" {
		currentDate = time.Now().Format("2006-01-02")
	}

	typeDirs, _ := filepath.Glob(filepath.Join(s.dirPath, "type=*"))
	for _, tDir := range typeDirs {
		dateDirs, _ := filepath.Glob(filepath.Join(tDir, "date=*"))
		for _, dDir := range dateDirs {
			base := filepath.Base(dDir)
			dStr := strings.TrimPrefix(base, "date=")
			if dStr >= currentDate {
				continue
			}
			if err := s.compactDateDir(dDir, maxCompactedRecordsPerFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ParquetLogDataStore) compactDateDir(dir string, maxRecords int) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.parquet"))
	if err != nil || len(files) <= 1 {
		return nil
	}

	var validFiles []string
	for _, filePath := range files {
		if strings.HasPrefix(filepath.Base(filePath), "compacting_") {
			_ = os.Remove(filePath)
		} else {
			validFiles = append(validFiles, filePath)
		}
	}

	if len(validFiles) <= 1 {
		return nil
	}
	sort.Strings(validFiles)

	var createdTempFiles []string
	var createdFinalFiles []string

	var currentWriter *parquet.GenericWriter[ParquetLogRecord]
	var currentFile *os.File
	var currentCount int
	seq := 0

	closeCurrentWriter := func() error {
		if currentWriter != nil {
			if err := currentWriter.Close(); err != nil {
				_ = currentFile.Close()
				return err
			}
			if err := currentFile.Close(); err != nil {
				return err
			}
			currentWriter = nil
			currentFile = nil
		}
		return nil
	}

	openNewWriter := func() error {
		if err := closeCurrentWriter(); err != nil {
			return err
		}

		var randBytes [4]byte
		_, _ = rand.Read(randBytes[:])
		randVal := binary.BigEndian.Uint32(randBytes[:])
		nowNano := time.Now().UnixNano()

		tmpFileName := fmt.Sprintf("compacting_%016x_%04x_%08x.parquet", nowNano, seq, randVal)
		finalFileName := fmt.Sprintf("compacted_%016x_%04x_%08x.parquet", nowNano, seq, randVal)
		seq++

		tmpFilePath := filepath.Join(dir, tmpFileName)
		finalFilePath := filepath.Join(dir, finalFileName)

		outFile, err := os.Create(tmpFilePath)
		if err != nil {
			return fmt.Errorf("create compacted parquet file: %w", err)
		}
		currentFile = outFile
		currentWriter = parquet.NewGenericWriter[ParquetLogRecord](outFile)
		currentCount = 0

		createdTempFiles = append(createdTempFiles, tmpFilePath)
		createdFinalFiles = append(createdFinalFiles, finalFilePath)
		return nil
	}

	defer func() {
		_ = closeCurrentWriter()
	}()

	buf := make([]ParquetLogRecord, 1024)
	totalProcessed := 0

	for _, filePath := range validFiles {
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}
		fi, err := f.Stat()
		if err != nil || fi.Size() == 0 {
			_ = f.Close()
			continue
		}
		pf, err := parquet.OpenFile(f, fi.Size())
		if err != nil {
			_ = f.Close()
			continue
		}

		reader := parquet.NewGenericReader[ParquetLogRecord](pf)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				if currentWriter == nil {
					if err := openNewWriter(); err != nil {
						_ = reader.Close()
						_ = f.Close()
						cleanupTempFiles(createdTempFiles)
						return err
					}
				}

				if _, writeErr := currentWriter.Write(buf[:n]); writeErr != nil {
					_ = reader.Close()
					_ = f.Close()
					cleanupTempFiles(createdTempFiles)
					return fmt.Errorf("write streaming parquet records: %w", writeErr)
				}
				currentCount += n
				totalProcessed += n

				if maxRecords > 0 && currentCount >= maxRecords {
					if err := openNewWriter(); err != nil {
						_ = reader.Close()
						_ = f.Close()
						cleanupTempFiles(createdTempFiles)
						return err
					}
				}
			}
			if err != nil {
				break
			}
		}
		_ = reader.Close()
		_ = f.Close()
	}

	if err := closeCurrentWriter(); err != nil {
		cleanupTempFiles(createdTempFiles)
		return err
	}

	if totalProcessed == 0 {
		cleanupTempFiles(createdTempFiles)
		return nil
	}

	for _, oldFile := range validFiles {
		_ = os.Remove(oldFile)
	}

	for i := range createdTempFiles {
		_ = os.Rename(createdTempFiles[i], createdFinalFiles[i])
	}

	return nil
}

func cleanupTempFiles(files []string) {
	for _, f := range files {
		_ = os.Remove(f)
	}
}

// Size returns total bytes consumed by the parquet logs directory.
func (s *ParquetLogDataStore) Size() int64 {
	var totalSize int64
	if s.dirPath == "" {
		return 0
	}
	_ = filepath.Walk(s.dirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize
}
