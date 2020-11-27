package template

import (
	"strings"
	gotemplate "text/template"
	"time"
)

func GetDefaultFuncs() gotemplate.FuncMap {
	return map[string]interface{}{
		"Now":         Now,
		"NowUnix":     NowUnix,
		"NowUnixNano": NowUnixNano,
		"ToLower":     ToLower,
		"Replace":     Replace,
		"Split":       Split,
	}
}

func Now(vv ...interface{}) (interface{}, error) {
	format := time.RFC3339
	if len(vv) > 0 {
		fmt, isString := vv[0].(string)
		if isString {
			format = fmt
		}
	}
	return time.Now().Format(format), nil
}

func NowUnix(vv ...interface{}) (interface{}, error) {
	return time.Now().Unix(), nil
}

func NowUnixNano(vv ...interface{}) (interface{}, error) {
	return time.Now().UnixNano(), nil
}

func ToLower(vv ...interface{}) (interface{}, error) {
	if len(vv) > 0 {
		value, isString := vv[0].(string)
		if !isString {
			return "", nil
		}
		return strings.ToLower(value), nil
	}
	return "", nil
}

func Replace(vv ...interface{}) (interface{}, error) {
	if len(vv) == 4 {
		s, isString := vv[0].(string)
		if !isString {
			return "", nil
		}
		old, isString := vv[1].(string)
		if !isString {
			return "", nil
		}
		new, isString := vv[2].(string)
		if !isString {
			return "", nil
		}
		n, isInt := vv[3].(int)
		if !isInt {
			return "", nil
		}
		return strings.Replace(s, old, new, n), nil
	}
	return "", nil
}

func Split(vv ...interface{}) (interface{}, error) {
	if len(vv) == 2 {
		s, isString := vv[0].(string)
		if !isString {
			return []string{}, nil
		}
		sep, isString := vv[1].(string)
		if !isString {
			return []string{}, nil
		}
		return strings.Split(s, sep), nil
	}
	return []string{}, nil
}
