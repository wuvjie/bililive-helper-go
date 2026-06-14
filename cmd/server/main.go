// Package main 是 Bililive Helper 应用程序的入口。
package main

import (
	"log"

	"bililive-helper/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
