package jsonpath

const (
	BadStructure         = "bad structure"
	NoMoreResults        = "no more results"
	UnexpectedToken      = "unexpected token in evaluation"
	AbruptTokenStreamEnd = "token reader is not sending anymore tokens"
)

var (
	bytesTrue  = []byte{'t', 'r', 'u', 'e'}
	bytesFalse = []byte{'f', 'a', 'l', 's', 'e'}
	bytesNull  = []byte{'n', 'u', 'l', 'l'}
)
