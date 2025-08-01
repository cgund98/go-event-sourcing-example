package eventsrc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

type PersistEventArgs struct {
	AggregateId   string
	AggregateType string
	EventType     string
	Data          []byte
}

type Event struct {
	EventId       int       `db:"event_id"`
	AggregateId   string    `db:"aggregate_id"`
	AggregateType string    `db:"aggregate_type"`
	EventType     string    `db:"event_type"`
	Data          []byte    `db:"event_data"`
	CreatedAt     time.Time `db:"created_at"`
}

type Store interface {
	Persist(ctx context.Context, tx pg.Tx, args PersistEventArgs) error
	ListByAggregateID(ctx context.Context, aggregateId string, aggregateType string) ([]Event, error)
}

func serializeAggregateId(aggregateId string, aggregateType string) string {
	return fmt.Sprintf("%s:%s", aggregateType, aggregateId)
}

func deserializeAggregateId(aggregateId string) (string, string) {
	parts := strings.Split(aggregateId, ":")
	return parts[1], parts[0]
}

/** Postgres Store */

type PostgresStore struct {
	db    *sqlx.DB
	table string
}

func NewPostgresStore(db *sqlx.DB, table string) *PostgresStore {
	return &PostgresStore{db: db, table: table}
}

func (s *PostgresStore) Persist(ctx context.Context, tx pg.Tx, args PersistEventArgs) error {
	// Compile query
	ds := pg.Dialect.Insert(s.table).Prepared(true).
		Cols("aggregate_id", "aggregate_type", "event_type", "event_data").
		Rows([]goqu.Record{
			{
				"aggregate_id":   serializeAggregateId(args.AggregateId, args.AggregateType),
				"aggregate_type": args.AggregateType,
				"event_type":     args.EventType,
				"event_data":     args.Data,
			},
		})

	query, queryArgs, err := ds.ToSQL()
	if err != nil {
		return pg.ErrorDsl(err)
	}

	// Execute query
	_, err = tx.ExecContext(ctx, query, queryArgs...)
	if err != nil {
		return pg.ErrorDb(err)
	}

	return nil
}

func (s *PostgresStore) ListByAggregateID(ctx context.Context, aggregateId string, aggregateType string) ([]Event, error) {
	// Compile query
	ds := pg.Dialect.From(s.table).Prepared(true).
		Select(&Event{}).
		Where(goqu.Ex{"aggregate_id": serializeAggregateId(aggregateId, aggregateType)})

	query, queryArgs, err := ds.ToSQL()
	if err != nil {
		return nil, pg.ErrorDsl(err)
	}

	rows, err := s.db.QueryxContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, pg.ErrorDb(err)
	}

	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var event Event
		err := rows.StructScan(&event)
		if err != nil {
			return nil, pg.ErrorUnmarshal(err)
		}

		events = append(events, event)
	}

	return events, nil
}

/** In-memory Store */

type InMemoryStore struct {
	Events map[string][]Event
	mu     sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{Events: make(map[string][]Event)}
}

func (s *InMemoryStore) Persist(ctx context.Context, tx pg.Tx, args PersistEventArgs) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aggregateId := serializeAggregateId(args.AggregateId, args.AggregateType)

	// Set event_id to be the length of the events slice
	eventID := len(s.Events[aggregateId])

	s.Events[aggregateId] = append(s.Events[aggregateId], Event{
		EventId:       eventID,
		AggregateId:   aggregateId,
		AggregateType: args.AggregateType,
		EventType:     args.EventType,
		Data:          args.Data,
		CreatedAt:     time.Now().UTC(),
	})

	return nil
}

func (s *InMemoryStore) ListByAggregateID(ctx context.Context, aggregateId string, aggregateType string) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.Events[serializeAggregateId(aggregateId, aggregateType)]
	// Return a copy to prevent external modification
	result := make([]Event, len(events))

	for idx := range events {
		aggregateId, _ := deserializeAggregateId(events[idx].AggregateId)
		events[idx].AggregateId = aggregateId
	}

	copy(result, events)
	return result, nil
}
