package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type optest struct {
	name     string
	path     string
	expected []int
}

var optests = []optest{
	{"single key (period) ", `$.aKey`, []int{opTypeName}},
	{"single key (bracket)", `$["aKey"]`, []int{opTypeName}},
	{"single key (period) ", `$.*`, []int{opTypeNameWild}},
	{"single index", `$[12]`, []int{opTypeIndex}},
	{"single key", `$[23:45]`, []int{opTypeIndexRange}},
	{"single key", `$[*]`, []int{opTypeIndexWild}},

	{"double key", `$["aKey"]["bKey"]`, []int{opTypeName, opTypeName}},
	{"double key", `$["aKey"].bKey`, []int{opTypeName, opTypeName}},
}

func TestQueryOperators(t *testing.T) {
	as := assert.New(t)

	for _, optest := range optests {
		t.Log(optest.name)
		path, err := parsePath(optest.path)
		as.NoError(err)

		as.EqualValues(len(optest.expected), len(path.operators))

		for x, op := range optest.expected {
			as.EqualValues(pathTokenNames[op], pathTokenNames[path.operators[x].typ])
		}
	}
}
