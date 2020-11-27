package main

import (
	"strings"
)

func main() {
}

func Username(v ...interface{}) (interface{}, error) {
	firstname, lastname := v[0].(string), v[1].(string)
	firstnames := strings.Split(firstname, " ")
	var username string
	for _, f := range firstnames {
		username += string(f[0])
	}
	username += lastname
	return username, nil
}
