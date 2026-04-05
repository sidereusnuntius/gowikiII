package sqlhelpers

import (
	"database/sql"
	"fmt"

	"github.com/sidereusnuntius/gowiki/internal/wikierr"
)

func HandleErr(err error) error {
	switch err {
	case sql.ErrNoRows:
		fmt.Println("not found")
		return wikierr.ErrNotFound
	default:
		return err
	}
}
