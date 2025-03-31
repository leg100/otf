package github

import "strconv"

type AppID int64

func (id AppID) String() string { return strconv.Itoa(int(id)) }
