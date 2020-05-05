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
	client.SetHeader("Content-Type", "application/json")
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

	repos := []BitbucketRepository{}

	page := 1
	// Get repos
	for {

		resp, err = client.R().
			SetResult(&BitbucketRepositoriesResponse{}).
			Get(fmt.Sprintf("/repositories/%s?page=%d", workspace, page))

		if err != nil {
			log.Fatalf("Couldnt do request err: %v", err)
		}

		reposResp := resp.Result().(*BitbucketRepositoriesResponse)

		if len(reposResp.Values) == 0 {
			break
		}

		for _, r := range reposResp.Values {
			repos = append(repos, r)
		}

		page++
	}

	// Set default reviewers

	for _, r := range repos {
		for _, m := range members.Values {
			resp, err = client.R().
				SetResult(&BitbucketRepositoriesResponse{}).
				Put(fmt.Sprintf("/repositories/%s/%s/default-reviewers/%s", workspace, r.Slug, m.User.UUID))

			if err != nil {
				log.Fatalf("Couldnt do request err: %v", err)
			}
			log.Printf("Added default reviewers for %s", r.Slug)
		}

		// Set Branching Model
		body := `{
			"branch_types": [
			  {
				"kind": "bugfix",
				"enabled": true,
				"prefix": "bugfix/"
			  },
			  {
				"kind": "feature",
				"enabled": true,
				"prefix": "feature/"
			  },
			  {
				"kind": "hotfix",
				"enabled": true,
				"prefix": "hotfix/"
			  }
			]
		  }`

		_, err = client.R().
			SetBody(body).
			Put(fmt.Sprintf("/repositories/%s/%s/branching-model/settings", workspace, r.Slug))

		if err != nil {
			log.Fatalf("Couldnt do request err: %v", err)
		}

		//Set Branch Restrictions
		bodyRequriesApproves := `{
			"kind": "require_approvals_to_merge",
			"branch_match_kind": "branching_model",
			"branch_type": "development",
			"pattern": "",
			"value": 2
		  }`
		_, err = client.R().
			SetBody(bodyRequriesApproves).
			Post(fmt.Sprintf("/repositories/%s/%s/branch-restrictions", workspace, r.Slug))

		if err != nil {
			log.Fatalf("Couldnt do request err: %v", err)
		}

		bodyResetApproves := `{
			"kind": "reset_pullrequest_approvals_on_change",
			"branch_match_kind": "branching_model",
			"branch_type": "development",
			"pattern": ""
		  }`
		_, err = client.R().
			SetBody(bodyResetApproves).
			Post(fmt.Sprintf("/repositories/%s/%s/branch-restrictions", workspace, r.Slug))

		if err != nil {
			log.Fatalf("Couldnt do request err: %v", err)
		}

		bodyPreventMergeWithUnresolved := `{
			"kind": "enforce_merge_checks",
			"branch_match_kind": "branching_model",
			"branch_type": "development",
			"pattern": ""
		  }`
		_, err = client.R().
			SetBody(bodyPreventMergeWithUnresolved).
			Post(fmt.Sprintf("/repositories/%s/%s/branch-restrictions", workspace, r.Slug))

		if err != nil {
			log.Fatalf("Couldnt do request err: %v", err)
		}
	}

}

type BitbucketMembersResponse struct {
	Pagelen int64             `json:"pagelen"`
	Values  []BitbucketMember `json:"values"`
	Page    int64             `json:"page"`
	Size    int64             `json:"size"`
}

type BitbucketMember struct {
	Links     BitbucketMemberValueLinks `json:"links"`
	Type      string                    `json:"type"`
	User      UserClass                 `json:"user"`
	Workspace WorkspaceClass            `json:"workspace"`
}

type BitbucketMemberValueLinks struct {
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

type BitbucketRepositoriesResponse struct {
	Pagelen int64                 `json:"pagelen"`
	Size    int64                 `json:"size"`
	Values  []BitbucketRepository `json:"values"`
	Page    int64                 `json:"page"`
	Next    string                `json:"next"`
}

type BitbucketRepository struct {
	SCM         string                        `json:"scm"`
	Website     string                        `json:"website"`
	HasWiki     bool                          `json:"has_wiki"`
	UUID        string                        `json:"uuid"`
	Links       BitbucketRepositoryValueLinks `json:"links"`
	ForkPolicy  string                        `json:"fork_policy"`
	FullName    string                        `json:"full_name"`
	Name        string                        `json:"name"`
	Project     ProjectClass                  `json:"project"`
	Language    string                        `json:"language"`
	CreatedOn   string                        `json:"created_on"`
	Mainbranch  *Mainbranch                   `json:"mainbranch"`
	Workspace   ProjectClass                  `json:"workspace"`
	HasIssues   bool                          `json:"has_issues"`
	Owner       Owner                         `json:"owner"`
	UpdatedOn   string                        `json:"updated_on"`
	Size        int64                         `json:"size"`
	Type        string                        `json:"type"`
	Slug        string                        `json:"slug"`
	IsPrivate   bool                          `json:"is_private"`
	Description string                        `json:"description"`
}

type BitbucketRepositoryValueLinks struct {
	Watchers     Avatar  `json:"watchers"`
	Branches     Avatar  `json:"branches"`
	Tags         Avatar  `json:"tags"`
	Commits      Avatar  `json:"commits"`
	Clone        []Clone `json:"clone"`
	Self         Avatar  `json:"self"`
	Source       Avatar  `json:"source"`
	HTML         Avatar  `json:"html"`
	Avatar       Avatar  `json:"avatar"`
	Hooks        Avatar  `json:"hooks"`
	Forks        Avatar  `json:"forks"`
	Downloads    Avatar  `json:"downloads"`
	Pullrequests Avatar  `json:"pullrequests"`
}

type Avatar struct {
	Href string `json:"href"`
}

type Clone struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type Mainbranch struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type Owner struct {
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Type        string     `json:"type"`
	UUID        string     `json:"uuid"`
	Links       OwnerLinks `json:"links"`
}

type OwnerLinks struct {
	Self   Avatar `json:"self"`
	HTML   Avatar `json:"html"`
	Avatar Avatar `json:"avatar"`
}

type ProjectClass struct {
	Links OwnerLinks `json:"links"`
	Type  string     `json:"type"`
	UUID  string     `json:"uuid"`
	Key   *string    `json:"key,omitempty"`
	Name  string     `json:"name"`
	Slug  *string    `json:"slug,omitempty"`
}
