package main

import (
	"regexp"
	"strconv"

	"noerkrieg.com/mrgrep/internal"
)

// The reduce function is called once for each key generated by the
// map tasks, with a list of all the values created for that key by
// any map task.
func Reduce(key string, values []string) string {
	// return the number of occurrences of this word.
	return strconv.Itoa(len(values))
}

// MapRegex function consumes a regex pattern and contents, producing a slice of rpcargs.KeyValue
func Map(match string, contents string) []internal.KeyValue {
	re := regexp.MustCompile(match)
	matches := re.FindAllString(contents, -1)
	kva := []internal.KeyValue{}
	for _, m := range matches {
		kv := internal.KeyValue{Key: m, Value: "1"}
		kva = append(kva, kv)
	}
	return kva
}
