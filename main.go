package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

func main() {
	log.Print("Starting bitubkcet central config")

	workspace := os.Getenv("BITBUCKET_WORKSPACE")
	username := os.Getenv("BITBUCKET_USERNAME")
	password := os.Getenv("BITBUCKET_PASSWORD")

	client := resty.New()
	client.SetDebug(true)

	client.SetHostURL("https://api.bitbucket.org/2.0")
	client.SetBasicAuth(username, password)

	// Getting members
	resp, err := client.R().
		SetResult(&BitbucketMembersResponse{}).
		Get(fmt.Sprintf("/workspaces/%s/members", workspace))

	if err != nil {
		log.Fatalf("Couldnt do request err: %v", err)
	}

	members := resp.Result().(*BitbucketMembersResponse)

	log.Print(len(members.Values))

	// Get repos

	// Set default reviewers

	// Set Branching Model

	// Set Branch Restrictions

}

type BitbucketMembersResponse struct {
	Pagelen int64             `json:"pagelen"`
	Values  []BitbucketMember `json:"values"`
	Page    int64             `json:"page"`
	Size    int64             `json:"size"`
}

type BitbucketMember struct {
	Links     ValueLinks     `json:"links"`
	Type      string         `json:"type"`
	User      UserClass      `json:"user"`
	Workspace WorkspaceClass `json:"workspace"`
}

type ValueLinks struct {
	Self Self `json:"self"`
}

type Self struct {
	Href string `json:"href"`
}

type UserClass struct {
	DisplayName string    `json:"display_name"`
	UUID        string    `json:"uuid"`
	Links       UserLinks `json:"links"`
	Nickname    string    `json:"nickname"`
	Type        string    `json:"type"`
	AccountID   string    `json:"account_id"`
}

type UserLinks struct {
	Self   Self `json:"self"`
	HTML   Self `json:"html"`
	Avatar Self `json:"avatar"`
}

type WorkspaceClass struct {
	Slug  string    `json:"slug"`
	Type  string    `json:"type"`
	Name  string    `json:"name"`
	Links UserLinks `json:"links"`
	UUID  string    `json:"uuid"`
}
