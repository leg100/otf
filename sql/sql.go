/*
Package sql implements persistent storage using the sql database.
*/
package sql

import (
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/leg100/otf"
)

func databaseError(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case noRowsInResultError(err):
		return otf.ErrResourceNotFound
	case errors.As(err, &pgErr):
		switch pgErr.Code {
		case "23505": // unique violation
			return otf.ErrResourcesAlreadyExists
		}
		fallthrough
	default:
		return err
	}
}

func noRowsInResultError(err error) bool {
	for {
		err = errors.Unwrap(err)
		if err == nil {
			return false
		} else if err.Error() == "no rows in result set" {
			return true
		}
	}
}

func includeRelation(includes *string, relation string) bool {
	if includes != nil {
		includes := strings.Split(*includes, ",")
		for _, inc := range includes {
			if inc == relation {
				return true
			}
		}
	}
	return false
}

func includeOrganization(includes *string) bool {
	return includeRelation(includes, "organization")
}

func includeWorkspace(includes *string) bool {
	return includeRelation(includes, "workspace")
}

func includeConfigurationVersion(includes *string) bool {
	return includeRelation(includes, "configuration_version")
}
