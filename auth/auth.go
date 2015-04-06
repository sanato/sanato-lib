package auth

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
)

// User reprents a user saved in the auth file
type User struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// AuthResource represents the user after a valid authentication
// EntityId is the primary identifier
// EntityType is the type of user to authorize, like user, group, robot ...
// DisplayName is the user friendly name for the user id
// Email is the email of the user
// AuthType is the type of authentication used for this user
// Extra is a flexible field for auth providers to provide extra information
type AuthResource struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

func newAuthResource(username, displayName, email string) *AuthResource {
	// do some validations
	// authType, username, userUser, email valid chars PORTABLE v3
	authRes := AuthResource{username, displayName, email}
	return &authRes
}

// AuthProvider is a AuthProvider that uses a MySQL database
// for authentication purposes.
type AuthProvider struct {
	authFile string
}

// NewAuthProvider returns a AuthPovider or an error in case of failure
func NewAuthProvider(authFile string) (*AuthProvider, error) {
	return &AuthProvider{authFile}, nil
}

// Authenticate authenticate the request agains the file and returns
// an AuthResource or a validation error or a serve error
func (a *AuthProvider) Authenticate(username, password string) (*AuthResource, error) {
	fd, err := os.Open(a.authFile)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	data, err := ioutil.ReadAll(fd)

	users := make([]*User, 0)
	err = json.Unmarshal(data, &users)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, user := range users {
		if user.Username == username && bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
			authRes := newAuthResource(user.Username, user.DisplayName, user.Email)
			return authRes, nil
		}
	}
	return nil, errors.New("user not found")
}
func (ap *AuthProvider) ExistsAuth() error {
	_, err := os.Open(ap.authFile)
	if err != nil {
		return err
	}
	return nil
}

func (ap *AuthProvider) CreateUser(user *User) error {
	users := []*User{user}
	fd, err := os.Create(ap.authFile)
	if err != nil {
		return err
	}
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}
	_, err = fd.Write(data)
	if err != nil {
		return err
	}
	return nil
}
