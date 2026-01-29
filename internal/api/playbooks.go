package api

import (
	"fmt"
)

const playbooksBaseURL = "https://playbooks.com"

// PlaybooksResponse represents the response from playbooks.com API
type PlaybooksResponse struct {
	Success bool             `json:"success"`
	Data    []PlaybooksSkill `json:"data"`
}

// PlaybooksSkill represents a skill from playbooks.com
type PlaybooksSkill struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ShortDescription string `json:"shortDescription"`
	RepoOwner        string `json:"repoOwner"`
	RepoName         string `json:"repoName"`
	Path             string `json:"path"`
	SkillSlug        string `json:"skillSlug"`
	Stars            int    `json:"stars"`
	IsOfficial       bool   `json:"isOfficial"`
}

// SearchPlaybooks searches playbooks.com API
func SearchPlaybooks(query string, limit int) ([]Skill, error) {
	client := NewClient(playbooksBaseURL)

	params := map[string]string{
		"search": query,
		"limit":  fmt.Sprintf("%d", limit),
	}

	data, err := client.Get("/api/skills", params)
	if err != nil {
		return nil, err
	}

	var response PlaybooksResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("playbooks API returned success=false")
	}

	var skills []Skill
	for _, s := range response.Data {
		source := s.RepoOwner
		if s.RepoName != "" {
			source = fmt.Sprintf("%s/%s", s.RepoOwner, s.RepoName)
		}

		skills = append(skills, Skill{
			ID:          s.SkillSlug,
			Name:        s.Name,
			Source:      source,
			Description: s.ShortDescription,
			Stars:       s.Stars,
			Registry:    "playbooks.com",
		})
	}

	return skills, nil
}

// GetPlaybooksTrending gets trending skills from playbooks.com
func GetPlaybooksTrending(limit int) ([]Skill, error) {
	client := NewClient(playbooksBaseURL)

	params := map[string]string{
		"limit": fmt.Sprintf("%d", limit),
	}

	data, err := client.Get("/api/skills", params)
	if err != nil {
		return nil, err
	}

	var response PlaybooksResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	var skills []Skill
	for _, s := range response.Data {
		source := s.RepoOwner
		if s.RepoName != "" {
			source = fmt.Sprintf("%s/%s", s.RepoOwner, s.RepoName)
		}

		skills = append(skills, Skill{
			ID:          s.SkillSlug,
			Name:        s.Name,
			Source:      source,
			Description: s.ShortDescription,
			Stars:       s.Stars,
			Registry:    "playbooks.com",
		})
	}

	return skills, nil
}
