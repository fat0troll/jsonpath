package jsonpath

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	name     string
	json     string
	path     string
	expected []Result
}

var tests = []test{
	{`key selection`, `{"aKey":32}`, `$.aKey+`, []Result{newResult(`32`, JSONNumber, `aKey`)}},
	{`nested key selection`, `{"aKey":{"bKey":32}}`, `$.aKey+`, []Result{newResult(`{"bKey":32}`, JSONObject, `aKey`)}},
	{`empty array`, `{"aKey":[]}`, `$.aKey+`, []Result{newResult(`[]`, JSONArray, `aKey`)}},
	{`multiple same-level keys, weird spacing`, `{    "aKey" 	: true ,    "bKey":  [	1 , 2	], "cKey" 	: true		} `, `$.bKey+`, []Result{newResult(`[1,2]`, JSONArray, `bKey`)}},

	{`array index selection`, `{"aKey":[123,456]}`, `$.aKey[1]+`, []Result{newResult(`456`, JSONNumber, `aKey`, 1)}},
	{`array wild index selection`, `{"aKey":[123,456]}`, `$.aKey[*]+`, []Result{newResult(`123`, JSONNumber, `aKey`, 0), newResult(`456`, JSONNumber, `aKey`, 1)}},
	{`array range index selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:3]+`, []Result{newResult(`22`, JSONNumber, `aKey`, 1), newResult(`33`, JSONNumber, `aKey`, 2)}},
	{`array range (no index) selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:1]+`, []Result{}},
	{`array range (no upper bound) selection`, `{"aKey":[11,22,33]}`, `$.aKey[1:]+`, []Result{newResult(`22`, JSONNumber, `aKey`, 1), newResult(`33`, JSONNumber, `aKey`, 2)}},

	{`empty array - try selection`, `{"aKey":[]}`, `$.aKey[1]+`, []Result{}},
	{`null selection`, `{"aKey":[null]}`, `$.aKey[0]+`, []Result{newResult(`null`, JSONNull, `aKey`, 0)}},
	{`empty object`, `{"aKey":{}}`, `$.aKey+`, []Result{newResult(`{}`, JSONObject, `aKey`)}},
	{`object w/ height=2`, `{"aKey":{"bKey":32}}`, `$.aKey.bKey+`, []Result{newResult(`32`, JSONNumber, `aKey`, `bKey`)}},
	{`array of multiple types`, `{"aKey":[1,{"s":true},"asdf"]}`, `$.aKey[1]+`, []Result{newResult(`{"s":true}`, JSONObject, `aKey`, 1)}},
	{`nested array selection`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey+`, []Result{newResult(`[123,456]`, JSONArray, `aKey`, `bKey`)}},
	{`nested array`, `[[[[[]], [true, false, []]]]]`, `$[0][0][1][2]+`, []Result{newResult(`[]`, JSONArray, 0, 0, 1, 2)}},
	{`index of array selection`, `{"aKey":{"bKey":[123, 456, 789]}}`, `$.aKey.bKey[1]+`, []Result{newResult(`456`, JSONNumber, `aKey`, `bKey`, 1)}},
	{`index of array selection (more than one)`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey[1]+`, []Result{newResult(`456`, JSONNumber, `aKey`, `bKey`, 1)}},
	{`multi-level object/array`, `{"1Key":{"aKey": null, "bKey":{"trash":[1,2]}, "cKey":[123,456] }, "2Key":false}`, `$.1Key.bKey.trash[0]+`, []Result{newResult(`1`, JSONNumber, `1Key`, `bKey`, `trash`, 0)}},
	{`multi-level array`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*].michael[1]+`, []Result{newResult(`6`, JSONNumber, `aKey`, 3, `michael`, 1)}},
	{`multi-level array 2`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*][1]+`, []Result{newResult(`"3"`, JSONString, `aKey`, 4, 1)}},

	{`evaluation literal equality`, `{"items":[ {"name":"alpha", "value":11}]}`, `$.items[*]?("bravo" == "bravo").value+`, []Result{newResult(`11`, JSONNumber, `items`, 0, `value`)}},
	{`evaluation based on string equal to path value`, `{"items":[ {"name":"alpha", "value":11}, {"name":"bravo", "value":22}, {"name":"charlie", "value":33} ]}`, `$.items[*]?(@.name == "bravo").value+`, []Result{newResult(`22`, JSONNumber, `items`, 1, `value`)}},
}

func TestPathQuery(t *testing.T) {
	as := assert.New(t)

	for _, t := range tests {
		paths, err := ParsePaths(t.path)
		if as.NoError(err) {
			eval, err := EvalPathsInBytes([]byte(t.json), paths)
			if as.NoError(err, "Testing: %s", t.name) {
				res := toResultArray(eval)

				if as.NoError(eval.Error) {
					as.EqualValues(t.expected, res, "Testing of %q", t.name)
				}
			}

			evalReader, err := EvalPathsInReader(strings.NewReader(t.json), paths)
			if as.NoError(err, "Testing: %s", t.name) {
				res := toResultArray(evalReader)

				if as.NoError(eval.Error) {
					as.EqualValues(t.expected, res, "Testing of %q", t.name)
				}
			}
		}
	}
}

func newResult(value string, typ int, keys ...interface{}) Result {
	keysChanged := make([]interface{}, len(keys))
	for i, k := range keys {
		switch v := k.(type) {
		case string:
			keysChanged[i] = []byte(v)
		default:
			keysChanged[i] = v
		}
	}

	return Result{
		Value: []byte(value),
		Keys:  keysChanged,
		Type:  typ,
	}
}

func toResultArray(e *Eval) []Result {
	vals := make([]Result, 0)
	for {
		if r, ok := e.Next(); ok {
			if r != nil {
				vals = append(vals, *r)
			}
		} else {
			break
		}
	}
	return vals
}
