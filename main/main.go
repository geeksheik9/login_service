// Package main Login Service API
//
// API for registering, logginging in, and getting user information
//
// version: 0.0.1-alpha
//
// swagger:meta
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/geeksheik9/login-service/config"
	"github.com/geeksheik9/login-service/pkg/db"
	"github.com/geeksheik9/login-service/pkg/handler"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var version string

func main() {
	//go:generate swagger generate spec
	logrus.Info("INITIALIZING LOGIN SERVICE")

	accessor := viper.New()

	secretsPath := os.Getenv("SECRETS_PATH")

	secret, err := config.GetSecret(secretsPath)
	if err != nil {
		log.Fatalf("error loading config: %s", err.Error())
	}

	config, err := config.New(accessor)
	if err != nil {
		logrus.Fatalf("ERROR LOADING CONFIG: %v", err.Error())
	}

	timeout := time.Second * 5
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := db.InitializeClients(ctx, *secret)
	if err != nil {
		logrus.Warnf("Failed to intialize client with error: %v, trying again", err)
		err = nil
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*60)
		client, err = db.InitializeClients(ctx, *secret)
		if err != nil {
			logrus.Fatalf("Failed to initialize database client a second time with error: %v", err)
		}
	}

	defer client.Disconnect(context.Background())

	database := db.InitializeDatabases(client, config)
	if database == nil {
		logrus.Fatalf("Error no database from client %v", client)
	}

	gearService := handler.LoginService{
		Version:  version,
		Database: database,
	}

	r := mux.NewRouter().StrictSlash(true)

	r = gearService.Routes(r)
	fmt.Printf("Server listen on port %v\n", config.Port)
	logrus.Info("END")
	logrus.Fatal(http.ListenAndServe(":"+config.Port, cors.AllowAll().Handler(r)))
}
