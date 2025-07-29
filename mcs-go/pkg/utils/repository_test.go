package utils

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseRepository(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Repository
		wantErr bool
	}{
		// Local paths
		{
			name:  "current directory",
			input: ".",
			want: &Repository{
				URL:     mustAbs("."),
				Name:    filepath.Base(mustAbs(".")),
				IsLocal: true,
			},
			wantErr: false,
		},
		{
			name:  "relative path",
			input: "./my-project",
			want: &Repository{
				URL:     mustAbs("./my-project"),
				Name:    "my-project",
				IsLocal: true,
			},
			wantErr: false,
		},
		{
			name:  "absolute path",
			input: "/home/user/projects/my-app",
			want: &Repository{
				URL:     "/home/user/projects/my-app",
				Name:    "my-app",
				IsLocal: true,
			},
			wantErr: false,
		},
		
		// Short format (owner/repo)
		{
			name:  "github short format",
			input: "facebook/react",
			want: &Repository{
				URL:   "https://github.com/facebook/react",
				Host:  "github.com",
				Owner: "facebook",
				Name:  "react",
			},
			wantErr: false,
		},
		{
			name:  "github short format with .git",
			input: "facebook/react.git",
			want: &Repository{
				URL:   "https://github.com/facebook/react.git",
				Host:  "github.com",
				Owner: "facebook",
				Name:  "react",
			},
			wantErr: false,
		},
		
		// HTTPS URLs
		{
			name:  "https github url",
			input: "https://github.com/golang/go",
			want: &Repository{
				URL:   "https://github.com/golang/go",
				Host:  "github.com",
				Owner: "golang",
				Name:  "go",
			},
			wantErr: false,
		},
		{
			name:  "https github url with .git",
			input: "https://github.com/golang/go.git",
			want: &Repository{
				URL:   "https://github.com/golang/go.git",
				Host:  "github.com",
				Owner: "golang",
				Name:  "go",
			},
			wantErr: false,
		},
		{
			name:  "https gitlab url",
			input: "https://gitlab.com/company/project.git",
			want: &Repository{
				URL:   "https://gitlab.com/company/project.git",
				Host:  "gitlab.com",
				Owner: "company",
				Name:  "project",
			},
			wantErr: false,
		},
		{
			name:  "https bitbucket url",
			input: "https://bitbucket.org/user/repo",
			want: &Repository{
				URL:   "https://bitbucket.org/user/repo",
				Host:  "bitbucket.org",
				Owner: "user",
				Name:  "repo",
			},
			wantErr: false,
		},
		
		// Git SSH URLs
		{
			name:  "git ssh github",
			input: "git@github.com:michaelkeevildown/michaels-codespaces.git",
			want: &Repository{
				URL:   "git@github.com:michaelkeevildown/michaels-codespaces.git",
				Host:  "github.com",
				Owner: "michaelkeevildown",
				Name:  "michaels-codespaces",
			},
			wantErr: false,
		},
		{
			name:  "git ssh gitlab",
			input: "git@gitlab.com:group/project.git",
			want: &Repository{
				URL:   "git@gitlab.com:group/project.git",
				Host:  "gitlab.com",
				Owner: "group",
				Name:  "project",
			},
			wantErr: false,
		},
		{
			name:  "git ssh custom host",
			input: "git@git.company.com:team/repository.git",
			want: &Repository{
				URL:   "git@git.company.com:team/repository.git",
				Host:  "git.company.com",
				Owner: "team",
				Name:  "repository",
			},
			wantErr: false,
		},
		
		// Nested paths
		{
			name:  "nested group path",
			input: "https://gitlab.com/group/subgroup/project",
			want: &Repository{
				URL:   "https://gitlab.com/group/subgroup/project",
				Host:  "gitlab.com",
				Owner: "group",
				Name:  "subgroup",
			},
			wantErr: false,
		},
		
		// Edge cases
		{
			name:  "url with port",
			input: "https://git.company.com:8080/team/repo.git",
			want: &Repository{
				URL:   "https://git.company.com:8080/team/repo.git",
				Host:  "git.company.com:8080",
				Owner: "team",
				Name:  "repo",
			},
			wantErr: false,
		},
		{
			name:  "url with auth",
			input: "https://user:pass@github.com/owner/repo",
			want: &Repository{
				URL:   "https://user:pass@github.com/owner/repo",
				Host:  "github.com",
				Owner: "owner",
				Name:  "repo",
			},
			wantErr: false,
		},
		
		// Error cases
		{
			name:    "invalid url",
			input:   "not a valid url at all",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only host",
			input:   "https://github.com",
			wantErr: true,
		},
		{
			name:    "only owner",
			input:   "https://github.com/owner",
			wantErr: true,
		},
		{
			name:    "malformed git ssh without colon",
			input:   "git@github.com/owner/repo",
			want: &Repository{
				URL:   "git@github.com/owner/repo",
				Host:  "github.com",
				Owner: "owner",
				Name:  "repo",
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepository(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr {
				return
			}
			
			// For local paths, we need to resolve the absolute path
			if tt.want.IsLocal {
				absPath, _ := filepath.Abs(tt.input)
				tt.want.URL = absPath
				if tt.input == "." {
					tt.want.Name = filepath.Base(absPath)
				}
			}
			
			if !repositoryEqual(got, tt.want) {
				t.Errorf("ParseRepository() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseRepository_GitSSHPreservation(t *testing.T) {
	// Special test to ensure git@ URLs are preserved correctly
	gitURL := "git@github.com:owner/repo.git"
	repo, err := ParseRepository(gitURL)
	
	if err != nil {
		t.Fatalf("ParseRepository() error = %v", err)
	}
	
	if repo.URL != gitURL {
		t.Errorf("Git SSH URL not preserved. Got %s, want %s", repo.URL, gitURL)
	}
	
	if repo.Host != "github.com" {
		t.Errorf("Host = %s, want github.com", repo.Host)
	}
	
	if repo.Owner != "owner" {
		t.Errorf("Owner = %s, want owner", repo.Owner)
	}
	
	if repo.Name != "repo" {
		t.Errorf("Name = %s, want repo", repo.Name)
	}
}

func TestParseRepository_ComplexPaths(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*Repository) error
	}{
		{
			name:  "deeply nested gitlab path",
			input: "https://gitlab.com/group/subgroup/subsubgroup/project.git",
			check: func(r *Repository) error {
				if r.Owner != "group" {
					return errors.New("should extract first path component as owner")
				}
				if r.Name != "subgroup" {
					return errors.New("should extract second path component as name")
				}
				return nil
			},
		},
		{
			name:  "url with query params",
			input: "https://github.com/owner/repo?tab=readme",
			check: func(r *Repository) error {
				if !strings.HasSuffix(r.URL, "?tab=readme") {
					return errors.New("should preserve query parameters")
				}
				return nil
			},
		},
		{
			name:  "url with fragment",
			input: "https://github.com/owner/repo#installation",
			check: func(r *Repository) error {
				if !strings.HasSuffix(r.URL, "#installation") {
					return errors.New("should preserve fragment")
				}
				return nil
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := ParseRepository(tt.input)
			if err != nil {
				t.Fatalf("ParseRepository() error = %v", err)
			}
			
			if err := tt.check(repo); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

// Helper functions
func repositoryEqual(a, b *Repository) bool {
	if a == nil || b == nil {
		return a == b
	}
	
	return a.URL == b.URL &&
		a.Host == b.Host &&
		a.Owner == b.Owner &&
		a.Name == b.Name &&
		a.IsLocal == b.IsLocal
}

func mustAbs(path string) string {
	abs, _ := filepath.Abs(path)
	return abs
}

// Benchmark tests
func BenchmarkParseRepository(b *testing.B) {
	inputs := []string{
		".",
		"facebook/react",
		"https://github.com/golang/go.git",
		"git@github.com:kubernetes/kubernetes.git",
		"https://gitlab.com/group/subgroup/project",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseRepository(inputs[i%len(inputs)])
	}
}

// Table-driven tests for edge cases
func TestParseRepository_EdgeCases(t *testing.T) {
	edgeCases := []struct {
		name        string
		input       string
		shouldError bool
		validate    func(*Repository) error
	}{
		{
			name:        "whitespace only",
			input:       "   ",
			shouldError: true,
		},
		{
			name:        "newline in input",
			input:       "github.com/owner/repo\n",
			shouldError: true,
		},
		{
			name:  "trailing slash in https",
			input: "https://github.com/owner/repo/",
			validate: func(r *Repository) error {
				if r.Name != "repo" {
					return errors.New("should handle trailing slash")
				}
				return nil
			},
		},
		{
			name:  "multiple .git extensions",
			input: "https://github.com/owner/repo.git.git",
			validate: func(r *Repository) error {
				if r.Name != "repo.git" {
					return errors.New("should only strip last .git")
				}
				return nil
			},
		},
		{
			name:  "repo name with dots",
			input: "https://github.com/owner/my.awesome.repo",
			validate: func(r *Repository) error {
				if r.Name != "my.awesome.repo" {
					return errors.New("should preserve dots in repo name")
				}
				return nil
			},
		},
		{
			name:  "repo name with hyphens and underscores",
			input: "git@github.com:user/my-awesome_repo.git",
			validate: func(r *Repository) error {
				if r.Name != "my-awesome_repo" {
					return errors.New("should preserve hyphens and underscores")
				}
				return nil
			},
		},
	}
	
	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := ParseRepository(tc.input)
			
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if tc.validate != nil {
				if err := tc.validate(repo); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

// Test Repository struct methods if any are added in the future
func TestRepository_String(t *testing.T) {
	// This test is a placeholder for future Repository methods
	repo := &Repository{
		URL:   "https://github.com/owner/repo",
		Host:  "github.com",
		Owner: "owner",
		Name:  "repo",
	}
	
	// Basic validation that the struct can be used
	if repo.URL == "" {
		t.Error("Repository URL should not be empty")
	}
}

// Fuzz test for robustness
func TestParseRepository_Fuzz(t *testing.T) {
	// Test with various malformed inputs to ensure no panics
	fuzzInputs := []string{
		"",
		" ",
		"///",
		":::",
		"git@",
		"https://",
		"ftp://invalid.com/repo",
		"git@github.com:",
		"https://github.com//double//slash",
		"git@github.com:owner/",
		"https://[::1]/repo",
		"https://github.com/" + strings.Repeat("a", 1000),
		"\x00\x01\x02",
		"https://github.com/\n/repo",
		"git@github.com:owner\x00/repo",
	}
	
	for _, input := range fuzzInputs {
		t.Run("fuzz_"+input, func(t *testing.T) {
			// We're not checking the result, just ensuring no panic
			_, _ = ParseRepository(input)
		})
	}
}

// Test type safety
func TestRepository_TypeSafety(t *testing.T) {
	// Ensure Repository fields have expected types
	var r Repository
	
	// This will fail to compile if types change
	var _ string = r.URL
	var _ string = r.Host
	var _ string = r.Owner
	var _ string = r.Name
	var _ bool = r.IsLocal
	
	// Test zero value behavior
	if r.URL != "" || r.Host != "" || r.Owner != "" || r.Name != "" || r.IsLocal != false {
		t.Error("Zero value Repository should have empty strings and false bool")
	}
}

// Integration test for common workflows
func TestParseRepository_CommonWorkflows(t *testing.T) {
	workflows := []struct {
		name     string
		scenario string
		input    string
		usage    func(*Repository) error
	}{
		{
			name:  "clone from github",
			input: "git@github.com:golang/go.git",
			usage: func(r *Repository) error {
				// Simulate using the URL for git clone
				if r.URL != "git@github.com:golang/go.git" {
					return errors.New("URL should be preserved for cloning")
				}
				// Generate container name
				containerName := r.Owner + "-" + r.Name
				if containerName != "golang-go" {
					return errors.New("container name generation failed")
				}
				return nil
			},
		},
		{
			name:  "local development",
			input: ".",
			usage: func(r *Repository) error {
				if !r.IsLocal {
					return errors.New("should identify as local repository")
				}
				if r.URL == "" {
					return errors.New("should have absolute path")
				}
				return nil
			},
		},
		{
			name:  "parse and display",
			input: "microsoft/vscode",
			usage: func(r *Repository) error {
				display := r.Owner + "/" + r.Name
				if display != "microsoft/vscode" {
					return errors.New("display format incorrect")
				}
				return nil
			},
		},
	}
	
	for _, wf := range workflows {
		t.Run(wf.name, func(t *testing.T) {
			repo, err := ParseRepository(wf.input)
			if err != nil {
				t.Fatalf("ParseRepository() error = %v", err)
			}
			
			if err := wf.usage(repo); err != nil {
				t.Errorf("Workflow validation failed: %v", err)
			}
		})
	}
}

// Test nil pointer safety
func TestParseRepository_NilSafety(t *testing.T) {
	// Ensure the function returns proper error for nil-like inputs
	_, err := ParseRepository("")
	if err == nil {
		t.Error("Expected error for empty string input")
	}
	
	// Ensure returned Repository is never nil on success
	repo, err := ParseRepository("owner/repo")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if repo == nil {
		t.Error("Repository should not be nil on success")
	}
	
	// Test Repository can be safely compared
	var nilRepo *Repository
	if reflect.DeepEqual(repo, nilRepo) {
		t.Error("Valid repository should not equal nil")
	}
}