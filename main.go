// Package main is the entry point for the dinnerclub application.
// It initializes and starts the web server for managing restaurant entries and reviews.
package main

import cmd "github.com/toozej/dinnerclub/cmd/dinnerclub"

func main() {
	cmd.Execute()
}
