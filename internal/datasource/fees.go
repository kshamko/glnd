package datasource

import (
	"context"
	"database/sql"
	"errors"
	"time"

	// pg lib.
	_ "github.com/lib/pq"
)

type Fee struct {
	Date time.Time
	Fee  float64
}

// ErrNotFound error.
var ErrNotFound = errors.New("data not found")

type FeesDS struct {
	db *sql.DB
}

func NewFeesDS(dsn string) (*FeesDS, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &FeesDS{db}, nil
}

// GetFees func.
func (f *FeesDS) GetFees(ctx context.Context) ([]Fee, error) {
	q := "SELECT date_trunc('hour', tr.block_time) as hr, sum(tr.gas_used * tr.gas_price/(10^18)) "
	q += "FROM transactions tr "
	q += "WHERE tr.from != '0x0000000000000000000000000000000000000000' AND "
	q += "tr.to != '0x0000000000000000000000000000000000000000' AND "
	q += "tr.to NOT IN (select address from contracts) AND "
	q += "tr.from NOT IN (select address from contracts) GROUP BY hr ORDER BY hr ASC"

	rows, err := f.db.QueryContext(ctx, q)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return []Fee{}, ErrNotFound
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []Fee{}

	for rows.Next() {
		fee := Fee{}
		err = rows.Scan(&fee.Date, &fee.Fee)

		if err != nil {
			return nil, err
		}

		result = append(result, fee)
	}

	return result, nil
}
