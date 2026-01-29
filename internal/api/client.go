package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is the base HTTP client for API calls
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Get performs a GET request
func (c *Client) Get(path string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// Skill represents a unified skill from any registry
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Installs    int    `json:"installs"`
	Stars       int    `json:"stars"`
	Registry    string `json:"registry"`
}

// SearchAll searches all configured registries
func SearchAll(query string, limit int) ([]Skill, error) {
	var allSkills []Skill

	// Search skills.sh
	skillsShResults, err := SearchSkillsSh(query, limit)
	if err == nil {
		allSkills = append(allSkills, skillsShResults...)
	}

	// Search playbooks.com
	playbooksResults, err := SearchPlaybooks(query, limit)
	if err == nil {
		allSkills = append(allSkills, playbooksResults...)
	}

	// Deduplicate by name (prefer skills.sh for duplicates)
	seen := make(map[string]bool)
	var unique []Skill
	for _, s := range allSkills {
		if !seen[s.Name] {
			seen[s.Name] = true
			unique = append(unique, s)
		}
	}

	return unique, nil
}

// FetchSkillContent fetches SKILL.md content from GitHub
func FetchSkillContent(owner, repo, skillPath string) (string, error) {
	// Try common paths
	paths := []string{
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s/SKILL.md", owner, repo, skillPath),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/skills/%s/SKILL.md", owner, repo, skillPath),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/%s/SKILL.md", owner, repo, skillPath),
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, path := range paths {
		resp, err := client.Get(path)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			return string(body), nil
		}
	}

	return "", fmt.Errorf("SKILL.md not found for %s/%s/%s", owner, repo, skillPath)
}

// parseJSON is a helper to unmarshal JSON responses
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
