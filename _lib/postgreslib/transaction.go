package postgreslib

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var _ QueryEngine = (*Transaction)(nil)

type Transaction struct {
	pgx.Tx
}

func (t *Transaction) Getx(ctx context.Context, dest interface{}, sqlizer Sqlizer) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PostgreSQL/Getx")
	defer span.Finish()

	query, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}

	span.SetTag("db.statement", query)
	span.SetTag("db.system", "postgresql")
	span.SetTag("db.transaction", true) // ← Помечаем что это в транзакции

	err = pgxscan.Get(ctx, t.Tx, dest, query, args...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return err
}

func (t *Transaction) Selectx(ctx context.Context, dest interface{}, sqlizer Sqlizer) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PostgreSQL/Selectx")
	defer span.Finish()

	query, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}

	span.SetTag("db.statement", query)
	span.SetTag("db.system", "postgresql")
	span.SetTag("db.transaction", true)

	err = pgxscan.Select(ctx, t.Tx, dest, query, args...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return err
}

func (t *Transaction) Execx(ctx context.Context, sqlizer Sqlizer) (pgconn.CommandTag, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PostgreSQL/Execx")
	defer span.Finish()

	query, args, err := sqlizer.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}

	span.SetTag("db.statement", query)
	span.SetTag("db.system", "postgresql")
	span.SetTag("db.transaction", true)

	tag, err := t.Tx.Exec(ctx, query, args...)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return tag, err
}
