// Package i18n provides internationalization support for TokMan CLI.
// 
// Languages: en, fr, zh, ja, es, de, ko
// Format: TOML files in locales/ directory
package i18n

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

//go:embed locales/*.toml
var localeFS embed.FS

// translations stores all loaded translations.
var (
	translations map[string]map[string]string
	once         sync.Once
	defaultLang  = "en"
	currentLang  = "en"
)

// Load loads all locale files.
func Load() error {
	var err error
	once.Do(func() {
		translations = make(map[string]map[string]string)
		
		// Try loading from filesystem first (for development)
		homeDir, _ := os.UserHomeDir()
		localeDir := filepath.Join(homeDir, ".local", "share", "tokman", "locales")
		
		if files, err := filepath.Glob(filepath.Join(localeDir, "*.toml")); err == nil && len(files) > 0 {
			for _, file := range files {
				lang := strings.TrimSuffix(filepath.Base(file), ".toml")
				if err := loadFile(translations, file, lang); err != nil {
					// Log error but continue
					continue
				}
			}
		}
		
		// If no external files, check embedded
		if len(translations) == 0 {
			if entries, err := localeFS.ReadDir("locales"); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}
					lang := strings.TrimSuffix(entry.Name(), ".toml")
					data, err := localeFS.ReadFile("locales/" + entry.Name())
					if err != nil {
						continue
					}
					translations[lang] = parseTOML(data)
				}
			}
		}
		
		// Ensure English is default
		if _, ok := translations["en"]; !ok {
			translations["en"] = make(map[string]string)
		}
	})
	return err
}

// SetLanguage sets the current language.
func SetLanguage(lang string) string {
	if _, ok := translations[lang]; ok {
		currentLang = lang
		return lang
	}
	// Fallback to English
	if _, ok := translations["en"]; ok {
		currentLang = "en"
		return "en"
	}
	return ""
}

// T translates a message key with optional arguments.
// Format: T("common.success", "action", "Filtering")
// Returns translated string with arguments substituted.
func T(key string, args ...string) string {
	if translations == nil {
		return key
	}
	
	// Try current language
	if msg, ok := translations[currentLang][key]; ok {
		return substitute(msg, args)
	}
	
	// Fallback to English
	if msg, ok := translations["en"][key]; ok {
		return substitute(msg, args)
	}
	
	return key
}

// GetAvailableLanguages returns list of available language codes.
func GetAvailableLanguages() []string {
	var langs []string
	for lang := range translations {
		langs = append(langs, lang)
	}
	return langs
}

// GetCurrentLanguage returns the current language code.
func GetCurrentLanguage() string {
	return currentLang
}

// GetLanguageName returns human-readable language name.
func GetLanguageName(code string) string {
	names := map[string]string{
		"en": "English",
		"fr": "Français",
		"zh": "中文",
		"ja": "日本語",
		"es": "Español",
		"de": "Deutsch",
		"ko": "한국어",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return code
}

func substitute(msg string, args []string) string {
	// Handle {action}, {error}, etc positional args
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		msg = strings.ReplaceAll(msg, placeholder, arg)
	}
	
	// Handle named args: args come in pairs (name, value)
	for i := 0; i+1 < len(args); i += 2 {
		// Skip positional args
		if _, err := fmt.Sscanf(args[i], "%d", new(int)); err == nil {
			continue
		}
		placeholder := "{" + args[i] + "}"
		msg = strings.ReplaceAll(msg, placeholder, args[i+1])
	}
	
	return msg
}

func loadFile(trans map[string]map[string]string, file, lang string) error {
	flat, err := flattenTOML(file)
	if err != nil {
		return err
	}
	trans[lang] = flat
	return nil
}

func flattenTOML(file string) (map[string]string, error) {
	var data map[string]interface{}
	if _, err := toml.DecodeFile(file, &data); err != nil {
		return nil, err
	}
	return flatten(data, ""), nil
}

func parseTOML(data []byte) map[string]string {
	var m map[string]interface{}
	if _, err := toml.Decode(string(data), &m); err != nil {
		return make(map[string]string)
	}
	return flatten(m, "")
}

func flatten(data map[string]interface{}, prefix string) map[string]string {
	result := make(map[string]string)
	for key, val := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := val.(type) {
		case string:
			result[fullKey] = v
		case map[string]interface{}:
			for k, v := range flatten(v, fullKey) {
				result[k] = v
			}
		}
	}
	return result
}
