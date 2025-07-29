package utils

import (
	"errors"
	"strings"
	"testing"
)

func TestGenerateFunName(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		validate func(result string) error
	}{
		{
			name: "generates adjective-noun format",
			base: "test",
			validate: func(result string) error {
				parts := strings.Split(result, "-")
				if len(parts) != 2 {
					return errors.New("expected format: adjective-noun")
				}
				
				// Check if adjective exists
				adjFound := false
				for _, adj := range adjectives {
					if parts[0] == adj {
						adjFound = true
						break
					}
				}
				if !adjFound {
					return errors.New("adjective not found in list")
				}
				
				// Check if noun exists
				nounFound := false
				for _, noun := range nouns {
					if parts[1] == noun {
						nounFound = true
						break
					}
				}
				if !nounFound {
					return errors.New("noun not found in list")
				}
				
				return nil
			},
		},
		{
			name: "generates different names on multiple calls",
			base: "",
			validate: func(result string) error {
				// Generate multiple names and check for diversity
				names := make(map[string]bool)
				for i := 0; i < 10; i++ {
					name := GenerateFunName("")
					names[name] = true
				}
				
				// We should get at least 2 different names in 10 tries
				if len(names) < 2 {
					return errors.New("not enough diversity in generated names")
				}
				return nil
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFunName(tt.base)
			if err := tt.validate(result); err != nil {
				t.Errorf("GenerateFunName() validation failed: %v", err)
			}
		})
	}
}

