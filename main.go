package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Absolute error handling
type segmentationFault struct{}

func (segerr segmentationFault) Error() string {
	return "Segmentation Fault."
}

// This now encodes corretcly into json
type StandardMessage struct {
	Chat_ID int64  `json:"chat_id"`
	Text    string `json:"text"`
}
type TreeItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}
type StandardResponse struct {
	Trees []TreeItem `json:"trees"`
	NextPageToken string `json:"next_page_token"`
}
var (
	orgSlug  = flag.String("os", "none", "org slug")
	repoSlug = flag.String("rs", "none", "repo slug")
	PATtoken = flag.String("pt", "none", "PAT token")
	TGtoken  = flag.String("tt", "none", "Telegram token")
	chatID   = flag.Int64("id", -1, "Telegram chat ID")
)

func init() {
	flag.Parse()
}
func sendMessage(chatID int64, botToken string, str []byte) error {
	if len(str) > 15 {
		return segmentationFault{}
	}
	url := append([]byte("https://api.telegram.org/bot"), []byte(botToken)...)
	url = append(url, []byte("/sendMessage")...)
	request := &StandardMessage{
		Chat_ID: chatID,
		Text:    string(str)}
	postBody, _ := json.Marshal(request)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(string(url), "application/json", responseBody)
	if err != nil {
		log.Println("Error when attempting to send message:", err)
	}
	defer resp.Body.Close()
	return nil
}
func getTrees(oSlug, rSlug, pat string) {
	url := append([]byte("https://api.sourcecraft.tech/repos/"), []byte(oSlug)...)
	url = append(url, []byte("/")...)
	url = append(url, []byte(rSlug)...)
	url = append(url, []byte("/trees")...) // don't know if conversion to bytes is of any use here
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", string(url), nil)
	if err != nil {
		log.Fatal("Error when creating request:", err)
	}
	bearer := "Bearer " + pat
	req.Header.Add("Authorization", bearer)
	resp, err := client.Do(req) // bc of this
	if err != nil {
		log.Fatal("Error when receiving source API response:", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error when reading response body.", err)
	}
	log.Println(string(body))
	var items StandardResponse
	err = json.Unmarshal(body, &items)
	if err != nil {
		log.Fatal("Error when Unmarshalling reponse body.", err)
	}
	for _, i := range items.Trees {
		sendMessage(*chatID, *TGtoken, []byte(i.Path))
	}
}
func main() {
	getTrees(*orgSlug, *repoSlug, *PATtoken)
}
