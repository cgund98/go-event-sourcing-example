package eventsrc

import (
	"context"
	"sync"
	"time"

	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

type PersistEventArgs struct {
	AggregateID   string
	AggregateType string
	EventType     string
	Data          []byte
}

type Event struct {
	ID            int
	AggregateID   string
	AggregateType string
	EventType     string
	Data          []byte
	CreatedAt     time.Time
}

type Store interface {
	Persist(ctx context.Context, tx pg.Tx, args PersistEventArgs) error
	ListByAggregateID(ctx context.Context, aggregateID string) ([]Event, error)
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
				"aggregate_id":   args.AggregateID,
				"aggregate_type": args.AggregateType,
				"event_type":     args.EventType,
				"event_data":     args.Data,
			},
		})

	query, queryArgs, err := ds.ToSQL()
	if err != nil {
		return pg.ErrorDsl(err)
	}
	logging.Logger.Info("Persisting event...", "query", query, "queryArgs", queryArgs)

	// Execute query
	_, err = tx.ExecContext(ctx, query, queryArgs...)
	if err != nil {
		return pg.ErrorDb(err)
	}

	return nil
}

func (s *PostgresStore) ListByAggregateID(ctx context.Context, aggregateID string) ([]Event, error) {
	// Compile query
	ds := pg.Dialect.From(s.table).Prepared(true).
		Select(&Event{}).
		Where(goqu.Ex{"aggregate_id": aggregateID})

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

	// Set event_id to be the length of the events slice
	eventID := len(s.Events[args.AggregateID])

	s.Events[args.AggregateID] = append(s.Events[args.AggregateID], Event{
		ID:            eventID,
		AggregateID:   args.AggregateID,
		AggregateType: args.AggregateType,
		EventType:     args.EventType,
		Data:          args.Data,
		CreatedAt:     time.Now().UTC(),
	})

	return nil
}

func (s *InMemoryStore) ListByAggregateID(ctx context.Context, aggregateID string) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.Events[aggregateID]
	// Return a copy to prevent external modification
	result := make([]Event, len(events))
	copy(result, events)
	return result, nil
}
