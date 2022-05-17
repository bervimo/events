package main

import (
	"fmt"
	"os"

	"github.com/bervimo/events/internal/adapters"
	"github.com/bervimo/events/internal/core/services"
	"github.com/bervimo/events/internal/repositories"
	"github.com/bervimo/events/server"
	"github.com/bervimo/events/utils"
	"github.com/rs/zerolog/log"

	secret_manager "github.com/NBN23dev/go-secret-manager"
)

func main() {
	// Secrets
	sm, _ := secret_manager.NewSecretManager()
	sm.AccessSecrets()

	// Env vars
	srvName, _ := utils.GetEnvOr("SERVICE_NAME", "unknown")
	port, _ := utils.GetEnvOr("PORT", 8080)
	dbURI, _ := utils.GetEnv[string]("DATABASE_URI")
	dbName, _ := utils.GetEnv[string]("DATABASE_NAME")
	dbCollection, _ := utils.GetEnv[string]("DATABASE_COLLECTION_EVENTS")

	// Application
	opts := repositories.MongoOptions{URI: dbURI, Database: dbName, Collection: dbCollection}
	repository, err := repositories.NewMongo(opts)

	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	service := services.NewService(repository)
	adapter := adapters.NewGRPCAdapter(service)

	// Create server
	srv, hs := server.NewServer(adapter)

	// Shutdown
	go server.GracefulShutdown(hs, func(sig os.Signal) {
		srv.GracefulStop()

		repository.Close()
		sm.Close()

		log.Info().Msg(fmt.Sprintf("'%s' service it is about to end", srvName))
	})

	log.Info().Msg(fmt.Sprintf("'%s' service it is about to start", srvName))

	// Start server
	server.StartServer(srv, port)
}
