package sqlite

import (
	"reflect"
	"testing"

	"github.com/jmoiron/sqlx/reflectx"
	"github.com/stretchr/testify/assert"
)

func TestFindUpdates(t *testing.T) {
	m := reflectx.NewMapper("db")

	type SS struct {
		C int `db:"c"`
		D int `db:"d"`
	}

	type S struct {
		A  int `db:"a"`
		B  int `db:"b"`
		SS `db:"ss"`
	}

	before := S{A: 1, B: 2, SS: SS{C: 3, D: 4}}
	after := S{A: 1, B: 9, SS: SS{C: 3, D: 99}}

	idx := diffIndex(reflect.ValueOf(before), reflect.ValueOf(after), nil, nil)

	assert.Equal(t, [][]int{{1}, {2, 1}}, idx)

	updates := FindUpdates(m, before, after)
	assert.Equal(t, map[string]interface{}{"b": 9, "ss.d": 99}, updates)
}
