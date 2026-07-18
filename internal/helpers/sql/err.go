package sqlhelpers

import (
	"database/sql"

	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

func HandleErr(err error) error {
	switch err {
	case sql.ErrNoRows:
		return wikierr.New(wikierr.ErrNotFound, "")
	default:
		return err
	}
}
