package proxmox

import (
	"encoding/json"
	"unicode/utf8"
	"errors"
	"fmt"
	"io"
	"log"
)

// User options for the Proxmox API
type ConfigUser struct {
	UserID    string   `json:"userid"`
	Comment   string   `json:"comment,omitempty"`
	Email     string   `json:"email,omitempty"`
	Enable    bool     `json:"enable"`
	Expire    int      `json:"expire"`
	Firstname string   `json:"firstname,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Keys      string   `json:"keys,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
}

func (config ConfigUser) MapUserValues()(params map[string]interface{}) {
	params = map[string]interface{}{
		"comment":   config.Comment,
		"email":     config.Email,
		"enable":    config.Enable,
		"expire":    config.Expire,
		"firstname": config.Firstname,
		"groups":    ArrayToCSV(config.Groups),
		"keys":      config.Keys,
		"lastname":  config.Lastname,
	}
	return
}

func (config ConfigUser) SetUser(userid string, password string, client *Client) (err error) {
	err = ValidateUserPassword(password)
	if err != nil {
		return err
	}

	config.UserID = userid

	userExists, err := client.CheckUserExistance(userid)

	if userExists == true {
		err = config.UpdateUser(client)
		if err != nil {
			return err
		}
		if password != "" {
			err = client.UpdateUserPassword(userid, password)
		}
	} else {
		err = config.CreateUser(password, client)
	}
	return
}

func (config ConfigUser) CreateUser(password string, client *Client) (err error) {
	params := config.MapUserValues()
	params["userid"] = config.UserID
	params["password"] = password
	exitStatus, err := client.CreateUser(params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating User: %v, error status: %s (params: %v)", err, exitStatus, string(params))
	}
	return
}

func (config ConfigUser) UpdateUser(client *Client) (err error) {
	params := config.MapUserValues()
	exitStatus, err := client.UpdateUser(config.UserID, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating User: %v, error status: %s (params: %v)", err, exitStatus, string(params))
	}
	return
}

func NewConfigUserFromApi(userid string, client *Client) (config *ConfigUser, err error) {
	// prepare json map to receive the information from the api
	var userConfig map[string]interface{}
	userConfig, err = client.GetUserConfig(userid)
	if err != nil {
		return nil, err
	}
	config = new(ConfigUser)

	config.UserID = userid

	if _, isSet := userConfig["comment"]; isSet {config.Comment = userConfig["comment"].(string)}
	if _, isSet := userConfig["email"]; isSet {config.Email = userConfig["email"].(string)}
	if _, isSet := userConfig["enable"]; isSet {config.Enable = Itob(int(userConfig["enable"].(float64)))}
	if _, isSet := userConfig["expire"]; isSet {config.Expire = int(userConfig["expire"].(float64))}
	if _, isSet := userConfig["firstname"]; isSet {config.Firstname = userConfig["firstname"].(string)}
	if _, isSet := userConfig["keys"]; isSet {config.Keys = userConfig["keys"].(string)}
	if _, isSet := userConfig["lastname"]; isSet {config.Lastname = userConfig["lastname"].(string)}
	if _, isSet := userConfig["groups"]; isSet {config.Groups = ArrayToStringType(userConfig["groups"].(interface{}).([]interface{}))}

	return
}

func NewConfigUserFromJson(io io.Reader) (config *ConfigUser, err error) {
	config = &ConfigUser{}
	err = json.NewDecoder(io).Decode(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return
}

func ValidateUserPassword(password string) error{
	if utf8.RuneCountInString(password) >= 5 || password == ""{
		return nil
	}
	return errors.New("error updating User: the minimum password length is 5")
}