package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

// Client wraps the Google Gemini API client.
type Client struct {
	client *genai.Client
}

// NewClient initializes the AI client.
// It automatically picks up the GEMINI_API_KEY environment variable.
func NewClient() (*Client, error) {
	client, err := genai.NewClient(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Client{client: client}, nil
}

// EnrichmentResult represents the structured data extracted by the LLM.
type EnrichmentResult struct {
	Summary      string   `json:"summary"`
	Topics       []string `json:"topics"`
	WhyItMatters string   `json:"why_it_matters"`
}

// EnrichContent asks the LLM to process the raw content and return structured metadata.
func (c *Client) EnrichContent(ctx context.Context, title string, body string) (EnrichmentResult, error) {
	prompt := fmt.Sprintf(`Analyze the following technology content:
Title: %s

Content (or summary):
%s

Extract the following information:
1. summary: A concise 1-2 sentence summary of what this is about.
2. topics: A list of the core technologies, frameworks, or concepts discussed (e.g. "Kubernetes", "Go", "Distributed Systems"). Extract between 1 and 4 topics. Keep them generic and reusable.
3. why_it_matters: A short, opinionated sentence explaining why a software engineer should care about this.`, title, body)

	// Enforce strict JSON output matching our schema
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"summary": {Type: genai.TypeString},
			"topics": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"why_it_matters": {Type: genai.TypeString},
		},
		Required: []string{"summary", "topics", "why_it_matters"},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	// gemini-2.5-flash is extremely fast, cheap, and perfect for structured extraction tasks.
	resp, err := c.client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), config)
	if err != nil {
		return EnrichmentResult{}, fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return EnrichmentResult{}, fmt.Errorf("no content returned from model")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if part.Text == "" {
		return EnrichmentResult{}, fmt.Errorf("unexpected empty or non-text response part")
	}

	var result EnrichmentResult
	if err := json.Unmarshal([]byte(part.Text), &result); err != nil {
		return EnrichmentResult{}, fmt.Errorf("failed to unmarshal JSON: %w (raw: %s)", err, part.Text)
	}

	return result, nil
}
