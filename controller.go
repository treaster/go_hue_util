package go_hue_util

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

func NewAPI(hostname string, username string) *HueAPI {
	return &HueAPI{
		sync.Mutex{},
		username,
		"http://" + hostname + "/api/" + username,
		&http.Client{},
	}
}

type HueAPI struct {
	lock sync.Mutex
	username string
	apiUrl string
	client *http.Client
}

type Light struct {
	State LightState
	Type string
	Name string
	ModelId string
	UniqueId string
}

/* Dump of an example Hue state
{
  "state":{
    "on":false,
    "bri":254,
	"hue":14956,
	"sat":140,
	"effect":"none",
	"xy":[0.4571,0.4097],
	"ct":366,
	"alert":"none",
	"colormode":"ct",
	"reachable":true
}
*/

type LightState struct {
	On *bool `json:"on"`
	Bri *int `json:"bri"`
	Hue *int `json:"hue"`
	Sat *int `json:"sat"`
	Effect *string `json:"effect"`
	Alert *string `json:"alert"`
}

func Bool(b bool) *bool {
	return &b
}

func Int(i int) *int {
	return &i
}

func Str(s string) *string {
	return &s
}

func (h *HueAPI) GetLights() (map[string]Light, error) {
	resp, err := http.Get(h.apiUrl + "/lights")
	if err != nil {
		return nil, err
	}

	output := map[string]Light{}
	err = unmarshalBody(resp.Body, &output)

	return output, err
}

func (h *HueAPI) SetLightState(lightId string, state LightState) error {
	output := []struct{
		Success map[string]bool
	}{}

	url := h.apiUrl + "/lights/" + lightId + "/state"
	err := h.put(url, "application/json", state, &output)

	log.Println(output, err)
	return err
}

func (h *HueAPI) SetGroupState(groupId string, state LightState) error {
	output := []struct{
		//Success map[string]bool
	}{}

	url := h.apiUrl + "/groups/" + groupId + "/action"
	err := h.put(url, "application/json", state, &output)

	log.Println("A", output, err)
	return err
}

func (h *HueAPI) put(url string,
	contentType string,
	dataObj interface{},
	output interface{}) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	data, err := json.Marshal(dataObj)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	req, err := http.NewRequest("PUT", url, buf)
	log.Println(url, dataObj)
	req.Header.Add("Content-type", contentType)

	log.Println("PUT", url)
	resp, err := h.client.Do(req)
	if err != nil {
		log.Println(err)
	}
	return unmarshalBody(resp.Body, output)
}

func unmarshalBody(body io.Reader, obj interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, obj)
}
