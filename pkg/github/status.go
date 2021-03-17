package github

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"

	"github.com/google/go-github/v32/github"
)

type StatusViewer struct {
	client *github.Client
}

func NewPullRequestStatusViewer(ctx context.Context, ghToken string) *StatusViewer {
	return &StatusViewer{client: github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken})))}
}

func (p *StatusViewer) Merged(ctx context.Context, url string) error {
	owner, repo, number, err := parsePullRequestMeta(url)
	if err != nil {
		return err
	}
	merged, _, err := p.client.PullRequests.IsMerged(ctx, owner, repo, number)
	_, _, err = p.client.Issues.AddLabelsToIssue(ctx, owner, repo, number, []string{"cherry-pick-approved"})
	if err == nil && merged == false {
		err = fmt.Errorf("not merged")
	}
	return err
}

