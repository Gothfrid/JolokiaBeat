package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/gothfrid/jolokiabeat/beater"
)

func main() {
	err := beat.Run("jolokiabeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
