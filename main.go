package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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

	deleteReviewers := ""
	ignoreReviewers := ""

	flag.StringVar(&deleteReviewers, "delete-reviewers", "", "to remove a user as default reviwer on all repos")
	flag.StringVar(&ignoreReviewers, "ignore-reviewers", "", "exclude user from being added as default reviwer")

	flag.Parse()

	if len(deleteReviewers) > 0 {
		deleteReviewersOnAllRepos(deleteReviewers, client, workspace)
		return
	}

	// Getting members
	resp, err := client.R().
		SetResult(&BitbucketMembersResponse{}).
		Get(fmt.Sprintf("/workspaces/%s/members", workspace))

	if err != nil {
		log.Fatalf("Couldnt do request err: %v", err)
	}

	members := resp.Result().(*BitbucketMembersResponse)

	listIgnoreReviewers := strings.Split(ignoreReviewers, ",")

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

		// resp, err = client.R().
		// 	SetResult(&BitbucketRepositoriesResponse{}).
		// 	Delete(fmt.Sprintf("/repositories/%s/%s/default-reviewers/%s", workspace, r.Slug, "{xxxx}"))

		// if err != nil {
		// 	log.Fatalf("Couldnt do request err: %v", err)
		// }

		// resp, err = client.R().
		// 	SetResult(&BitbucketRepositoriesResponse{}).
		// 	Delete(fmt.Sprintf("/repositories/%s/%s/default-reviewers/%s", workspace, r.Slug, "{xxxx}"))

		// if err != nil {
		// 	log.Fatalf("Couldnt do request err: %v", err)
		// }

		for _, m := range members.Values {
			ignore := false
			for _, im := range listIgnoreReviewers {
				if im == m.User.AccountID {
					ignore = true
				}
			}

			if ignore {
				log.Printf("User was on ignore list %s", m.User.UUID)
				continue
			}

			resp, err = client.R().
				SetResult(&BitbucketRepositoriesResponse{}).
				Put(fmt.Sprintf("/repositories/%s/%s/default-reviewers/%s", workspace, r.Slug, m.User.UUID))

			if err != nil {
				log.Fatalf("Couldnt do request err: %v", err)
			}
			log.Printf("Added default reviewers for %s body %s", r.Slug, string(resp.Body()))
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
			SetHeader("Content-Type", "application/json").
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
			SetHeader("Content-Type", "application/json").
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
			SetHeader("Content-Type", "application/json").
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
			SetHeader("Content-Type", "application/json").
			Post(fmt.Sprintf("/repositories/%s/%s/branch-restrictions", workspace, r.Slug))

		if err != nil {
			log.Fatalf("Couldnt do request err: %v", err)
		}
	}

}

func deleteReviewersOnAllRepos(reviewersToDelete string, client *resty.Client, workspace string) {
	listOfReviewersToDelete := strings.Split(reviewersToDelete, ",")

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

	for _, r := range repos {
		for _, m := range members.Values {
			for _, u := range listOfReviewersToDelete {
				if u == m.User.AccountID {
					resp, err = client.R().
						SetResult(&BitbucketRepositoriesResponse{}).
						Delete(fmt.Sprintf("/repositories/%s/%s/default-reviewers/%s", workspace, r.Slug, m.User.UUID))

					if err != nil {
						log.Fatalf("Couldnt do request err: %v", err)
					}
					log.Printf("Deleted default for user %s reviewers for %s body %s", m.User.Nickname, r.Slug, string(resp.Body()))
					break
				}
			}
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
