package api

import (
	"fmt"
)

const skillsShBaseURL = "https://skills.sh"

// SkillsShResponse represents the response from skills.sh API
type SkillsShResponse struct {
	Skills []SkillsShSkill `json:"skills"`
}

// SkillsShSkill represents a skill from skills.sh
type SkillsShSkill struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Installs  int    `json:"installs"`
	TopSource string `json:"topSource"`
}

// SearchSkillsSh searches skills.sh API
func SearchSkillsSh(query string, limit int) ([]Skill, error) {
	client := NewClient(skillsShBaseURL)

	params := map[string]string{
		"q":     query,
		"limit": fmt.Sprintf("%d", limit),
	}

	data, err := client.Get("/api/search", params)
	if err != nil {
		return nil, err
	}

	var response SkillsShResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	var skills []Skill
	for _, s := range response.Skills {
		skills = append(skills, Skill{
			ID:       s.ID,
			Name:     s.Name,
			Source:   s.TopSource,
			Installs: s.Installs,
			Registry: "skills.sh",
		})
	}

	return skills, nil
}

// GetSkillsShTrending gets trending skills from skills.sh
func GetSkillsShTrending(limit int) ([]Skill, error) {
	client := NewClient(skillsShBaseURL)

	params := map[string]string{
		"limit": fmt.Sprintf("%d", limit),
	}

	data, err := client.Get("/api/skills", params)
	if err != nil {
		return nil, err
	}

	var response SkillsShResponse
	if err := parseJSON(data, &response); err != nil {
		return nil, err
	}

	var skills []Skill
	for _, s := range response.Skills {
		skills = append(skills, Skill{
			ID:       s.ID,
			Name:     s.Name,
			Source:   s.TopSource,
			Installs: s.Installs,
			Registry: "skills.sh",
		})
	}

	return skills, nil
}
