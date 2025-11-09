package harrislrc

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

type lrcMessageOp string

const (
	_CHANGE       lrcMessageOp = ":"
	_CHANGENOTIFY lrcMessageOp = "!"
	_QUERY        lrcMessageOp = "?"
	_QUERYRESP    lrcMessageOp = "%"
)

type lrcArgType string

const (
	_STRING  lrcArgType = "$"
	_NUMERIC lrcArgType = "#"
	_UTF     lrcArgType = "&"
)

type lrcMessageArg struct {
	name    string
	argType lrcArgType
	values  []string
}

func lrcMessageArgFromString(str string) lrcMessageArg {
	arg := lrcMessageArg{}
	typeIdx := -1
	for i, c := range str {
		switch string(c) {
		case string(_STRING):
			typeIdx = i
			arg.argType = _STRING
		case string(_NUMERIC):
			typeIdx = i
			arg.argType = _NUMERIC
		case string(_UTF):
			typeIdx = i
			arg.argType = _UTF
		}
		if typeIdx >= 0 {
			break
		}
	}
	if typeIdx < 0 {
		log.Error("Harris LRC Router: Error parsing argument", str)
		return arg
	}
	arg.name = str[:typeIdx]
	arg.values = strings.Split(str[typeIdx+2:len(str)-1], ",")
	return arg
}

type lrcMessage struct {
	msgType string
	op      lrcMessageOp
	args    map[string]lrcMessageArg
}

func lrcMessageFromString(str string) lrcMessage {
	str = strings.Trim(str, " \r\n")
	msg := lrcMessage{}
	opIdx := -1
	for i, c := range str {
		switch string(c) {
		case string(_CHANGE):
			opIdx = i
			msg.op = _CHANGE
		case string(_CHANGENOTIFY):
			opIdx = i
			msg.op = _CHANGENOTIFY
		case string(_QUERY):
			opIdx = i
			msg.op = _QUERY
		case string(_QUERYRESP):
			opIdx = i
			msg.op = _QUERYRESP
		}
		if opIdx >= 0 {
			break
		}
	}
	if str[0] != '~' || str[len(str)-1] != '\\' || opIdx < 0 {
		log.Error("Harris LRC Router: Error parsing", str)
		return msg
	}
	msg.msgType = str[1:opIdx]
	msg.args = make(map[string]lrcMessageArg)
	argsStr := str[opIdx+1 : len(str)-1]
	argStrs := strings.Split(argsStr, ";")
	for _, argStr := range argStrs {
		arg := lrcMessageArgFromString(argStr)
		msg.args[arg.name] = arg
	}
	return msg
}
