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

// Ping checks that the database is running
func (u *UserDB) Ping() error {
	err := u.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		logrus.Errorf("ERROR connecting to database %v", err)
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
		"roles":     result.Roles,
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
func (u *UserDB) CreateRole(role *models.Role) error {
	logrus.Debug("BEGIN - CreateRole")

	collection := u.client.Database(u.databaseName).Collection(u.roleCollection)

	_, err := collection.InsertOne(context.Background(), role)

	return err
}

// DeleteRole removes a role from the role collection
func (u *UserDB) DeleteRole(role *models.Role) error {
	logrus.Debug("BEGIN - DeleteRole")

	collection := u.client.Database(u.databaseName).Collection(u.roleCollection)

	_, err := collection.DeleteOne(context.Background(), role)

	return err
}

// GetRoles returns all existing roles in the role collection
func (u *UserDB) GetRoles(queryParams url.Values) ([]models.Role, error) {
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

	var roles []models.Role
	for cur.Next(context.Background()) {
		var role models.Role
		err := cur.Decode(&role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// AddUserRole adds a role that exists in the role collection to the specified user
func (u *UserDB) AddUserRole(user models.User, role *models.Role) error {
	logrus.Debug("Begin - AdddUserRole")

	collection := u.client.Database(u.databaseName).Collection(u.userCollection)

	opts := options.FindOne().
		SetMaxTime(30 * time.Second).
		SetSkip(int64(0)).
		SetSort(bson.D{{
			Key:   "username",
			Value: 1,
		}})

	result := collection.FindOne(context.Background(), bson.M{"username": user.Username}, opts)
	if result.Err() != nil {
		return result.Err()
	}

	var found models.User

	result.Decode(&found)
	found.Roles = append(found.Roles, *role)

	_, err := collection.UpdateOne(context.Background(), bson.M{"username": found.Username}, bson.D{{
		Key:   "$set",
		Value: found,
	}})

	return err
}

// RemoveUserRole removes a role assigned to a user
func (u *UserDB) RemoveUserRole(user models.User, role *models.Role) error {
	logrus.Debug("Begin - RemoveUserRole")

	collection := u.client.Database(u.databaseName).Collection(u.userCollection)

	opts := options.FindOne().
		SetMaxTime(30 * time.Second).
		SetSkip(int64(0)).
		SetSort(bson.D{{
			Key:   "username",
			Value: 1,
		}})

	result := collection.FindOne(context.Background(), bson.M{"username": user.Username}, opts)
	if result.Err() != nil {
		return result.Err()
	}

	var found models.User

	result.Decode(&found)
	for i, foundRole := range found.Roles {
		if foundRole.Name == role.Name {
			found.Roles = append(found.Roles[:i], found.Roles[i+1:]...)
		}
	}

	_, err := collection.UpdateOne(context.Background(), bson.M{"username": found.Username}, bson.D{{
		Key:   "$set",
		Value: found,
	}})

	return err
}
