package main

import (
	"github.com/trustpkg/trustpkg-api/db"
	"github.com/trustpkg/trustpkg-api/internal/npm"
)

func main() {
	db.ConnectDb()

	npm.Pipeline()
}
