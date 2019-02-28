package main

import "os"

import "application"

func main() {
	app := application.App{}
	app.Initialize(
		os.Getenv("COUCHBASE_USER"),
		os.Getenv("COUCHBASE_PASS"),
		os.Getenv("COUCHBASE_DB"))
	app.Run(os.Getenv("PORT"))
}
