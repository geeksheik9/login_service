package db

import (
	"context"
	"errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/geeksheik9/login-service/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

// UserDB is the data access object for user login
type UserDB struct {
	client         *mongo.Client
	databaseName   string
	collectionName string
}

//Ping checks that the database is running
func (u *UserDB) Ping() error {
	err := u.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		logrus.Errorf("ERROR connectiong to database %v", err)
	}
	return err
}

// RegisterUser creates and inserts a user into the database
func (u *UserDB) RegisterUser(user *models.User) error {
	collection := u.client.Database(u.databaseName).Collection(u.collectionName)

	var result models.User
	err := collection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&result)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)
			if err != nil {
				return err
			}
			user.Password = string(hash)

			_, err = collection.InsertOne(context.TODO(), user)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	err = errors.New("Username already Exists")
	return err
}

// LoginUser is the implementation to login a user in the database
func (u *UserDB) LoginUser(user *models.User) (string, error) {
	collection := u.client.Database(u.databaseName).Collection(u.collectionName)

	var result models.User
	err := collection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&result)
	if err != nil {
		return result.Token, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		return result.Token, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":  result.Username,
		"firstname": result.FirstName,
		"lastname":  result.LastName,
	})

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		return result.Token, err
	}

	result.Token = tokenString
	result.Password = ""

	return result.Token, nil
}
