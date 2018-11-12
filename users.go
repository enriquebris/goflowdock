package goflowdock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserManager struct {
	token string
}

type Users []struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Nick    string `json:"nick"`
	Avatar  string `json:"avatar"`
	Website string `json:"website"`
}

func NewUserManager(token string) *UserManager {
	ret := &UserManager{}
	ret.initialize(token)

	return ret
}

func (st *UserManager) initialize(token string) {
	st.token = token
}

// GetAll read and returns all users (based on the API token).
func (st *UserManager) GetAll() (Users, error) {
	request, err := http.NewRequest("GET", URLUsers, nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Basic %v", st.token))

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	s := buf.String()

	var users Users
	err = json.Unmarshal([]byte(s), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}
