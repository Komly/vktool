package vk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type VKErrorResponse struct {
	Error *VKError `json:"error"`
}

type VKError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	// TODO: request_params
}

type VkLongPollUpdate []interface{}

type VkLongPollResp struct {
	Ts      int                `json:"ts"`
	Updates []VkLongPollUpdate `json:"updates"`
}

type VkLongPollAddNewMessage struct {
	MessageID int
	Flags     int
	PeerId    int
	Timestamp int
	Subject   string
	Text      string
}

func (err VKError) Error() string {
	return err.ErrorMsg
}

type VKResponse struct {
	Response *map[string]interface{} `json:"response"`
}

func ApiCall(method string, params map[string]string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.vk.com/method/%s", method)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("%s", body)

	if err != nil {
		return nil, err
	}

	vkError := VKErrorResponse{}
	if err := json.Unmarshal(body, &vkError); err == nil {
		if vkError.Error != nil {
			return nil, vkError.Error
		}
	}

	vkResp := VKResponse{}
	if err := json.Unmarshal(body, &vkResp); err == nil {
		if vkResp.Response != nil {
			return *vkResp.Response, nil
		}
	}

	return nil, errors.New("Invalid api response")

}

func MakeLongPollRequest(server, key string, ts int) ([]interface{}, int, error) {
	lastTs := ts
	url := fmt.Sprintf("https://%s?act=a_check&key=%s&ts=%d&wait=3&mode=2&version=1", server, key, ts)
	lpResp, err := http.Get(url)

	body, err := ioutil.ReadAll(lpResp.Body)
	if err != nil {
		return nil, 0, err
	}
	defer lpResp.Body.Close()

	resp := VkLongPollResp{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, 0, err
	}
	lastTs = resp.Ts
	res := make([]interface{}, 0)
	for _, update := range resp.Updates {
		code, ok := update[0].(float64)
		if !ok {
			return nil, 0, errors.New("Invalid event code")
		}
		intCode := int(code)

		switch intCode {
		case 4:
			msg := VkLongPollAddNewMessage{
				MessageID: int(update[0].(float64)),
				Flags:     1,
			}
			res = append(res, msg)
		default:
			log.Printf("Unsupported update code: %d", intCode)
		}

	}

	return res, lastTs, nil
}
