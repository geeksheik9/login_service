package db

import (
	"context"
	"os"

	"github.com/geeksheik9/login-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitializeClients returns a mongo client.
func InitializeClients(ctx context.Context) (*mongo.Client, error) {

	options := options.Client().ApplyURI(os.Getenv("MONGO_URI"))

	err := options.Validate()
	if err != nil {
		return nil, err
	}

	client, err := mongo.Connect(ctx, options)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, err
}

// InitializeDatabases Factory for the dao implementation. Returns a dao connected to the designated MongoDB database for DB operations.
// The database connection is made using configuration in the config.go file
func InitializeDatabases(client *mongo.Client, config *config.Config) *UserDB {

	database := &UserDB{
		client:         client,
		databaseName:   config.UserDatabase,
		userCollection: config.UserCollection,
		roleCollection: config.RoleCollection,
	}

	return database
}
