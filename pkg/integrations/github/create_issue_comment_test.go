package github

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/superplanehq/superplane/pkg/core"
	contexts "github.com/superplanehq/superplane/test/support/contexts"
)

func Test__CreateIssueComment__Setup(t *testing.T) {
	helloRepo := Repository{ID: 123456, Name: "hello", URL: "https://github.com/testhq/hello"}
	component := CreateIssueComment{}

	t.Run("repository is required", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{}
		err := component.Setup(core.SetupContext{
			Integration:   integrationCtx,
			Metadata:      &contexts.MetadataContext{},
			Configuration: map[string]any{"repository": ""},
		})

		require.ErrorContains(t, err, "repository is required")
	})

	t.Run("issue number is required", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}
		err := component.Setup(core.SetupContext{
			Integration: integrationCtx,
			Metadata:    &contexts.MetadataContext{},
			Configuration: map[string]any{
				"repository":  "hello",
				"issueNumber": "",
				"body":        "Test comment",
			},
		})

		require.ErrorContains(t, err, "issue number is required")
	})

	t.Run("body is required", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}
		err := component.Setup(core.SetupContext{
			Integration: integrationCtx,
			Metadata:    &contexts.MetadataContext{},
			Configuration: map[string]any{
				"repository":  "hello",
				"issueNumber": "42",
				"body":        "",
			},
		})

		require.ErrorContains(t, err, "body is required")
	})

	t.Run("repository is not accessible", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}
		err := component.Setup(core.SetupContext{
			Integration: integrationCtx,
			Metadata:    &contexts.MetadataContext{},
			Configuration: map[string]any{
				"repository":  "world",
				"issueNumber": "42",
				"body":        "Test comment",
			},
		})

		require.ErrorContains(t, err, "repository world is not accessible to app installation")
	})

	t.Run("metadata is set successfully", func(t *testing.T) {
		integrationCtx := &contexts.IntegrationContext{
			Metadata: Metadata{
				Repositories: []Repository{helloRepo},
			},
		}

		nodeMetadataCtx := contexts.MetadataContext{}
		require.NoError(t, component.Setup(core.SetupContext{
			Integration: integrationCtx,
			Metadata:    &nodeMetadataCtx,
			Configuration: map[string]any{
				"repository":  "hello",
				"issueNumber": "42",
				"body":        "Test comment",
			},
		}))

		require.Equal(t, nodeMetadataCtx.Get(), NodeMetadata{Repository: &helloRepo})
	})
}

func Test__CreateIssueComment__Configuration(t *testing.T) {
	component := CreateIssueComment{}

	t.Run("has correct fields", func(t *testing.T) {
		fields := component.Configuration()

		require.Len(t, fields, 3)

		// Repository field
		require.Equal(t, "repository", fields[0].Name)
		require.Equal(t, "Repository", fields[0].Label)
		require.True(t, fields[0].Required)

		// Issue Number field
		require.Equal(t, "issueNumber", fields[1].Name)
		require.Equal(t, "Issue Number", fields[1].Label)
		require.True(t, fields[1].Required)

		// Body field
		require.Equal(t, "body", fields[2].Name)
		require.Equal(t, "Body", fields[2].Label)
		require.True(t, fields[2].Required)
	})
}

func Test__CreateIssueComment__Metadata(t *testing.T) {
	component := CreateIssueComment{}

	t.Run("returns correct name", func(t *testing.T) {
		require.Equal(t, "github.createIssueComment", component.Name())
	})

	t.Run("returns correct label", func(t *testing.T) {
		require.Equal(t, "Create Issue Comment", component.Label())
	})

	t.Run("returns correct description", func(t *testing.T) {
		require.Equal(t, "Add a comment to a GitHub issue or pull request", component.Description())
	})

	t.Run("returns correct icon", func(t *testing.T) {
		require.Equal(t, "github", component.Icon())
	})

	t.Run("returns correct color", func(t *testing.T) {
		require.Equal(t, "gray", component.Color())
	})

	t.Run("returns default output channel", func(t *testing.T) {
		channels := component.OutputChannels(nil)
		require.Len(t, channels, 1)
		require.Equal(t, core.DefaultOutputChannel, channels[0])
	})
}
