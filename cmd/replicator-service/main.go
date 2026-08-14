package main

import (
	"fmt"

	"github.com/trustpkg/trustpkg-api/db"
	"github.com/trustpkg/trustpkg-api/internal/npm"
)

func main() {
	db.ConnectDb()

	err := npm.Pipeline()
	if err != nil {
		fmt.Println("commit error: ", err)
	}
}
