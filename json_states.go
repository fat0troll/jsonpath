package jsonpath

const (
	jsonError = iota
	jsonEOF

	jsonBraceLeft
	jsonBraceRight
	jsonBracketLeft
	jsonBracketRight
	jsonColon
	jsonComma
	jsonNumber
	jsonString
	jsonNull
	jsonKey
	jsonBool
)

var trueBytes = []byte{'t', 'r', 'u', 'e'}
var falseBytes = []byte{'f', 'a', 'l', 's', 'e'}
var nullBytes = []byte{'n', 'u', 'l', 'l'}

var jsonTokenNames = map[int]string{
	jsonEOF:   "EOF",
	jsonError: "ERROR",

	jsonBraceLeft:    "{",
	jsonBraceRight:   "}",
	jsonBracketLeft:  "[",
	jsonBracketRight: "]",
	jsonColon:        ":",
	jsonComma:        ",",
	jsonNumber:       "NUMBER",
	jsonString:       "STRING",
	jsonNull:         "NULL",
	jsonKey:          "KEY",
	jsonBool:         "BOOL",
}

var JSON = lexJsonRoot

func lexJsonRoot(l lexer, state *intStack) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()
	var next stateFn
	switch cur {
	case '{':
		next = stateJSONObjectOpen
	case '[':
		next = stateJSONArrayOpen
	default:
		next = l.errorf("Expected '{' or '[' at root of JSON instead of %#U", cur)
	}
	return next
}

func stateJSONObjectOpen(l lexer, state *intStack) stateFn {
	cur := l.take()
	if cur != '{' {
		return l.errorf("Expected '{' as start of object instead of %#U", cur)
	}
	l.emit(jsonBraceLeft)
	state.push(jsonBraceLeft)

	return stateJSONObject
}

func stateJSONArrayOpen(l lexer, state *intStack) stateFn {
	cur := l.take()
	if cur != '[' {
		return l.errorf("Expected '[' as start of array instead of %#U", cur)
	}
	l.emit(jsonBracketLeft)
	state.push(jsonBracketLeft)

	return stateJSONArray
}

func stateJSONObject(l lexer, state *intStack) stateFn {
	var next stateFn
	cur := l.peek()
	switch cur {
	case '}':
		if top, ok := state.peek(); ok && top != jsonBraceLeft {
			next = l.errorf("Received %#U but has no matching '{'", cur)
			break
		}
		l.take()
		l.emit(jsonBraceRight)
		state.pop()
		next = stateJSONAfterValue
	case '"':
		next = stateJSONKey
	default:
		next = l.errorf("Expected '}' or \" within an object instead of %#U", cur)
	}
	return next
}

func stateJSONArray(l lexer, state *intStack) stateFn {
	var next stateFn
	cur := l.peek()
	switch cur {
	case ']':
		if top, ok := state.peek(); ok && top != jsonBracketLeft {
			next = l.errorf("Received %#U but has no matching '['", cur)
			break
		}
		l.take()
		l.emit(jsonBracketRight)
		state.pop()
		next = stateJSONAfterValue
	default:
		next = stateJSONValue
	}
	return next
}

func stateJSONAfterValue(l lexer, state *intStack) stateFn {
	cur := l.take()
	top, ok := state.peek()
	topVal := noValue
	if ok {
		topVal = top
	}

	switch cur {
	case ',':
		l.emit(jsonComma)
		switch topVal {
		case jsonBraceLeft:
			return stateJSONKey
		case jsonBracketLeft:
			return stateJSONValue
		case noValue:
			return l.errorf("Found %#U outside of array or object", cur)
		default:
			return l.errorf("Unexpected character in lexer stack: %#U", cur)
		}
	case '}':
		l.emit(jsonBraceRight)
		state.pop()
		switch topVal {
		case jsonBraceLeft:
			return stateJSONAfterValue
		case jsonBracketLeft:
			return l.errorf("Unexpected %#U in array", cur)
		case noValue:
			return stateJSONAfterRoot
		}
	case ']':
		l.emit(jsonBracketRight)
		state.pop()
		switch topVal {
		case jsonBraceLeft:
			return l.errorf("Unexpected %#U in object", cur)
		case jsonBracketLeft:
			return stateJSONAfterValue
		case noValue:
			return stateJSONAfterRoot
		}
	case eof:
		if state.len() == 0 {
			l.emit(jsonEOF)
			return nil
		} else {
			return l.errorf("Unexpected EOF instead of value")
		}
	default:
		return l.errorf("Unexpected character after json value token: %#U", cur)
	}
	return nil
}

func stateJSONKey(l lexer, state *intStack) stateFn {
	if err := l.takeString(); err != nil {
		return l.errorf(err.Error())
	}
	l.emit(jsonKey)

	return stateJSONColon
}

func stateJSONColon(l lexer, state *intStack) stateFn {
	cur := l.take()
	if cur != ':' {
		return l.errorf("Expected ':' after key string instead of %#U", cur)
	}
	l.emit(jsonColon)

	return stateJSONValue
}

func stateJSONValue(l lexer, state *intStack) stateFn {
	cur := l.peek()

	switch cur {
	case eof:
		return l.errorf("Unexpected EOF instead of value")
	case '"':
		return stateJSONString
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return stateJSONNumber
	case 't', 'f':
		return stateJSONBool
	case 'n':
		return stateJSONNull
	case '{':
		return stateJSONObjectOpen
	case '[':
		return stateJSONArrayOpen
	default:
		return l.errorf("Unexpected character as start of value: %#U", cur)
	}
}

func stateJSONString(l lexer, state *intStack) stateFn {
	if err := l.takeString(); err != nil {
		return l.errorf(err.Error())
	}
	l.emit(jsonString)
	return stateJSONAfterValue
}

func stateJSONNumber(l lexer, state *intStack) stateFn {
	if err := takeJSONNumeric(l); err != nil {
		return l.errorf(err.Error())
	}
	l.emit(jsonNumber)
	return stateJSONAfterValue
}

func stateJSONBool(l lexer, state *intStack) stateFn {
	cur := l.peek()
	var match []byte
	switch cur {
	case 't':
		match = trueBytes
	case 'f':
		match = falseBytes
	}

	if !takeExactSequence(l, match) {
		return l.errorf("Could not parse value as 'true' or 'false'")
	}
	l.emit(jsonBool)
	return stateJSONAfterValue
}

func stateJSONNull(l lexer, state *intStack) stateFn {
	if !takeExactSequence(l, nullBytes) {
		return l.errorf("Could not parse value as 'null'")
	}
	l.emit(jsonNull)
	return stateJSONAfterValue
}

func stateJSONAfterRoot(l lexer, state *intStack) stateFn {
	cur := l.take()
	if cur != eof {
		return l.errorf("Expected EOF instead of %#U", cur)
	}
	l.emit(jsonEOF)
	return nil
}
