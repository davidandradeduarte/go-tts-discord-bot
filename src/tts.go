package main

import (
	"context"
	"io/ioutil"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

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

	return "output.mp3", nil
}
