package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
)

var (
	skillsCmd = &cobra.Command{
		Use:   "skills",
		Short: "Skills system for reusable command templates",
		Long: `Skills are reusable command templates that can be invoked with /skill-name.
		
Skills are markdown files in ~/.config/tokman/skills/ with the following format:
---
name: my-skill
description: Brief description
---
# Skill content here

Examples:
  tokman skills list         # List all skills
  tokman skills create foo   # Create new skill template
  tokman skills run foo     # Run a skill
  tokman skills edit foo    # Edit a skill`,
	}

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all installed skills",
		RunE:  runSkillsList,
	}

	createCmd = &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new skill template",
		RunE:  runSkillsCreate,
	}

	runCmd = &cobra.Command{
		Use:   "run [name]",
		Short: "Run a skill",
		RunE:  runSkillsRun,
	}

	editCmd = &cobra.Command{
		Use:   "edit [name]",
		Short: "Edit a skill",
		RunE:  runSkillsEdit,
	}

	searchCmd = &cobra.Command{
		Use:   "search [query]",
		Short: "Search skills",
		RunE:  runSkillsSearch,
	}
)

func init() {
	skillsCmd.AddCommand(listCmd)
	skillsCmd.AddCommand(createCmd)
	skillsCmd.AddCommand(runCmd)
	skillsCmd.AddCommand(editCmd)
	skillsCmd.AddCommand(searchCmd)

	registry.Add(func() {
		registry.Register(skillsCmd)
	})
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Path        string `json:"path"`
}

func getSkillsDir() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "tokman", "skills")
}

func loadSkill(name string) (*Skill, error) {
	skillsDir := getSkillsDir()
	skillPath := filepath.Join(skillsDir, name+".md")

	data, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %s", name)
	}

	content := string(data)
	desc := extractDescription(content)

	return &Skill{
		Name:        name,
		Description: desc,
		Content:     content,
		Path:        skillPath,
	}, nil
}

func extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	desc := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "description:") {
			desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			break
		}
	}
	if desc == "" && len(lines) > 3 {
		desc = strings.TrimSpace(lines[3])
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
	}
	return desc
}

func listSkills() []*Skill {
	skillsDir := getSkillsDir()
	os.MkdirAll(skillsDir, 0755)

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var skills []*Skill
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			if skill, err := loadSkill(name); err == nil {
				skills = append(skills, skill)
			}
		}
	}

	return skills
}

func runSkillsList(cmd *cobra.Command, args []string) error {
	skills := listSkills()

	if len(skills) == 0 {
		fmt.Println("No skills installed")
		fmt.Println("Create one with: tokman skills create <name>")
		return nil
	}

	fmt.Println("=== Installed Skills ===")
	for _, s := range skills {
		fmt.Printf("/%-20s %s\n", s.Name, s.Description)
	}

	return nil
}

func runSkillsCreate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("skill name required")
	}

	name := args[0]
	skillsDir := getSkillsDir()
	os.MkdirAll(skillsDir, 0755)

	skillPath := filepath.Join(skillsDir, name+".md")
	if _, err := os.Stat(skillPath); err == nil {
		return fmt.Errorf("skill already exists: %s", name)
	}

	template := fmt.Sprintf(`---
name: %s
description: Brief description of what this skill does
version: 1.0.0
---

# %s Skill

Describe what this skill does and how the AI should use it.

## Usage

Explain when and how to use this skill.

## Examples

Provide examples of how to use this skill.
`, name, name)

	if err := os.WriteFile(skillPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create skill: %w", err)
	}

	fmt.Printf("Created skill: %s at %s\n", name, skillPath)
	fmt.Printf("Edit with: tokman skills edit %s\n", name)

	return nil
}

func runSkillsRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("skill name required")
	}

	skill, err := loadSkill(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Running skill: %s\n", skill.Name)
	fmt.Printf("Description: %s\n", skill.Description)
	fmt.Println("\n--- Skill Content ---")
	fmt.Println(skill.Content)

	return nil
}

func runSkillsEdit(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("skill name required")
	}

	skill, err := loadSkill(args[0])
	if err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmdExec := exec.Command(editor, skill.Path)
	cmdExec.Stdout = os.Stdout
	cmdExec.Stdin = os.Stdin
	cmdExec.Stderr = os.Stderr

	return cmdExec.Run()
}

func runSkillsSearch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search query required")
	}

	query := strings.ToLower(args[0])
	skills := listSkills()

	var matched []*Skill
	for _, s := range skills {
		if strings.Contains(strings.ToLower(s.Name), query) ||
			strings.Contains(strings.ToLower(s.Description), query) {
			matched = append(matched, s)
		}
	}

	if len(matched) == 0 {
		fmt.Printf("No skills found matching: %s\n", query)
		return nil
	}

	fmt.Println("=== Search Results ===")
	for _, s := range matched {
		fmt.Printf("/%-20s %s\n", s.Name, s.Description)
	}

	return nil
}
