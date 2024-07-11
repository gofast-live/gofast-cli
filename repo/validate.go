package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Response struct {
	Token string `json:"token"`
}

func (m *model) valdiate() (string, error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Could not get user config dir")
	}
	config := path + "/gofast.json"
	jsonFile, err := os.OpenFile(config, os.O_RDWR, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not open config file")
	}
	defer jsonFile.Close()
	data, err := io.ReadAll(jsonFile)
	if err != nil {
		return "", fmt.Errorf("Could not read config file")
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return "", fmt.Errorf("Could not unmarshal config data")
	}

	// make http call with query email and header authorization with api key
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.gofast.com/validate", nil)
	if err != nil {
		return "", fmt.Errorf("Could not create request")
	}
	req.Header.Add("Authorization", "Bearer "+c.ApiKey)
	q := req.URL.Query()
	q.Add("email", c.Email)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Could not make request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Invalid credentials")
	}
	var r Response
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "", fmt.Errorf("Could not decode response")
	}
	return r.Token, nil
}
