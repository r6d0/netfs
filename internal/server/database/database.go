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

// Field of the record.
type RecordField int8

type Record interface {
	SetRecordId(uint64)
	GetRecordId() uint64
	SetUint64(RecordField, uint64)
	GetUint64(RecordField) uint64
	SetUint8(RecordField, uint8)
	GetUint8(RecordField) uint8
	SetField(RecordField, []byte)
	GetField(RecordField) []byte
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

func (record *inMemoryRecord) SetUint64(index RecordField, value uint64) {
	field := make([]byte, 8)
	binary.BigEndian.PutUint64(field, value)
	record.Fields[index] = field
}

func (record *inMemoryRecord) GetUint64(index RecordField) uint64 {
	return binary.BigEndian.Uint64(record.Fields[index])
}

func (record *inMemoryRecord) SetUint8(index RecordField, value uint8) {
	record.Fields[index] = []byte{value}
}

func (record *inMemoryRecord) GetUint8(index RecordField) uint8 {
	return uint8(record.Fields[index][0])
}

func (record *inMemoryRecord) SetField(index RecordField, field []byte) {
	record.Fields[index] = field
}

func (record *inMemoryRecord) GetField(index RecordField) []byte {
	return record.Fields[index]
}

func NewRecord(size int8) Record {
	return &inMemoryRecord{Fields: make([][]byte, size)}
}

type queryCondition interface {
	match(Record) bool
}

type queryContext struct {
	Conditions []queryCondition
	Limit      int16
}

type Option interface {
	apply(*queryContext)
}

type idOption struct {
	Value uint64
}

func (opt *idOption) match(record Record) bool {
	return record.GetRecordId() == opt.Value
}

func (opt *idOption) apply(ctx *queryContext) {
	ctx.Conditions = append(ctx.Conditions, opt)
}

func Id(value uint64) Option {
	return &idOption{Value: value}
}

type eqOption struct {
	Field RecordField
	Value []byte
}

func (opt *eqOption) match(record Record) bool {
	return bytes.Equal(record.GetField(opt.Field), opt.Value)
}

func (opt *eqOption) apply(ctx *queryContext) {
	ctx.Conditions = append(ctx.Conditions, opt)
}

func Eq(field RecordField, value []byte) Option {
	return &eqOption{Field: field, Value: value}
}

type limitOption struct {
	Limit int16
}

func (opt *limitOption) apply(ctx *queryContext) {
	ctx.Limit = opt.Limit
}

func Limit(limit int16) Option {
	return &limitOption{Limit: limit}
}

type Table interface {
	// Returns table name.
	Name() string
	// Returns next internal identifier.
	NextId() uint64
	// Returns records by options.
	Get(...Option) ([]Record, error)
	// Sets records to database.
	Set(...Record) error
	// Deletes records by options.
	Del(...Option) error
}

type inMemoryTable struct {
	name         string
	lastRecordId uint64
	lock         *sync.RWMutex
	records      []Record
}

func (tb *inMemoryTable) Name() string {
	return tb.name
}

// Returns next internal identifier.
func (tb *inMemoryTable) NextId() uint64 {
	tb.lock.Lock()
	defer tb.lock.Unlock()

	tb.lastRecordId++
	return tb.lastRecordId
}

// Returns records by options.
func (tb *inMemoryTable) Get(options ...Option) ([]Record, error) {
	tb.lock.RLock()
	defer tb.lock.RUnlock()

	ctx := &queryContext{Limit: -1}
	for _, option := range options {
		option.apply(ctx)
	}

	count := int16(0)
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
func (tb *inMemoryTable) Set(records ...Record) error {
	tb.lock.Lock()
	defer tb.lock.Unlock()

	for _, record := range records {
		for idx := range tb.records {
			if tb.records[idx].GetRecordId() == record.GetRecordId() {
				tb.records[idx] = record
				return nil
			}
		}

		tb.records = append(tb.records, record)
	}
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
