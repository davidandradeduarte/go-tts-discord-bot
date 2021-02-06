package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

const ttsURL = "https://ttsmp3.com/makemp3_new.php"
const lang = "Ines"
const source = "ttsmp3"

type Response struct {
	Error    int    `json:"Error"`
	Speaker  string `json:"Speaker"`
	Cached   int    `json:"Cached"`
	Text     string `json:"Text"`
	tasktype string `json:"tasktype"`
	URL      string `json:"URL"`
	MP3      string `json:"MP3"`
}

func getVoiceFromText(message string) (Response, error) {
	formData := url.Values{"msg": {message}, "lang": {lang}, "source": {source}}

	resp, err := http.PostForm(ttsURL, formData)

	if err != nil {
		return Response{}, errors.New("error posting tts message")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	data := Response{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return Response{}, errors.New("error parsing JSON response")
	}

	return data, nil
}
