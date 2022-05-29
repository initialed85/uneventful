package main

import (
	"github.com/segmentio/ksuid"
	"log"
)

func main() {
	randomKSUID := ksuid.New()

	log.Printf("random_ksuid_struct=%#+v", randomKSUID)
	log.Printf("random_ksuid_string=%#+v", randomKSUID.String())
}
