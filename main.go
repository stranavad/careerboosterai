package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"os"
)

var openAiClient *openai.Client

func init() {
	fmt.Println("Initializing env...")
	envErr := godotenv.Load()
	if envErr != nil {
		fmt.Println("Error loading .env file")
	}
	loadOpenAIClient(&openAiClient)
}

func loadOpenAIClient(client **openai.Client) {
	*client = openai.NewClient(os.Getenv("OPENAI_KEY"))
}

var client = &http.Client{}

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
	response, err := apiRequest("/private/user-profile/"+userId, "GET", nil)

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

func apiRequest(url, method string, body *bytes.Buffer) (*http.Response, error) {
	var req *http.Request
	var err error

	baseUrl := os.Getenv("BASE_URL")

	if body != nil {
		req, err = http.NewRequest(method, baseUrl+url, body)
	} else {
		req, err = http.NewRequest(method, baseUrl+url, nil)
	}

	if err != nil {
		return nil, err
	}

	authHeader := os.Getenv("AUTH_HEADER")

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authHeader)
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return response, nil
}

type UserSearch struct {
	UserIDS []string `json:"userIds"`
}

type ScoreResponse struct {
	Score uint `json:"score"`
}

func getUserIds() []string {
	/* REQUEST FOR FILTER */
	requestBody := bytes.NewBuffer([]byte(`{"tags": []}`))
	response, err := apiRequest("/private/user-profiles/search", "POST", requestBody)

	if err != nil {
		fmt.Println("Error loading user ids")
		fmt.Println(err)
		return []string{}
	}

	var userIds UserSearch
	body, err := io.ReadAll(response.Body)
	fmt.Println(body)
	err = json.Unmarshal(body, &userIds)
	fmt.Println(userIds)

	if err != nil {
		fmt.Println(err)
		fmt.Println("Error deserializing user ids")
		return []string{}
	}

	return userIds.UserIDS
}

func getJobByRequestId(requestId string) *Job {
	response, err := apiRequest("/private/job/by-request/"+requestId, "GET", nil)

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

func generateUser(data GenerateUserRequest) {
	jsonData, err := json.Marshal(data)

	if err != nil {
		fmt.Println("Error jsoning")
		return
	}

	_, err = apiRequest("/private/ai/generate-user", "POST", bytes.NewBuffer(jsonData))

	if err != nil {
		fmt.Println(err)
	}
}

func getGptMatch(job, user string) *ScoreResponse {
	resp, err := openAiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4TurboPreview,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "I need you to check if candidate is relevant for this job post, " +
						"I will send you the candidate info in stringified json and the job post in stringified json as well" +
						"And I need you to return json -> {'score' : score} -> where score will be a number between 0 and 100 (included) based on how relevant the candidate is for this job post",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Candidate: " + user + "   ... and job post: " + job,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		},
	)

	if err != nil {
		fmt.Println(err)
		fmt.Println("Error generating gpt response")
		return nil
	}

	var score ScoreResponse
	fmt.Println(user)
	fmt.Println(job)
	fmt.Println(resp.Choices[0].Message.Content)
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &score)

	if err != nil {
		fmt.Println("Error serializing GPT response")
		return nil
	}

	return &score
}

type GenerateUserRequest struct {
	RequestID string `json:"requestId"`
	UserID    string `json:"userId"`
	Score     uint   `json:"score"`
}

func process(requestId string) {
	job := getJobByRequestId(requestId)
	if job == nil {
		fmt.Println("No job in process")
		return
	}

	userIds := getUserIds()

	if len(userIds) == 0 {
		fmt.Println("No users found")
		return
	}

	for _, v := range userIds {
		user := getUserProfileByUserId(v)

		if user == nil {
			fmt.Println("No user in iteration")
			continue
		}

		aiJob, err := json.Marshal(job.toAIJob())
		var aiJobString string
		if err != nil {
			fmt.Println("Failed converting job to ai json")
			continue
		} else {
			aiJobString = string(aiJob)
		}

		aiUser, err := json.Marshal(user.toAIUserProfile())
		var aiUserString string
		if err != nil {
			fmt.Println("Failed converting user to ai json")
			continue
		} else {
			aiUserString = string(aiUser)
		}

		gptResponse := getGptMatch(aiJobString, aiUserString)

		if gptResponse == nil {
			// Second try may fix few things
			gptResponse = getGptMatch(aiJobString, aiUserString)
		}

		fmt.Println(gptResponse)

		if gptResponse != nil {
			requestData := GenerateUserRequest{
				RequestID: requestId,
				UserID:    v,
				Score:     gptResponse.Score,
			}
			generateUser(requestData)
		}
	}
}

func main() {
	ctx := context.Background()
	opt, _ := redis.ParseURL(os.Getenv("REDIS_URL"))
	client := redis.NewClient(opt)

	pubsub := client.Subscribe(ctx, "ai-job-match")
	ch := pubsub.Channel()

	fmt.Println("Started")

	for msg := range ch {
		requestId := msg.Payload
		go process(requestId)
	}
}
