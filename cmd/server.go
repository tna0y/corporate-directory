package main

import (
	"corporate-directory/pkg/lca"
	"corporate-directory/pkg/service"
	"corporate-directory/pkg/transport"
	"log"
)

func main() {
	// Prepare solver
	solver := &lca.OnlineLCASolver{}

	// Prepare service
	svc := service.NewCorporateDirectoryService(solver)

	// Prepare server
	server := transport.SetupHttpTransport(svc)

	// Run
	log.Fatalln(server.ListenAndServe())
}
