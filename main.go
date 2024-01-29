package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
)

/* CONFIG */
var baseUrl = "http://localhost:3001"
var openAiClient = openai.NewClient("")

type Address struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type UserWorkHistory struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Url         string  `json:"url"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	StartDate   string  `json:"startDate"`
	EndDate     *string `json:"endDate"`
}

type UserCertification struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Url       string  `json:"url"`
	Title     string  `json:"title"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
}

type UserLanguage struct {
	Language string `json:"language"`
	Level    string `json:"level"`
}

type UserSocialSite struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type UserProfile struct {
	Banner         *string             `json:"banner"`
	Address        *Address            `json:"address"`
	Description    string              `json:"description"`
	Name           string              `json:"name"`
	Avatar         *string             `json:"avatar"`
	Skills         []Tag               `json:"skills"`
	SocialSites    []UserSocialSite    `json:"SocialSites"`
	WorkHistory    []UserWorkHistory   `json:"WorkHistory"`
	Certifications []UserCertification `json:"certifications"`
	Languages      []UserLanguage      `json:"languages"`
}

type Tag struct {
	Id    string `json:"id"`
	Label string `json:"label"`
	Key   string `json:"key"`
}

type Job struct {
	Id                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Reason             string   `json:"reason"`
	CollaborationTypes []string `json:"CollaborationTypes"`
	Bond               string   `json:"bond"`
	Remote             bool     `json:"remote"`
	MaxSalary          uint     `json:"maxSalary"`
	MinSalary          uint     `json:"minSalary"`
	Currency           string   `json:"string"`
	TagIds             []string `json:"tagIds"`
	Tags               []Tag    `json:"tags"`
}

func getUserProfileByUserId(userId string) UserProfile {
	response, err := http.Get(baseUrl + "/private/user-profile/" + userId)
	if err != nil {
		panic("Error getting user profile")
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic("Error reading user profile response")
	}

	var profile UserProfile
	err = json.Unmarshal(data, &profile)
	if err != nil {
		panic("Error deserializing user profile")
	}
	return profile
}

func getUserIds() []string {
	/* REQUEST FOR FILTER */
	requestBody := bytes.NewBuffer([]byte(`{"tags": []}`))
	postResponse, err := http.Post(baseUrl+"/private/user-profiles/search", "application/json", requestBody)

	if err != nil {
		panic("Error loading user ids")
	}

	var userIds []string
	body, err := io.ReadAll(postResponse.Body)
	err = json.Unmarshal(body, &userIds)
	if err != nil {
		panic("Error deserializing user ids")
	}

	return userIds
}

func getJobById(jobId string) Job {
	response, err := http.Get(baseUrl + "/private/job/" + jobId)
	if err != nil {
		panic("Error getting job")
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic("Error reading job response")
	}

	var job Job
	err = json.Unmarshal(data, &job)
	if err != nil {
		panic("Error deserializing job")
	}

	return job
}

func main() {
	ctx := context.Background()
	opt, _ := redis.ParseURL("")
	client := redis.NewClient(opt)

	pubsub := client.Subscribe(ctx, "ai-job-match")
	ch := pubsub.Channel()

	fmt.Println("Started")

	for msg := range ch {
		jobId := msg.Payload

		job := getJobById(jobId)
		fmt.Println(job)
		userIds := getUserIds()

		for _, v := range userIds {
			user := getUserProfileByUserId(v)
			fmt.Println(user)
			resp, err := openAiClient.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT4TurboPreview,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleSystem,
							Content: "Return either {'correct': true} or {'correct': false} based randomly, no matter the user message in json",
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: "Return random boolean",
						},
					},
					ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
				},
			)
			if err != nil {
				fmt.Printf("ChatCompletion error: %v\n", err)
				return
			}

			fmt.Println(resp.Choices[0].Message.Content)
		}
	}
}
