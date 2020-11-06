package db

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/geeksheik9/login-service/models"
	"github.com/geeksheik9/login-service/pkg/api"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

// UserDB is the data access object for user login
type UserDB struct {
	client         *mongo.Client
	databaseName   string
	userCollection string
	roleCollection string
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
	collection := u.client.Database(u.databaseName).Collection(u.userCollection)

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
	collection := u.client.Database(u.databaseName).Collection(u.userCollection)

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

// CreateRole inserts role into the role collection
func (u *UserDB) CreateRole(role string) error {
	logrus.Debug("BEGIN - CreateRole")

	collection := u.client.Database(u.databaseName).Collection(u.roleCollection)

	_, err := collection.InsertOne(context.Background(), role)

	return err
}

// DeleteRole removes a role from the role collection
func (u *UserDB) DeleteRole(role string) error {
	logrus.Debug("BEGIN - DeleteRole")

	collection := u.client.Database(u.databaseName).Collection(u.roleCollection)

	_, err := collection.DeleteOne(context.Background(), role)

	return err
}

// GetRoles returns all existing roles in the role collection
func (u *UserDB) GetRoles(queryParams url.Values) ([]string, error) {
	logrus.Debug("Begin - GetRoles")

	collection := u.client.Database(u.databaseName).Collection(u.roleCollection)

	pageNumber, pageCount, sort, filter := api.BuildFilter(queryParams)
	skip := 0
	if pageNumber > 0 {
		skip = (pageNumber - 1) * pageCount
	}

	opts := options.Find().
		SetMaxTime(30 * time.Second).
		SetSkip(int64(skip)).
		SetLimit(int64(pageCount)).
		SetSort(bson.D{{
			Key:   sort,
			Value: 1,
		}})

	cur, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}

	var roles []string
	for cur.Next(context.Background()) {
		var role string
		err := cur.Decode(&role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// AddUserRole adds a role that exists in the role collection to the specified user
func (u *UserDB) AddUserRole(user models.User, role string) error {
	return nil
}

// RemoveUserRole removes a role assigned to a user
func (u *UserDB) RemoveUserRole(user models.User, role string) error {
	return nil
}
