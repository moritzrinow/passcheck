package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	db, err := leveldb.OpenFile("./data/db", nil)
	if err != nil {
		fmt.Println(err)
	}

	defer db.Close()

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		fmt.Println(key, val)
	}
}
