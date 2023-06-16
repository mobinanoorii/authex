package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// -----------------------------------------------------------------------------
// utils
// -----------------------------------------------------------------------------

func Get(url string) (code int, reply string, err error) {
	client := &http.Client{
		Timeout: time.Hour * 2,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	code = resp.StatusCode
	defer resp.Body.Close()
	replyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	reply = string(replyBytes)
	return
}

// Post make a json rest request
func Post(url string, data interface{}) (code int, reply string, err error) {
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	code = resp.StatusCode
	defer resp.Body.Close()
	replyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	reply = string(replyBytes)
	return
}

func PrintResponse(code int, data string) {
	os.Stderr.Write([]byte(fmt.Sprintf("response code: %d\n", code)))
	os.Stdout.Write([]byte(data))
}
