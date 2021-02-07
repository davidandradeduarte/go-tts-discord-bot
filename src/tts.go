package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

const ttsURL = "https://ttsmp3.com/makemp3_new.php"
const lang = "Ines"
const source = "ttsmp3"

type Response struct {
	Error    string    `json:"Error"`
	Speaker  string `json:"Speaker"`
	Cached   int    `json:"Cached"`
	Text     string `json:"Text"`
	tasktype string `json:"tasktype"`
	URL      string `json:"URL"`
	MP3      string `json:"MP3"`
}

// SynthesizeText synthesizes plain text and saves the output to outputFile.
func SynthesizeText(text string) (string, error) {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return "", err
	}

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		// Note: the voice can also be specified by name.
		// Names of voices can be retrieved with client.ListVoices().
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "pt-PT",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile("output.mp3", resp.AudioContent, 0644)
	if err != nil {
		return "", err
	}
	//fmt.Fprintf(w, "Audio content written to file: %v\n", outputFile)
	return "output.mp3", nil
}

func getVoiceFromText(message string) (Response, error) {
	formData := url.Values{"msg": {message}, "lang": {lang}, "source": {source}}

	log.Println("Getting audio from text:", message)
	
	resp, err := http.PostForm(ttsURL, formData)

	if err != nil {
		return Response{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(body))
	
	data := Response{}
	err = json.Unmarshal(body, &data)

	if err != nil {
		return Response{}, err
	}

	return data, nil
}