func TestGenerateCodespaceName(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		repoName string
		expected string
	}{
		{
			name:     "simple owner and repo",
			owner:    "facebook",
			repoName: "react",
			expected: "facebook-react",
		},
		{
			name:     "owner with uppercase",
			owner:    "Facebook",
			repoName: "React",
			expected: "facebook-react",
		},
		{
			name:     "names with special characters",
			owner:    "user_name",
			repoName: "my-repo.js",
			expected: "user-name-my-repo-js",
		},
		{
			name:     "names with numbers",
			owner:    "user123",
			repoName: "repo456",
			expected: "user123-repo456",
		},
		{
			name:     "empty owner",
			owner:    "",
			repoName: "repo",
			expected: "codespace-repo",
		},
		{
			name:     "empty repo",
			owner:    "owner",
			repoName: "",
			expected: "owner-codespace",
		},
		{
			name:     "both empty",
			owner:    "",
			repoName: "",
			expected: "codespace-codespace",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCodespaceName(tt.owner, tt.repoName)
			if result != tt.expected {
				t.Errorf("GenerateCodespaceName(%q, %q) = %q, expected %q", 
					tt.owner, tt.repoName, result, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueCodespaceName(t *testing.T) {
	tests := []struct {
		name         string
		owner        string
		repoName     string
		existingNames map[string]bool
		validate     func(result string) error
	}{
		{
			name:     "base name doesn't exist",
			owner:    "facebook",
			repoName: "react",
			existingNames: map[string]bool{
				"google-angular": true,
			},
			validate: func(result string) error {
				if result != "facebook-react" {
					return errors.New("should return base name when it doesn't exist")
				}
				return nil
			},
		},
		{
			name:     "base name exists, adds fun suffix",
			owner:    "facebook",
			repoName: "react",
			existingNames: map[string]bool{
				"facebook-react": true,
			},
			validate: func(result string) error {
				if !strings.HasPrefix(result, "facebook-react-") {
					return errors.New("should have base name prefix")
				}
				parts := strings.Split(result, "-")
				if len(parts) < 4 { // facebook-react-adjective-noun
					return errors.New("should have fun suffix")
				}
				return nil
			},
		},
		{
			name:     "fallback to timestamp when all attempts fail",
			owner:    "test",
			repoName: "repo",
			existingNames: map[string]bool{
				// This will match any generated name
			},
			validate: func(result string) error {
				// Override checkExists to always return true, forcing timestamp fallback
				checkExists := func(name string) bool { return true }
				result = GenerateUniqueCodespaceName("test", "repo", checkExists)
				
				if !strings.HasPrefix(result, "test-repo-") {
					return errors.New("should have base name prefix")
				}
				
				// Check if it ends with a timestamp (Unix time)
				parts := strings.Split(result, "-")
				lastPart := parts[len(parts)-1]
				
				// The timestamp should be within a reasonable range (10-11 digits for Unix timestamp)
				if len(lastPart) < 10 || len(lastPart) > 11 {
					return errors.New("fallback should use Unix timestamp")
				}
				
				return nil
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkExists := func(name string) bool {
				return tt.existingNames[name]
			}
			
			result := GenerateUniqueCodespaceName(tt.owner, tt.repoName, checkExists)
			if err := tt.validate(result); err != nil {
				t.Errorf("GenerateUniqueCodespaceName() validation failed: %v", err)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase conversion",
			input:    "MyRepo",
			expected: "myrepo",
		},
		{
			name:     "replace spaces with hyphens",
			input:    "my repo name",
			expected: "my-repo-name",
		},
		{
			name:     "replace special characters",
			input:    "my@repo#name",
			expected: "my-repo-name",
		},
		{
			name:     "preserve numbers",
			input:    "repo123",
			expected: "repo123",
		},
		{
			name:     "preserve existing hyphens",
			input:    "my-repo",
			expected: "my-repo",
		},
		{
			name:     "multiple consecutive special chars",
			input:    "my@@##repo",
			expected: "my-repo",
		},
		{
			name:     "trim hyphens from start and end",
			input:    "---repo---",
			expected: "repo",
		},
		{
			name:     "empty string returns codespace",
			input:    "",
			expected: "codespace",
		},
		{
			name:     "only special chars returns codespace",
			input:    "@#$%",
			expected: "codespace",
		},
		{
			name:     "mixed case with underscores",
			input:    "My_Repo_Name",
			expected: "my-repo-name",
		},
		{
			name:     "dots become hyphens",
			input:    "my.repo.name",
			expected: "my-repo-name",
		},
		{
			name:     "unicode characters",
			input:    "my-répo-🚀",
			expected: "my-r-po",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSecurePassword(t *testing.T) {
	t.Run("generates 16 character password", func(t *testing.T) {
		password := GenerateSecurePassword()
		if len(password) != 16 {
			t.Errorf("Expected password length 16, got %d", len(password))
		}
	})
	
	t.Run("generates different passwords", func(t *testing.T) {
		passwords := make(map[string]bool)
		for i := 0; i < 10; i++ {
			pwd := GenerateSecurePassword()
			passwords[pwd] = true
		}
		
		if len(passwords) < 10 {
			t.Errorf("Expected 10 unique passwords, got %d", len(passwords))
		}
	})
	
	t.Run("password contains alphanumeric characters", func(t *testing.T) {
		password := GenerateSecurePassword()
		hasLetter := false
		
		for _, r := range password {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				hasLetter = true
			}
		}
		
		if !hasLetter {
			t.Error("Password should contain at least one letter")
		}
		// Note: Base64 encoding might not always include numbers
	})
	
	t.Run("password characters are valid", func(t *testing.T) {
		// Test that the password only contains valid characters
		password := GenerateSecurePassword()
		
		// Base64 URL encoding uses these characters
		validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		for _, r := range password {
			if !strings.ContainsRune(validChars, r) {
				t.Errorf("Password contains invalid character: %c", r)
			}
		}
	})
}

// Benchmark tests
func BenchmarkGenerateFunName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateFunName("")
	}
}

func BenchmarkGenerateSecurePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateSecurePassword()
	}
}

func BenchmarkSanitizeName(b *testing.B) {
	testCases := []string{
		"MyRepo",
		"my@repo#name",
		"my-repo-name-with-many-parts",
		"UPPERCASE_WITH_UNDERSCORES",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeName(testCases[i%len(testCases)])
	}
}