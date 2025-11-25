package database

import (
	"bytes"
	"encoding/binary"
	"sync"
)

// Database configuration.
type DatabaseConfig struct {
	Path string
}

type Record interface {
	SetRecordId(uint64)
	GetRecordId() uint64
	SetUint64(uint8, uint64)
	GetUint64(uint8) uint64
	SetUint8(uint8, uint8)
	GetUint8(uint8) uint8
	SetField(uint8, []byte)
	GetField(uint8) []byte
}

// Database record.
type inMemoryRecord struct {
	RecordId uint64
	Fields   [][]byte
}

func (record *inMemoryRecord) SetRecordId(recordId uint64) {
	record.RecordId = recordId
}

func (record *inMemoryRecord) GetRecordId() uint64 {
	return record.RecordId
}

func (record *inMemoryRecord) SetUint64(index uint8, value uint64) {
	field := make([]byte, 8)
	binary.BigEndian.PutUint64(field, value)
	record.Fields[index] = field
}

func (record *inMemoryRecord) GetUint64(index uint8) uint64 {
	return binary.BigEndian.Uint64(record.Fields[index])
}

func (record *inMemoryRecord) SetUint8(index uint8, value uint8) {
	record.Fields[index] = []byte{value}
}

func (record *inMemoryRecord) GetUint8(index uint8) uint8 {
	return uint8(record.Fields[index][0])
}

func (record *inMemoryRecord) SetField(index uint8, field []byte) {
	record.Fields[index] = field
}

func (record *inMemoryRecord) GetField(index uint8) []byte {
	return record.Fields[index]
}

func NewRecord(size uint8) Record {
	return &inMemoryRecord{Fields: make([][]byte, size)}
}

type queryCondition interface {
	match(Record) bool
}

type queryContext struct {
	Conditions []queryCondition
	Limit      uint16
}

type Option interface {
	apply(*queryContext)
}

type equalsOption struct {
	Field uint8
	Value []byte
}

func (opt *equalsOption) match(record Record) bool {
	return bytes.Equal(record.GetField(opt.Field), opt.Value)
}

func (opt *equalsOption) apply(ctx *queryContext) {
	ctx.Conditions = append(ctx.Conditions, opt)
}

func Eq(field uint8, value []byte) Option {
	return &equalsOption{Field: field, Value: value}
}

type limitOption struct {
	Limit uint16
}

func (opt *limitOption) apply(ctx *queryContext) {
	ctx.Limit = opt.Limit
}

func Limit(limit uint16) Option {
	return &limitOption{Limit: limit}
}

type Table interface {
	// Returns table name.
	Name() string
	// Returns records by options.
	Get(...Option) ([]Record, error)
	// Sets record to database.
	Set(Record) error
	// Deletes records by options.
	Del(...Option) error
}

type inMemoryTable struct {
	name    string
	lock    *sync.RWMutex
	records []Record
}

func (tb *inMemoryTable) Name() string {
	return tb.name
}

// Returns records by options.
func (tb *inMemoryTable) Get(options ...Option) ([]Record, error) {
	tb.lock.RLock()
	defer tb.lock.RUnlock()

	ctx := &queryContext{}
	for _, option := range options {
		option.apply(ctx)
	}

	count := uint16(0)
	result := []Record{}
	for _, record := range tb.records {
		match := true
		for _, cond := range ctx.Conditions {
			match = match && cond.match(record)
		}

		if match {
			result = append(result, record)
			count++
		}

		if count == ctx.Limit {
			break
		}
	}
	return result, nil
}

// Sets record to database.
func (tb *inMemoryTable) Set(record Record) error {
	tb.lock.Lock()
	defer tb.lock.Unlock()

	for idx := range tb.records {
		if tb.records[idx].GetRecordId() == record.GetRecordId() {
			tb.records[idx] = record
			return nil
		}
	}

	tb.records = append(tb.records, record)
	return nil
}

// Deletes records by options.
func (tb *inMemoryTable) Del(options ...Option) error {
	tb.lock.Lock()
	defer tb.lock.Unlock()

	ctx := &queryContext{}
	for _, option := range options {
		option.apply(ctx)
	}

	result := make([]Record, len(tb.records))
	for _, record := range tb.records {
		match := true
		for _, cond := range ctx.Conditions {
			match = match && cond.match(record)
		}

		if !match {
			result = append(result, record)
		}
	}
	tb.records = result
	return nil
}

// Common database of netfs server.
type Database interface {
	// Returns database table.
	Table(string) Table
	// Starts database.
	Start() error
	// Stops database.
	Stop() error
}

// Create new instance of database.
func NewDatabase(config DatabaseConfig) Database {
	return &inMemoryDatabase{lock: &sync.RWMutex{}, tables: []Table{}}
}

// Simple in memory database.
// TODO. Replace to persistable storage.
type inMemoryDatabase struct {
	lock   *sync.RWMutex
	tables []Table
}

// Returns database table.
func (db *inMemoryDatabase) Table(name string) Table {
	db.lock.Lock()
	defer db.lock.Unlock()

	for _, table := range db.tables {
		if table.Name() == name {
			return table
		}
	}

	table := &inMemoryTable{name: name, lock: &sync.RWMutex{}, records: []Record{}}
	db.tables = append(db.tables, table)
	return table
}

// Starts database.
func (*inMemoryDatabase) Start() error {
	return nil
}

// Stops database.
func (*inMemoryDatabase) Stop() error {
	return nil
}
