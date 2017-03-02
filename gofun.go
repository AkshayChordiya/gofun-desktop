package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/andlabs/ui"
	"github.com/howeyc/gopass"
	"github.com/ricardolonga/jsongo"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Server Endpoint for authenticating and returning Auth-Token
const TOKEN_API_URL = "http://ttl-server.herokuapp.com/api/api-token-auth/"

// Server Endpoint for starting the countdown time ⏰
const START_TIME_API_URL = "http://ttl-server.herokuapp.com/api/entry/"

const GO_FUN_MESSAGE = "Go home and have fun, buddy!"

// The data structure to un-marshal JSON from server into it's object
type Token struct {
	Token string `json:"token"`
}

// The beginning of the end!
func main() {
	StartTime()
}

// Starts the countdown if the user is already logged in else prompts the user to
// authenticate with the server
func StartTime() {
	fmt.Println("Welcome to GoFun")
	token := GetToken()
	if token != "" {
		beginTime(token)
	} else {
		token = AuthenticateUser()
		beginTime(token)
	}
}

// Send the time start request to server and start the countdown
func beginTime(token string) {
	// Start time here
	values := jsongo.Object().Put("in_time", GetCurrentTimestamp())
	request, _ := http.NewRequest("POST", START_TIME_API_URL, bytes.NewBufferString(values.String()))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Token "+token)
	client := &http.Client{}
	response, _ := client.Do(request)
	if response.StatusCode == 201 {
		// We are good to go!
		fmt.Println("Commencing countdown ⏰ to Go Home 🏠")
		hour, min := 7, 30
		endTime := GetWorkEndTime(hour, min)

		fmt.Println("Going Home at", endTime.Format(time.Kitchen), "after", endTime.Sub(time.Now()))

		// Create new ticker
		ticker := time.NewTicker(time.Minute * 1)
		go StartTimeTicker(*time.NewTicker(time.Minute * 1), endTime)

		// Let's keep the app running till working hour is complete
		<-time.After(time.Duration(hour*60+min) * time.Minute)

		// Completion hours are complete, let's stop the timer
		ticker.Stop()
		ShowGoHomeWindow(GO_FUN_MESSAGE)
	}
}

// Authenticates the user with the entered credentials to the server and
// stores the auth-token returned by server into the file.
// Returns true if everything goes as planned else false.
func AuthenticateUser() string {
	var token Token
	// Read user credentials
	username, password := ReadUserCredentials()
	values := jsongo.Object().Put("username", username).Put("password", password)
	request, _ := http.NewRequest("POST", TOKEN_API_URL, bytes.NewBufferString(values.String()))
	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	response, _ := client.Do(request)
	body, _ := ioutil.ReadAll(response.Body)
	// Un-marshal JSON into object
	json.Unmarshal(body, &token)
	// Persist the object aka Token into file
	file, _ := os.Create("../token")
	file.WriteString(token.Token)
	return token.Token
}

// Show dialog to notify user about completion of his/her working hours
// and can go home to live :-P
// It shows the message provided in the parameter
func ShowGoHomeWindow(message string) {
	err := ui.Main(func() {
		home := ui.NewLabel(message)

		// Layout
		box := ui.NewVerticalBox()
		box.Append(home, false)

		// Window
		window := ui.NewWindow("Go Home!", 200, 100, false)
		window.SetChild(box)
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
	if err != nil {
		panic(err)
	}
}

// Get username and password from the user and returns them.
// It reads the credentials from console
func ReadUserCredentials() (username string, password string) {
	// Reading the user credentials
	consoleReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter email address: ")
	username, _ = consoleReader.ReadString('\n')
	fmt.Print("Enter password: ")
	bpassword, _ := gopass.GetPasswdMasked()
	password = string(bpassword)
	return
}

// Prints the remaining time periodically using Ticker
// endTime is the exact time
func StartTimeTicker(ticker time.Ticker, endTime time.Time) {
	for t := range ticker.C {
		fmt.Println("Going Home at", endTime.Format(time.Kitchen), "after", endTime.Sub(t))
	}
}

// Get the exact time when the working hours will end.
// Initially it checks for saved instance of end time
// in the database if found returns it else builds fresh end time.
// It returns the exact time (Ex. HH:MM => 05:30)
func GetWorkEndTime(hour, min int) time.Time {
	t := time.Now()
	t = t.Add(time.Hour * time.Duration(hour))
	t = t.Add(time.Minute * time.Duration(min))
	return t
}

// Reads the token from file and returns it if the token exists
func GetToken() string {
	tokenFile, _ := ioutil.ReadFile("../token")
	return string(tokenFile)
}

// Returns the current timestamp in millis
func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
