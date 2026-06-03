package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installSkillCmd = &cobra.Command{
	Use:   "install-skill --target <assistant>",
	Short: "Install technocore skill for an AI assistant",
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target == "" {
			fmt.Println("Error: --target required (claude, opencode, cursor, codex)")
			os.Exit(1)
		}

		var skillDir, destDir string
		switch target {
		case "claude":
			skillDir = "skills/claude"
			base := os.Getenv("CLAUDE_CONFIG_DIR")
			if base == "" {
				base = filepath.Join(os.Getenv("HOME"), ".claude")
			}
			destDir = filepath.Join(base, "skills", "technocore")
			installClaudeHook(base, skillDir)
		case "opencode":
			skillDir = "skills/opencode"
			base := filepath.Join(os.Getenv("HOME"), ".opencode")
			destDir = filepath.Join(base, "skills", "technocore")
		case "cursor":
			skillDir = "skills/cursor"
			base := filepath.Join(os.Getenv("HOME"), ".cursor")
			destDir = filepath.Join(base, "skills", "technocore")
		case "codex":
			skillDir = "skills/codex"
			base := filepath.Join(os.Getenv("HOME"), ".codex")
			destDir = filepath.Join(base, "skills", "technocore")
			installCodexHook(base, skillDir)
		default:
			fmt.Printf("Unsupported target: %s\n", target)
			os.Exit(1)
		}

		// Find source skill files
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		srcDir := filepath.Join(exeDir, "..", skillDir)
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			// Fallback to CWD for dev builds
			srcDir = skillDir
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			fmt.Println("Error creating skill directory:", err)
			os.Exit(1)
		}

		entries, err := os.ReadDir(srcDir)
		if err != nil {
			fmt.Println("Error reading skill source:", err)
			os.Exit(1)
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
			if err != nil {
				fmt.Printf("Warning: could not read %s: %v\n", e.Name(), err)
				continue
			}
			if err := os.WriteFile(filepath.Join(destDir, e.Name()), data, 0644); err != nil {
				fmt.Printf("Warning: could not write %s: %v\n", e.Name(), err)
				continue
			}
			fmt.Printf("Installed %s\n", e.Name())
		}

		fmt.Printf("Technocore skill installed for %s at %s\n", target, destDir)
	},
}

func installClaudeHook(baseDir, skillDir string) {
	hooksDir := filepath.Join(baseDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Println("Warning: could not create hooks dir:", err)
		return
	}

	// Find source hook.sh
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	srcHook := filepath.Join(exeDir, "..", skillDir, "hook.sh")
	if _, err := os.Stat(srcHook); os.IsNotExist(err) {
		srcHook = filepath.Join(skillDir, "hook.sh")
	}

	data, err := os.ReadFile(srcHook)
	if err != nil {
		fmt.Println("Warning: could not read hook.sh:", err)
		return
	}

	hookPath := filepath.Join(hooksDir, "technocore-hook.sh")
	if err := os.WriteFile(hookPath, data, 0755); err != nil {
		fmt.Println("Warning: could not write hook:", err)
		return
	}
	fmt.Println("Installed hook.sh to", hookPath)

	// Update settings.json
	settingsPath := filepath.Join(baseDir, "settings.json")
	settingsData, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Println("Warning: could not read settings.json:", err)
		return
	}

	// Backup
	backupPath := settingsPath + ".bak.technocore"
	os.WriteFile(backupPath, settingsData, 0644)

	var settings map[string]interface{}
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		fmt.Println("Warning: could not parse settings.json:", err)
		return
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
		settings["hooks"] = hooks
	}

	userPromptHooks, ok := hooks["UserPromptSubmit"].([]interface{})
	if !ok {
		userPromptHooks = []interface{}{}
	}

	// Check if technocore hook already exists
	for _, h := range userPromptHooks {
		entry, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		innerHooks, ok := entry["hooks"].([]interface{})
		if !ok {
			continue
		}
		for _, ih := range innerHooks {
			inner, ok := ih.(map[string]interface{})
			if !ok {
				continue
			}
			cmd, _ := inner["command"].(string)
			if contains(cmd, "technocore") {
				fmt.Println("Technocore UserPromptSubmit hook already registered")
				return
			}
		}
	}

	// Add technocore hook
	newHook := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{
				"command": hookPath,
				"timeout": 10,
				"type":    "command",
			},
		},
	}
	userPromptHooks = append(userPromptHooks, newHook)
	hooks["UserPromptSubmit"] = userPromptHooks

	updated, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		fmt.Println("Warning: could not marshal settings:", err)
		return
	}
	if err := os.WriteFile(settingsPath, updated, 0644); err != nil {
		fmt.Println("Warning: could not write settings.json:", err)
		return
	}
	fmt.Println("Registered UserPromptSubmit hook in settings.json")
	fmt.Println("Backup saved to", backupPath)
}

