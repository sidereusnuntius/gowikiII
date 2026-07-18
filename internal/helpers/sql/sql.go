package sqlhelpers

import (
	"database/sql"
	"time"
)

type Scanner interface {
	Scan(dest ...any) error
}

func NullableString(str string) sql.NullString {
	return sql.NullString{
		String: str,
		Valid:  len(str) > 0,
	}
}

func NullableTimeUnix(t time.Time) sql.NullInt64 {
	return sql.NullInt64{
		Int64: t.Unix(),
		Valid: !t.IsZero(),
	}
}

func NullableInt64(n int64) sql.NullInt64 {
	return sql.NullInt64{
		Int64: n,
		Valid: n != 0,
	}
}
