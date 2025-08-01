package orders

import (
	"context"
	"time"

	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

const (
	ProjectionTable = "order_projection"
)

type DbProjection struct {
	OrderId        string    `db:"order_id"`
	PaymentStatus  string    `db:"payment_status"`
	ShippingStatus string    `db:"shipping_status"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type UpsertArgs struct {
	OrderId        string
	PaymentStatus  string
	ShippingStatus string
}

type ListArgs struct {
	Limit  uint
	Offset uint
}

type ProjectionRepo interface {
	Upsert(ctx context.Context, tx pg.Tx, args UpsertArgs) error

	List(ctx context.Context, args ListArgs) ([]DbProjection, error)
}

// Postgres implementation
type PgProjectionRepo struct {
	db *sqlx.DB
}

func NewPgProjectionRepo(db *sqlx.DB) ProjectionRepo {
	return &PgProjectionRepo{db: db}
}

func (r *PgProjectionRepo) Upsert(ctx context.Context, tx pg.Tx, args UpsertArgs) error {
	// Compile query
	ds := pg.Dialect.Insert(ProjectionTable).Prepared(true).
		Cols("order_id", "payment_status", "shipping_status").
		Rows([]goqu.Record{
			{
				"order_id":        args.OrderId,
				"payment_status":  args.PaymentStatus,
				"shipping_status": args.ShippingStatus,
			},
		}).
		OnConflict(goqu.DoUpdate("order_id", goqu.Record{
			"payment_status":  goqu.I("excluded.payment_status"),
			"shipping_status": goqu.I("excluded.shipping_status"),
		}))

	query, queryArgs, err := ds.ToSQL()
	if err != nil {
		return pg.ErrorDsl(err)
	}

	_, err = tx.ExecContext(ctx, query, queryArgs...)
	if err != nil {
		return pg.ErrorDb(err)
	}

	return nil
}

func (r *PgProjectionRepo) List(ctx context.Context, args ListArgs) ([]DbProjection, error) {

	ds := pg.Dialect.From(ProjectionTable).Prepared(true).
		Select(&DbProjection{}).
		Order(goqu.I("created_at").Desc()).
		Limit(args.Limit).
		Offset(args.Offset)

	query, queryArgs, err := ds.ToSQL()
	if err != nil {
		return nil, pg.ErrorDsl(err)
	}

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, pg.ErrorDb(err)
	}

	projections := []DbProjection{}
	err = sqlx.StructScan(rows, &projections)
	if err != nil {
		return nil, pg.ErrorUnmarshal(err)
	}

	return projections, nil
}
