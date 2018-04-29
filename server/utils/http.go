package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpReq(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	HandleError(err)
	client := http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Do(req)
	HandleError(err)

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Status Code: %d", resp.StatusCode))
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	HandleError(err)

	return content
}
