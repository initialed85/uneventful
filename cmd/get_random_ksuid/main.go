package main

import (
	"log"

	"github.com/segmentio/ksuid"
)

func main() {
	randomKSUID := ksuid.New()

	log.Printf("random_ksuid_struct=%#+v", randomKSUID)
	log.Printf("random_ksuid_string=%#+v", randomKSUID.String())
}
