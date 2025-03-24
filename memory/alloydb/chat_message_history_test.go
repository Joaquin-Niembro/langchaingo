package alloydb_test

import (
	"context"
	alloydbutil2 "github.com/tmc/langchaingo/util/alloydbutil"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory/alloydb"
)

type chatMsg struct{}

func (c chatMsg) GetType() llms.ChatMessageType {
	return llms.ChatMessageTypeHuman
}

func (c chatMsg) GetContent() string {
	return "test content"
}

func getEnvVariables(t *testing.T) (string, string, string, string, string, string, string) {
	t.Helper()

	username := os.Getenv("ALLOYDB_USERNAME")
	if username == "" {
		t.Skip("ALLOYDB_USERNAME environment variable not set")
	}
	password := os.Getenv("ALLOYDB_PASSWORD")
	if password == "" {
		t.Skip("ALLOYDB_PASSWORD environment variable not set")
	}
	database := os.Getenv("ALLOYDB_DATABASE")
	if database == "" {
		t.Skip("ALLOYDB_DATABASE environment variable not set")
	}
	projectID := os.Getenv("ALLOYDB_PROJECT_ID")
	if projectID == "" {
		t.Skip("ALLOYDB_PROJECT_ID environment variable not set")
	}
	region := os.Getenv("ALLOYDB_REGION")
	if region == "" {
		t.Skip("ALLOYDB_REGION environment variable not set")
	}
	instance := os.Getenv("ALLOYDB_INSTANCE")
	if instance == "" {
		t.Skip("ALLOYDB_INSTANCE environment variable not set")
	}
	cluster := os.Getenv("ALLOYDB_CLUSTER")
	if cluster == "" {
		t.Skip("ALLOYDB_CLUSTER environment variable not set")
	}

	return username, password, database, projectID, region, instance, cluster
}

func setEngine(t *testing.T, ctx context.Context) (alloydbutil2.PostgresEngine, error) {
	username, password, database, projectID, region, instance, cluster := getEnvVariables(t)

	pgEngine, err := alloydbutil2.NewPostgresEngine(ctx,
		alloydbutil2.WithUser(username),
		alloydbutil2.WithPassword(password),
		alloydbutil2.WithDatabase(database),
		alloydbutil2.WithAlloyDBInstance(projectID, region, cluster, instance),
	)

	return *pgEngine, err
}

func TestValidateTable(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine, err := setEngine(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	tcs := []struct {
		desc      string
		tableName string
		sessionID string
		err       string
	}{
		{
			desc:      "Succesful creation of Chat Message History",
			tableName: "items",
			sessionID: "session",
			err:       "",
		},
		{
			desc:      "Creation of Chat Message History with missing table",
			tableName: "",
			sessionID: "session",
			err:       "",
		},
		{
			desc:      "Creation of Chat Message History with missing session ID",
			tableName: "items",
			sessionID: "",
			err:       "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			chatMsgHistory, err := alloydb.NewChatMessageHistory(ctx, engine, tc.tableName, tc.sessionID)
			if tc.err != "" && (err == nil || !strings.Contains(err.Error(), tc.err)) {
				t.Fatalf("unexpected error: got %q, want %q", err, tc.err)
			} else {
				errStr := err.Error()
				if errStr != tc.err {
					t.Fatalf("unexpected error: got %q, want %q", errStr, tc.err)
				}
			}
			// if the chat message history was created succesfully, continue with the other methods tests
			if err == nil {
				err = chatMsgHistory.AddMessage(ctx, chatMsg{})
				if err != nil {
					t.Fatal(err)
				}
				err = chatMsgHistory.AddAIMessage(ctx, "AI message")
				if err != nil {
					t.Fatal(err)
				}
				err = chatMsgHistory.AddUserMessage(ctx, "user message")
				if err != nil {
					t.Fatal(err)
				}
				err = chatMsgHistory.Clear(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
