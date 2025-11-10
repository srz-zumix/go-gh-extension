package gh

import (
	"context"
	"net/url"

	"github.com/cli/go-gh/v2/pkg/repository"
)

type OrganizationProfile struct {
	Name string
	Plan string
	Host string
}

func (o *OrganizationProfile) IsGitHubEnterprise() bool {
	return o.Plan == "enterprise"
}

func (o *OrganizationProfile) IsGitHubEnterpriseServer() bool {
	return !o.IsGitHubCom() && o.IsGitHubEnterprise()
}

func (o *OrganizationProfile) IsGitHubEnterpriseCloud() bool {
	return o.IsGitHubCom() && o.IsGitHubEnterprise()
}

func (o *OrganizationProfile) IsGitHubCom() bool {
	return o.Host == "" || o.Host == "github.com"
}

func (o *OrganizationProfile) IsUser() bool {
	return o.Plan == "user"
}

func (o *OrganizationProfile) IsFreePlan() bool {
	return o.Plan == "free"
}

func (o *OrganizationProfile) IsProPlan() bool {
	return o.Plan == "pro"
}

func (o *OrganizationProfile) IsTeamPlan() bool {
	return o.Plan == "team"
}

func GetOrganizationProfile(ctx context.Context, g *GitHubClient, repo repository.Repository) (*OrganizationProfile, error) {
	org, err := GetOrg(ctx, g, repo.Owner)
	if err != nil {
		return &OrganizationProfile{
			Name: repo.Owner,
			Plan: "user",
			Host: repo.Host,
		}, nil
	}
	url, _ := url.Parse(*org.HTMLURL)
	host := url.Hostname()
	return &OrganizationProfile{
		Name: org.GetLogin(),
		Plan: org.GetPlan().GetName(),
		Host: host,
	}, nil
}
