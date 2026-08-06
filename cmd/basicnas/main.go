package main

import (
	"BasicNAS/internal/smb"
	"log"
)

func main() {
	var err error
	var srv *smb.Server = smb.NewServer(":4450")

	if err = srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
}
