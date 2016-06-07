package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const api_token = "API_TOKEN_HERE"
const api_url = "https://api.telegram.org/bot"
const debug = 1 // 0 = disable debugging, 1 = enable debugging

type Response struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	ErrorCode   int    `json:"error_code"`
}

type UserResponse struct {
	Response
	User User `json:"result"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type UpdateResponse struct {
	Response
	Update []Update `json:"result"`
}

type Update struct {
	ID      int     `json:"update_id"`
	Message Message `json:"message"`
}

type Message struct {
	ID   int    `json:"message_id"`
	User User   `json:"from"`
	Text string `json:"text"`
	Chat Chat   `json:chat`
}

type Chat struct {
	ID int `json:id`
}

func get_fortune_cookie() (string, error) {
	// get fortune cookie and normalize it
	out, err := exec.Command("fortune").Output()
	out_string := string(out)
	if debug == 1 {
		fmt.Println("\n\n======\nPrecomputed fortune:\n" + out_string)
	}
	out_string = strings.Replace(out_string, "\"", "'", -1)

	return out_string, err
}

func loop() {
	var update UpdateResponse
	var message string
	var chatid string
	var offset int
	var url_getUpdates = api_url + api_token + "/getUpdates"
	var url_sendMessage = api_url + api_token + "/sendMessage"
	client := &http.Client{}

	resp, err := http.Get(url_getUpdates)
	if err != nil {
		fmt.Println(err)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(contents, &update)

	message_start := len(update.Update) - 1
	if message_start > -1 {
		chatid = strconv.Itoa(update.Update[message_start].Message.Chat.ID)
		offset = update.Update[message_start].ID
	}

	message, _ = get_fortune_cookie()
	for {
		var jsonStr_update = []byte(`{"offset":` + strconv.Itoa(offset) + `}`)
		req, _ := http.NewRequest("POST", url_getUpdates, bytes.NewBuffer(jsonStr_update))
		req.Header.Set("Content-Type", "application/json")
		resp, err = client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		contents, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(contents, &update)

		if len(update.Update) < 1 {
			time.Sleep(3 * time.Second)
			continue
		}
		if offset < 1 {
			offset = update.Update[0].ID
		}
		chatid = strconv.Itoa(update.Update[0].Message.Chat.ID)

		if debug == 1 {
			fmt.Println("From User: " + update.Update[0].Message.User.Username)
			fmt.Println("Chat ID: " + chatid)
			fmt.Println("Received Message: " + update.Update[0].Message.Text)
		}

		if strings.Contains(update.Update[0].Message.Text, "/fortune") == true {
			if debug == 1 {
				fmt.Println("Sending message...")
			}
			var jsonStr_message = []byte(`{"text":"` + message + `", "chat_id":` + chatid + `}`)
			req, _ = http.NewRequest("POST", url_sendMessage, bytes.NewBuffer(jsonStr_message))
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)
			if err != nil {
				fmt.Println(err)
			}
			if debug == 1 {
				fmt.Println("Message sent.")
			}
		}

		message, _ = get_fortune_cookie()
		offset = offset + 1

	}
}

func main() {
	var botuser UserResponse

	fmt.Println("Starting...")

	if debug == 1 {
		fmt.Println("Debug Mode activated...")
		fmt.Println("Checking fortunes installation...")
	}

	// check 'fortune' installation
	_, err := get_fortune_cookie()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Please install 'fortune'")
		return
	}

	url := api_url + api_token + "/GetMe"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	json.Unmarshal(contents, &botuser)
	fmt.Println("Bot-ID: " + strconv.Itoa(botuser.User.ID))
	fmt.Println("Bot-Username: " + botuser.User.Username)

	loop()
}
