package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/alloydbutil"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/alloydb"
	"log"
	"os"
)

func getEnvVariables() (string, string, string, string, string, string, string, string, string) {
	// Requires environment variable ALLOYDB_USERNAME to be set.
	username := os.Getenv("ALLOYDB_USERNAME")
	if username == "" {
		log.Fatal("env variable ALLOYDB_USERNAME is empty")
	}
	// Requires environment variable ALLOYDB_PASSWORD to be set.
	password := os.Getenv("ALLOYDB_PASSWORD")
	if password == "" {
		log.Fatal("env variable ALLOYDB_PASSWORD is empty")
	}
	// Requires environment variable ALLOYDB_DATABASE to be set.
	database := os.Getenv("ALLOYDB_DATABASE")
	if database == "" {
		log.Fatal("env variable ALLOYDB_DATABASE is empty")
	}
	// Requires environment variable PROJECT_ID to be set.
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("env variable PROJECT_ID is empty")
	}
	// Requires environment variable ALLOYDB_REGION to be set.
	region := os.Getenv("ALLOYDB_REGION")
	if region == "" {
		log.Fatal("env variable ALLOYDB_REGION is empty")
	}
	// Requires environment variable ALLOYDB_INSTANCE to be set.
	instance := os.Getenv("ALLOYDB_INSTANCE")
	if instance == "" {
		log.Fatal("env variable ALLOYDB_INSTANCE is empty")
	}
	// Requires environment variable ALLOYDB_CLUSTER to be set.
	cluster := os.Getenv("ALLOYDB_CLUSTER")
	if cluster == "" {
		log.Fatal("env variable ALLOYDB_CLUSTER is empty")
	}
	// Requires environment variable ALLOYDB_TABLE to be set.
	table := os.Getenv("ALLOYDB_TABLE")
	if table == "" {
		log.Fatal("env variable ALLOYDB_TABLE is empty")
	}

	// Requires environment variable VERTEX_LOCATION to be set.
	location := os.Getenv("VERTEX_LOCATION")
	if location == "" {
		log.Fatal("env variable VERTEX_LOCATION is empty")
	}

	return username, password, database, projectID, region, instance, cluster, table, location
}

func main() {
	// Requires the Environment variables to be set as indicated in the getEnvVariables function.
	username, password, database, projectID, region, instance, cluster, table, vertexLocation := getEnvVariables()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithUser(username),
		alloydbutil.WithPassword(password),
		alloydbutil.WithDatabase(database),
		alloydbutil.WithAlloyDBInstance(projectID, region, cluster, instance),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Vectorstore table using InitVectorstoreTable method
	err = pgEngine.InitVectorstoreTable(ctx, "tableName", 4096, "public", "content",
		"embedding",
		[]alloydbutil.Column{
			alloydbutil.Column{
				Name:     "area",
				DataType: "int",
				Nullable: false,
			},
			alloydbutil.Column{
				Name:     "population",
				DataType: "int",
				Nullable: false,
			},
		},
		"langchain_metadata",
		alloydbutil.Column{Name: "langchain_id", DataType: "UUID", Nullable: false},
		false,
		true,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize VertexAI LLM
	llm, err := vertex.New(ctx, vertex.WithCloudProject(projectID), vertex.WithCloudLocation(vertexLocation), vertex.WithDefaultModel("text-embedding-005"))
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new alloydb vectorstore .

	vs, err := alloydb.NewVectorStore(ctx, pgEngine, e, table)

	_, err = vs.AddDocuments(ctx, []schema.Document{
		{
			PageContent: "Tokyo",
			Metadata: map[string]any{
				"population": 38,
				"area":       2190,
			},
		},
		{
			PageContent: "Paris",
			Metadata: map[string]any{
				"population": 11,
				"area":       105,
			},
		},
		{
			PageContent: "London",
			Metadata: map[string]any{
				"population": 9.5,
				"area":       1572,
			},
		},
		{
			PageContent: "Santiago",
			Metadata: map[string]any{
				"population": 6.9,
				"area":       641,
			},
		},
		{
			PageContent: "Buenos Aires",
			Metadata: map[string]any{
				"population": 15.5,
				"area":       203,
			},
		},
		{
			PageContent: "Rio de Janeiro",
			Metadata: map[string]any{
				"population": 13.7,
				"area":       1200,
			},
		},
		{
			PageContent: "Sao Paulo",
			Metadata: map[string]any{
				"population": 22.6,
				"area":       1523,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	docs, err := vs.SimilaritySearch(ctx, "Japan")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Docs:", docs)
}
