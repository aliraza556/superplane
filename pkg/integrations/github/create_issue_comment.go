package github

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/go-github/v74/github"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

type CreateIssueComment struct{}

type CreateIssueCommentConfiguration struct {
	Repository  string `json:"repository" mapstructure:"repository"`
	IssueNumber string `json:"issueNumber" mapstructure:"issueNumber"`
	Body        string `json:"body" mapstructure:"body"`
}

func (c *CreateIssueComment) Name() string {
	return "github.createIssueComment"
}

func (c *CreateIssueComment) Label() string {
	return "Create Issue Comment"
}

func (c *CreateIssueComment) Description() string {
	return "Add a comment to a GitHub issue or pull request"
}

func (c *CreateIssueComment) Documentation() string {
	return `The Create Issue Comment component adds a comment to a GitHub issue or pull request.

## Use Cases

- **Deployment status updates**: Post deployment status or remediation updates to GitHub issues from SuperPlane
- **Runbook linking**: Add runbook links, error details, or status for responders
- **Cross-platform sync**: Sync Slack or PagerDuty notes into GitHub as comments
- **Automated notifications**: Add automated workflow status updates to issues
- **Bot interactions**: Post automated responses or acknowledgments to issues

## Configuration

- **Repository**: Select the GitHub repository containing the issue (required)
- **Issue Number**: The issue or PR number to comment on (supports expressions, required)
- **Body**: The comment text - supports Markdown formatting (required)

## Output

Returns the created comment object with details including:
- Comment ID
- Body text
- Author information
- Created timestamp
- HTML URL to the comment`
}

func (c *CreateIssueComment) Icon() string {
	return "github"
}

func (c *CreateIssueComment) Color() string {
	return "gray"
}

func (c *CreateIssueComment) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *CreateIssueComment) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:     "repository",
			Label:    "Repository",
			Type:     configuration.FieldTypeIntegrationResource,
			Required: true,
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type:           "repository",
					UseNameAsValue: true,
				},
			},
		},
		{
			Name:        "issueNumber",
			Label:       "Issue Number",
			Type:        configuration.FieldTypeString,
			Required:    true,
			Placeholder: "e.g., 42",
			Description: "The issue or PR number to comment on",
		},
		{
			Name:        "body",
			Label:       "Body",
			Type:        configuration.FieldTypeText,
			Required:    true,
			Placeholder: "Enter your comment (Markdown supported)",
			Description: "The comment text - supports Markdown formatting",
		},
	}
}

func (c *CreateIssueComment) Setup(ctx core.SetupContext) error {
	// First validate repository access
	if err := ensureRepoInMetadata(
		ctx.Metadata,
		ctx.Integration,
		ctx.Configuration,
	); err != nil {
		return err
	}

	// Then validate other required fields
	var config CreateIssueCommentConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	if config.IssueNumber == "" {
		return errors.New("issue number is required")
	}

	if config.Body == "" {
		return errors.New("body is required")
	}

	return nil
}

func (c *CreateIssueComment) Execute(ctx core.ExecutionContext) error {
	var config CreateIssueCommentConfiguration
	if err := mapstructure.Decode(ctx.Configuration, &config); err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}

	issueNumber, err := strconv.Atoi(config.IssueNumber)
	if err != nil {
		return fmt.Errorf("issue number is not a valid number: %w", err)
	}

	var appMetadata Metadata
	if err := mapstructure.Decode(ctx.Integration.GetMetadata(), &appMetadata); err != nil {
		return fmt.Errorf("failed to decode application metadata: %w", err)
	}

	client, err := NewClient(ctx.Integration, appMetadata.GitHubApp.ID, appMetadata.InstallationID)
	if err != nil {
		return fmt.Errorf("failed to initialize GitHub client: %w", err)
	}

	// Create the comment
	comment, _, err := client.Issues.CreateComment(
		context.Background(),
		appMetadata.Owner,
		config.Repository,
		issueNumber,
		&github.IssueComment{
			Body: &config.Body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create issue comment: %w", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		"github.issueComment",
		[]any{comment},
	)
}

func (c *CreateIssueComment) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *CreateIssueComment) HandleWebhook(ctx core.WebhookRequestContext) (int, error) {
	return 200, nil
}

func (c *CreateIssueComment) Actions() []core.Action {
	return []core.Action{}
}

func (c *CreateIssueComment) HandleAction(ctx core.ActionContext) error {
	return nil
}

func (c *CreateIssueComment) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *CreateIssueComment) Cleanup(ctx core.SetupContext) error {
	return nil
}
