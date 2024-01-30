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

func (u *UserProfile) toAIUserProfile() AIUserProfile {
	tags := make([]string, len(u.Skills))
	for i, v := range u.Skills {
		tags[i] = v.Label
	}

	return AIUserProfile{
		Description:    u.Description,
		Skills:         tags,
		Languages:      u.Languages,
		WorkHistory:    u.WorkHistory,
		Certifications: u.Certifications,
	}
}

type AIUserProfile struct {
	Description    string              `json:"description"`
	Skills         []string            `json:"skills"`
	Languages      []UserLanguage      `json:"languages"`
	WorkHistory    []UserWorkHistory   `json:"WorkHistory"`
	Certifications []UserCertification `json:"certifications"`
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
	CollaborationTypes []string `json:"collaborationTypes"`
	Bond               string   `json:"bond"`
	Remote             bool     `json:"remote"`
	MaxSalary          uint     `json:"maxSalary"`
	MinSalary          uint     `json:"minSalary"`
	Currency           string   `json:"string"`
	TagIds             []string `json:"tagIds"`
	Tags               []Tag    `json:"tags"`
}

func (j *Job) toAIJob() AIJob {
	tags := make([]string, len(j.Tags))

	for i, v := range j.Tags {
		tags[i] = v.Label
	}

	return AIJob{
		Name:               j.Name,
		Description:        j.Description,
		Reason:             j.Reason,
		CollaborationTypes: j.CollaborationTypes,
		Bond:               j.Bond,
		Remote:             j.Remote,
		Skills:             tags,
	}
}

type AIJob struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Reason             string   `json:"reason"`
	CollaborationTypes []string `json:"collaborationTypes"`
	Bond               string   `json:"bond"`
	Remote             bool     `json:"remote"`
	Skills             []string `json:"skills"`
	//MaxSalary          uint     `json:"maxSalary"`
	//MinSalary          uint     `json:"minSalary"`
	//Currency           string   `json:"string"`
	//TagIds             []string `json:"tagIds"`
	//Tags []Tag `json:"tags"`
}

func getUserProfileByUserId(userId string) *UserProfile {
	response, err := http.Get(baseUrl + "/private/user-profile/" + userId)

	if err != nil {
		fmt.Println("Error loading user profile")
		fmt.Println(err)
		return nil
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading using profile")
		fmt.Println(err)
		return nil
	}

	var profile UserProfile
	err = json.Unmarshal(data, &profile)

	if err != nil {
		fmt.Println("Error deserializing user profile")
		fmt.Println(err)
		return nil
	}

	return &profile
}

func getUserIds() []string {
	/* REQUEST FOR FILTER */
	requestBody := bytes.NewBuffer([]byte(`{"tags": []}`))
	postResponse, err := http.Post(baseUrl+"/private/user-profiles/search", "application/json", requestBody)

	if err != nil {
		fmt.Println("Error loading user ids")
		fmt.Println(er)
		return []string{}
	}

	var userIds []string
	body, err := io.ReadAll(postResponse.Body)
	err = json.Unmarshal(body, &userIds)

	if err != nil {
		fmt.Println(err)
		fmt.Println("Error deserializing user ids")
		return []string{}
	}

	return userIds
}

func getJobById(jobId string) *Job {
	response, err := http.Get(baseUrl + "/private/job/" + jobId)

	if err != nil {
		fmt.Println("Error getting job")
		fmt.Println(err)
		return nil
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading job response")
		fmt.Println(err)
		return nil
	}

	var job Job
	err = json.Unmarshal(data, &job)

	if err != nil {
		fmt.Println("Error deserializing job")
		fmt.Println(err)
		return nil
	}

	return &job
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
