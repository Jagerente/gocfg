package parsers

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	defaultTypeParsers = map[reflect.Type]func(v string) (interface{}, error){
		reflect.TypeOf(time.Duration(83)): func(v string) (interface{}, error) {
			return time.ParseDuration(v)
		},
		reflect.TypeOf([]byte{}): func(v string) (interface{}, error) {
			return []byte(v), nil
		},
		reflect.TypeOf([]string{}): func(v string) (interface{}, error) {
			return strings.Split(v, ","), nil
		},
		reflect.TypeOf([]int{}): func(v string) (interface{}, error) {
			parts := strings.Split(v, ",")
			result := make([]int, len(parts))
			for i, p := range parts {
				n, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil {
					return nil, err
				}
				result[i] = n
			}
			return result, nil
		},
	}

	defaultKindParsers = map[reflect.Kind]func(v string) (interface{}, error){
		reflect.Bool: func(v string) (interface{}, error) {
			return strconv.ParseBool(v)
		},
		reflect.String: func(v string) (interface{}, error) {
			return v, nil
		},
		reflect.Int: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int(i), err
		},
		reflect.Int16: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 16)
			return int16(i), err
		},
		reflect.Int32: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int32(i), err
		},
		reflect.Int64: func(v string) (interface{}, error) {
			return strconv.ParseInt(v, 10, 64)
		},
		reflect.Int8: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 8)
			return int8(i), err
		},
		reflect.Uint: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint(i), err
		},
		reflect.Uint16: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 16)
			return uint16(i), err
		},
		reflect.Uint32: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint32(i), err
		},
		reflect.Uint64: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 64)
			return i, err
		},
		reflect.Uint8: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 8)
			return uint8(i), err
		},
		reflect.Float64: func(v string) (interface{}, error) {
			return strconv.ParseFloat(v, 64)
		},
		reflect.Float32: func(v string) (interface{}, error) {
			f, err := strconv.ParseFloat(v, 32)
			return float32(f), err
		},
	}
)

type DefaultParserProvider struct {
}

func NewDefaultParserProvider() *DefaultParserProvider {
	return &DefaultParserProvider{}
}

func (p *DefaultParserProvider) Get(value reflect.Value) (parser func(v string) (interface{}, error), ok bool) {
	if parser, ok = defaultTypeParsers[value.Type()]; !ok {
		if parser, ok = defaultKindParsers[value.Kind()]; !ok {
			return
		}
	}
	return
}
