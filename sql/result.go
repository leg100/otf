package sql

import "github.com/leg100/otf"

func convertTimestamps(ts Timestamps) otf.Timestamps {
	return otf.Timestamps{
		CreatedAt: ts.GetCreatedAt(),
		UpdatedAt: ts.GetUpdatedAt(),
	}
}
