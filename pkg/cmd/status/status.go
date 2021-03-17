package status

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	v1 "github.com/mfojtik/patchmanager/pkg/api/v1"
	"github.com/mfojtik/patchmanager/pkg/github"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

// statusOptions holds values to drive the status command.
type statusOptions struct {
	githubToken string
	inFile      string
}

// NewStatusCommand creates a status command.
func NewStatusCommand(ctx context.Context) *cobra.Command {
	runOpts := statusOptions{}
	cmd := &cobra.Command{
		Use: "status",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runOpts.Complete(); err != nil {
				klog.Exit(err)
			}
			if err := runOpts.Validate(); err != nil {
				klog.Exit(err)
			}
			if err := runOpts.Run(ctx); err != nil {
				klog.Exit(err)
			}
		},
	}

	runOpts.AddFlags(cmd.Flags())

	return cmd
}

func (r *statusOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&r.githubToken, "github-token", "", "Github Access Token (GITHUB_TOKEN env variable)")
	fs.StringVarP(&r.inFile, "file", "f", "", "Set input file to read the list of candidates")
}

func (r *statusOptions) Validate() error {
	if len(r.githubToken) == 0 {
		return fmt.Errorf("github-token flag must be specified or GITHUB_TOKEN environment must be set")
	}
	if len(r.inFile) == 0 {
		return fmt.Errorf("input file must be specified")
	}
	return nil
}

func (r *statusOptions) Complete() error {
	if len(r.githubToken) == 0 {
		r.githubToken = os.Getenv("GITHUB_TOKEN")
	}
	return nil
}

func (r *statusOptions) Run(ctx context.Context) error {
	content, err := ioutil.ReadFile(r.inFile)
	if err != nil {
		return err
	}

	var approved v1.ApprovedCandidateList

	if err := yaml.Unmarshal(content, &approved); err != nil {
		return err
	}

	status := github.NewPullRequestStatusViewer(ctx, r.githubToken)

	for _, pr := range approved.Items {
		if pr.PullRequest.Decision != "pick" {
			continue
		}

		if err := status.Merged(ctx, pr.PullRequest.URL); err != nil {
			fmt.Fprintf(os.Stdout, "%q: %v\n", pr.PullRequest.URL, err)
		} else {
			fmt.Fprintf(os.Stdout, "%q: merged\n", pr.PullRequest.URL)
		}
	}

	return nil
}