func installCodexHook(baseDir, skillDir string) {
	hooksDir := filepath.Join(baseDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Println("Warning: could not create hooks dir:", err)
		return
	}

	// Find source hook.sh
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	srcHook := filepath.Join(exeDir, "..", skillDir, "hook.sh")
	if _, err := os.Stat(srcHook); os.IsNotExist(err) {
		srcHook = filepath.Join(skillDir, "hook.sh")
	}

	data, err := os.ReadFile(srcHook)
	if err != nil {
		fmt.Println("Warning: could not read hook.sh:", err)
		return
	}

	hookPath := filepath.Join(hooksDir, "technocore-hook.sh")
	if err := os.WriteFile(hookPath, data, 0755); err != nil {
		fmt.Println("Warning: could not write hook:", err)
		return
	}
	fmt.Println("Installed hook.sh to", hookPath)

	// Update hooks.json
	hooksJSONPath := filepath.Join(baseDir, "hooks.json")
	hooksJSONData, err := os.ReadFile(hooksJSONPath)
	if err != nil {
		// hooks.json may not exist yet; create minimal one
		hooksData := map[string]interface{}{
			"UserPromptSubmit": []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": hookPath,
							"timeout": 10,
						},
					},
				},
			},
		}
		updated, _ := json.MarshalIndent(hooksData, "", "  ")
		os.WriteFile(hooksJSONPath, updated, 0644)
		fmt.Println("Created hooks.json with UserPromptSubmit hook")
		return
	}

	var hooksData map[string]interface{}
	if err := json.Unmarshal(hooksJSONData, &hooksData); err != nil {
		fmt.Println("Warning: could not parse hooks.json:", err)
		return
	}

	userPromptHooks, ok := hooksData["UserPromptSubmit"].([]interface{})
	if !ok {
		userPromptHooks = []interface{}{}
	}

	// Check if technocore hook already exists
	for _, h := range userPromptHooks {
		entry, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		innerHooks, ok := entry["hooks"].([]interface{})
		if !ok {
			continue
		}
		for _, ih := range innerHooks {
			inner, ok := ih.(map[string]interface{})
			if !ok {
				continue
			}
			cmd, _ := inner["command"].(string)
			if contains(cmd, "technocore") {
				fmt.Println("Technocore UserPromptSubmit hook already registered")
				return
			}
		}
	}

	newHook := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": hookPath,
				"timeout": 10,
			},
		},
	}
	userPromptHooks = append(userPromptHooks, newHook)
	hooksData["UserPromptSubmit"] = userPromptHooks

	updated, err := json.MarshalIndent(hooksData, "", "  ")
	if err != nil {
		fmt.Println("Warning: could not marshal hooks.json:", err)
		return
	}
	if err := os.WriteFile(hooksJSONPath, updated, 0644); err != nil {
		fmt.Println("Warning: could not write hooks.json:", err)
		return
	}
	fmt.Println("Registered UserPromptSubmit hook in hooks.json")
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (filepath.Base(s) == substr || filepath.Base(filepath.Dir(s)) == substr))
}

func init() {
	rootCmd.AddCommand(installSkillCmd)
	installSkillCmd.Flags().StringP("target", "t", "", "Assistant target (claude, opencode, cursor, codex)")
}
