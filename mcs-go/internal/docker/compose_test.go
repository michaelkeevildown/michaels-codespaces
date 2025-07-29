package docker

import (
	"strings"
	"testing"

	"github.com/michaelkeevildown/mcs/internal/components"
	"github.com/stretchr/testify/assert"
)

// TestGenerateDockerCompose tests the GenerateDockerCompose function
func TestGenerateDockerCompose(t *testing.T) {
	tests := []struct {
		name           string
		config         ComposeConfig
		expectedSubstr []string
		notExpected    []string
	}{
		{
			name: "Basic configuration",
			config: ComposeConfig{
				ContainerName: "mcs-test-project",
				CodespaceName: "test-project",
				Image:         "mcs/code-server:latest",
				Password:      "secret123",
				Ports: map[string]string{
					"8443": "8080",
					"3000": "3000",
				},
				Environment: map[string]string{
					"NODE_ENV": "development",
				},
				Labels: map[string]string{
					"custom.label": "value",
				},
			},
			expectedSubstr: []string{
				"services:",
				"mcs-test-project:",
				"image: mcs/code-server:latest",
				"container_name: mcs-test-project",
				"restart: unless-stopped",
				"PASSWORD=secret123",
				"NODE_ENV=development",
				"\"8443:8080\"",
				"\"3000:3000\"",
				"./src:/home/coder/test-project",
				"./data:/home/coder/.local/share/code-server",
				"./config:/home/coder/.config",
				"./logs:/home/coder/logs",
				"custom.label: \"value\"",
				"mcs.codespace: \"test-project\"",
				"mcs.managed: \"true\"",
				"working_dir: /home/coder/test-project",
				"code-server --bind-addr 0.0.0.0:8080 --auth password /home/coder/test-project",
				"networks:",
				"mcs-network",
				"external: true",
			},
		},
		{
			name: "Configuration with build context",
			config: ComposeConfig{
				ContainerName: "mcs-build-project",
				CodespaceName: "build-project",
				BuildContext:  "./dockerfiles",
				Dockerfile:    "Dockerfile.node",
				Password:      "buildpass",
				Ports: map[string]string{
					"8080": "8080",
				},
			},
			expectedSubstr: []string{
				"build:",
				"context: ./dockerfiles",
				"dockerfile: Dockerfile.node",
				"mcs-build-project:",
				"container_name: mcs-build-project",
				"PASSWORD=buildpass",
				"working_dir: /home/coder/build-project",
			},
		},
		{
			name: "Configuration with components",
			config: ComposeConfig{
				ContainerName: "mcs-component-project",
				CodespaceName: "component-project",
				Image:         "mcs/code-server:latest",
				Password:      "comppass",
				Ports: map[string]string{
					"8080": "8080",
				},
				Components: []components.Component{
					{Name: "Node.js", Selected: true},
					{Name: "Python", Selected: true},
				},
			},
			expectedSubstr: []string{
				"./components:/home/coder/.components:ro",
				"./init:/docker-entrypoint-initdb.d:ro",
				"if [ -f /docker-entrypoint-initdb.d/init.sh ]; then",
				"echo 'Installing components...'",
				"/docker-entrypoint-initdb.d/init.sh",
			},
		},
		{
			name: "Empty configuration with defaults",
			config: ComposeConfig{
				ContainerName: "mcs-empty",
				CodespaceName: "empty",
				Password:      "empty123",
			},
			expectedSubstr: []string{
				"image: codercom/code-server:latest", // Default image
				"mcs.codespace: \"empty\"",
				"mcs.managed: \"true\"",
			},
		},
		{
			name: "Configuration without ports",
			config: ComposeConfig{
				ContainerName: "mcs-no-ports",
				CodespaceName: "no-ports",
				Image:         "custom:latest",
				Password:      "noports",
			},
			expectedSubstr: []string{
				"image: custom:latest",
				"container_name: mcs-no-ports",
				"ports:", // ports: will always appear even if empty
			},
		},
		{
			name: "Configuration with multiple environment variables",
			config: ComposeConfig{
				ContainerName: "mcs-env-test",
				CodespaceName: "env-test",
				Image:         "test:latest",
				Password:      "envpass",
				Environment: map[string]string{
					"DEBUG":      "true",
					"API_URL":    "https://api.example.com",
					"MAX_MEMORY": "2G",
				},
			},
			expectedSubstr: []string{
				"DEBUG=true",
				"API_URL=https://api.example.com",
				"MAX_MEMORY=2G",
			},
		},
		{
			name: "Configuration with custom labels",
			config: ComposeConfig{
				ContainerName: "mcs-labels-test",
				CodespaceName: "labels-test",
				Image:         "test:latest",
				Password:      "labelpass",
				Labels: map[string]string{
					"version":     "1.0.0",
					"maintainer":  "test@example.com",
					"environment": "development",
				},
			},
			expectedSubstr: []string{
				"version: \"1.0.0\"",
				"maintainer: \"test@example.com\"",
				"environment: \"development\"",
				"mcs.codespace: \"labels-test\"",
				"mcs.managed: \"true\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateDockerCompose(tt.config)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)

			resultStr := string(result)

			// Check expected substrings
			for _, expected := range tt.expectedSubstr {
				assert.Contains(t, resultStr, expected, "Expected substring not found: %s", expected)
			}

			// Check not expected substrings
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, resultStr, notExpected, "Unexpected substring found: %s", notExpected)
			}

			// Verify it's valid YAML-like structure
			lines := strings.Split(resultStr, "\n")
			assert.True(t, len(lines) > 10, "Generated compose should have multiple lines")

			// Check for proper indentation (basic YAML structure)
			hasServices := false
			for _, line := range lines {
				if strings.TrimSpace(line) == "services:" {
					hasServices = true
					break
				}
			}
			assert.True(t, hasServices, "Generated compose should have services section")
		})
	}
}

// TestGenerateDockerCompose_TemplateError tests template execution errors
func TestGenerateDockerCompose_TemplateError(t *testing.T) {
	// Test with a valid config to ensure no template errors occur
	config := ComposeConfig{
		ContainerName: "test",
		CodespaceName: "test",
		Password:      "pass",
	}

	result, err := GenerateDockerCompose(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

// TestGenerateInitScript tests the GenerateInitScript function
func TestGenerateInitScript(t *testing.T) {
	tests := []struct {
		name           string
		components     []components.Component
		expectedSubstr []string
		notExpected    []string
	}{
		{
			name: "Multiple selected components",
			components: []components.Component{
				{
					Name:      "Node.js",
					Selected:  true,
					Installer: "install-node.sh",
				},
				{
					Name:      "Python",
					Selected:  true,
					Installer: "install-python.sh",
				},
				{
					Name:      "Docker",
					Selected:  false,
					Installer: "install-docker.sh",
				},
			},
			expectedSubstr: []string{
				"#!/bin/bash",
				"# Component installation script",
				"set -e",
				"🚀 Installing MCS components...",
				"echo \"📦 Installing Node.js...\"",
				"/home/coder/.components/install-node.sh install",
				"echo \"✅ Node.js installed successfully\"",
				"echo \"📦 Installing Python...\"",
				"/home/coder/.components/install-python.sh install",
				"echo \"✅ Python installed successfully\"",
				"✨ Component installation complete!",
				"mkdir -p /home/coder/.local/bin /home/coder/.local/share",
				"mkdir -p /home/coder/.npm-global/bin",
				"mkdir -p /home/coder/.mcs/components",
				"export PATH=\"/home/coder/.npm-global/bin:/home/coder/.local/bin:$PATH\"",
				"export NPM_PREFIX=\"/home/coder/.npm-global\"",
			},
			notExpected: []string{
				"Installing Docker", // Should not be included since not selected
				"install-docker.sh",
			},
		},
		{
			name: "No selected components",
			components: []components.Component{
				{
					Name:      "Node.js",
					Selected:  false,
					Installer: "install-node.sh",
				},
				{
					Name:      "Python",
					Selected:  false,
					Installer: "install-python.sh",
				},
			},
			expectedSubstr: []string{
				"#!/bin/bash",
				"🚀 Installing MCS components...",
				"✨ Component installation complete!",
			},
			notExpected: []string{
				"Installing Node.js",
				"Installing Python",
				"install-node.sh",
				"install-python.sh",
			},
		},
		{
			name:       "Empty components list",
			components: []components.Component{},
			expectedSubstr: []string{
				"#!/bin/bash",
				"# Component installation script",
				"set -e",
				"🚀 Installing MCS components...",
				"✨ Component installation complete!",
			},
		},
		{
			name: "Single selected component",
			components: []components.Component{
				{
					Name:      "Go",
					Selected:  true,
					Installer: "install-go.sh",
				},
			},
			expectedSubstr: []string{
				"echo \"📦 Installing Go...\"",
				"/home/coder/.components/install-go.sh install",
				"echo \"✅ Go installed successfully\"",
			},
		},
		{
			name: "Component with complex installer path",
			components: []components.Component{
				{
					Name:      "Custom Tool",
					Selected:  true,
					Installer: "tools/custom/install-custom-tool.sh",
				},
			},
			expectedSubstr: []string{
				"echo \"📦 Installing Custom Tool...\"",
				"/home/coder/.components/tools/custom/install-custom-tool.sh install",
				"echo \"✅ Custom Tool installed successfully\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateInitScript(tt.components)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)

			resultStr := string(result)

			// Check expected substrings
			for _, expected := range tt.expectedSubstr {
				assert.Contains(t, resultStr, expected, "Expected substring not found: %s", expected)
			}

			// Check not expected substrings
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, resultStr, notExpected, "Unexpected substring found: %s", notExpected)
			}

			// Verify it starts with shebang
			lines := strings.Split(resultStr, "\n")
			assert.True(t, len(lines) > 0, "Script should have at least one line")
			assert.Equal(t, "#!/bin/bash", lines[0], "Script should start with bash shebang")

			// Verify it has proper structure
			hasSetE := false
			for _, line := range lines {
				if strings.Contains(line, "set -e") {
					hasSetE = true
					break
				}
			}
			assert.True(t, hasSetE, "Script should have 'set -e' for error handling")
		})
	}
}

// TestGenerateInitScript_TemplateError tests template execution errors
func TestGenerateInitScript_TemplateError(t *testing.T) {
	// Test with valid components to ensure no template errors
	components := []components.Component{
		{Name: "Test", Selected: true, Installer: "test.sh"},
	}

	result, err := GenerateInitScript(components)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

// TestGetImageInfo tests the GetImageInfo function
func TestGetImageInfo(t *testing.T) {
	tests := []struct {
		name           string
		language       string
		components     []components.Component
		expectedImage  string
		expectedDocker string
		expectedFallback string
	}{
		{
			name:     "Python without Node.js requirements",
			language: "python",
			components: []components.Component{
				{Name: "Python Lint", Selected: true, Requires: []string{"python"}},
			},
			expectedImage:    "mcs/code-server-python:latest",
			expectedDocker:   "Dockerfile.python",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Python with Node.js requirements",
			language: "python",
			components: []components.Component{
				{Name: "ESLint", Selected: true, Requires: []string{"nodejs"}},
			},
			expectedImage:    "mcs/code-server-python-node:latest",
			expectedDocker:   "Dockerfile.python-node",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Node.js language",
			language: "node",
			components: []components.Component{
				{Name: "TypeScript", Selected: true, Requires: []string{"nodejs"}},
			},
			expectedImage:    "mcs/code-server-node:latest",
			expectedDocker:   "Dockerfile.node",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Node.js without additional requirements",
			language: "node",
			components: []components.Component{
				{Name: "Basic Tool", Selected: true, Requires: []string{}},
			},
			expectedImage:    "mcs/code-server-node:latest",
			expectedDocker:   "Dockerfile.node",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Go without Node.js requirements",
			language: "go",
			components: []components.Component{
				{Name: "Go Tools", Selected: true, Requires: []string{"go"}},
			},
			expectedImage:    "mcs/code-server-go:latest",
			expectedDocker:   "Dockerfile.go",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Go with Node.js requirements",
			language: "go",
			components: []components.Component{
				{Name: "Webpack", Selected: true, Requires: []string{"nodejs"}},
			},
			expectedImage:    "mcs/code-server-go-node:latest",
			expectedDocker:   "Dockerfile.go-node",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Generic language without Node.js",
			language: "generic",
			components: []components.Component{
				{Name: "Basic Tool", Selected: true, Requires: []string{"shell"}},
			},
			expectedImage:    "mcs/code-server-base:latest",
			expectedDocker:   "Dockerfile.base",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Generic language with Node.js",
			language: "generic",
			components: []components.Component{
				{Name: "Prettier", Selected: true, Requires: []string{"nodejs"}},
			},
			expectedImage:    "mcs/code-server-node:latest",
			expectedDocker:   "Dockerfile.node",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Unsupported language without Node.js",
			language: "rust",
			components: []components.Component{
				{Name: "Cargo", Selected: true, Requires: []string{"rust"}},
			},
			expectedImage:    "mcs/code-server-base:latest",
			expectedDocker:   "Dockerfile.rust",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Unsupported language with Node.js",
			language: "rust",
			components: []components.Component{
				{Name: "Web Tools", Selected: true, Requires: []string{"nodejs"}},
			},
			expectedImage:    "mcs/code-server-node:latest",
			expectedDocker:   "Dockerfile.rust-node",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:             "No components",
			language:         "python",
			components:       []components.Component{},
			expectedImage:    "mcs/code-server-python:latest",
			expectedDocker:   "Dockerfile.python",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Components not selected",
			language: "python",
			components: []components.Component{
				{Name: "ESLint", Selected: false, Requires: []string{"nodejs"}},
				{Name: "Python Tools", Selected: false, Requires: []string{"python"}},
			},
			expectedImage:    "mcs/code-server-python:latest",
			expectedDocker:   "Dockerfile.python",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Mixed selected and unselected components",
			language: "python",
			components: []components.Component{
				{Name: "ESLint", Selected: true, Requires: []string{"nodejs"}},
				{Name: "Docker", Selected: false, Requires: []string{"docker"}},
				{Name: "Python Tools", Selected: true, Requires: []string{"python"}},
			},
			expectedImage:    "mcs/code-server-python-node:latest",
			expectedDocker:   "Dockerfile.python-node",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Case insensitive language",
			language: "PYTHON",
			components: []components.Component{
				{Name: "Basic", Selected: true, Requires: []string{}},
			},
			expectedImage:    "mcs/code-server-python:latest",
			expectedDocker:   "Dockerfile.python",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Java language (fallback to base)",
			language: "java",
			components: []components.Component{
				{Name: "Maven", Selected: true, Requires: []string{"java"}},
			},
			expectedImage:    "mcs/code-server-base:latest",
			expectedDocker:   "Dockerfile.java",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "PHP language (fallback to base)",
			language: "php",
			components: []components.Component{
				{Name: "Composer", Selected: true, Requires: []string{"php"}},
			},
			expectedImage:    "mcs/code-server-base:latest",
			expectedDocker:   "Dockerfile.php",
			expectedFallback: "codercom/code-server:latest",
		},
		{
			name:     "Ruby language (fallback to base)",
			language: "ruby",
			components: []components.Component{
				{Name: "Bundler", Selected: true, Requires: []string{"ruby"}},
			},
			expectedImage:    "mcs/code-server-base:latest",
			expectedDocker:   "Dockerfile.ruby",
			expectedFallback: "codercom/code-server:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetImageInfo(tt.language, tt.components)

			assert.Equal(t, tt.expectedImage, result.Image, "Image should match expected")
			assert.Equal(t, tt.expectedDocker, result.Dockerfile, "Dockerfile should match expected")
			assert.Equal(t, tt.expectedFallback, result.FallbackImage, "Fallback image should match expected")
		})
	}
}

// TestGetImageInfo_EdgeCases tests edge cases for GetImageInfo
func TestGetImageInfo_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		language   string
		components []components.Component
		validate   func(t *testing.T, result ImageInfo)
	}{
		{
			name:       "Empty language string",
			language:   "",
			components: []components.Component{},
			validate: func(t *testing.T, result ImageInfo) {
				// Empty language should be treated as generic
				assert.Equal(t, "mcs/code-server-base:latest", result.Image)
				assert.Equal(t, "Dockerfile.", result.Dockerfile) // Empty language results in Dockerfile.
			},
		},
		{
			name:     "Component with multiple requirements including nodejs",
			language: "python",
			components: []components.Component{
				{
					Name:     "Full Stack Tool",
					Selected: true,
					Requires: []string{"python", "nodejs", "docker"},
				},
			},
			validate: func(t *testing.T, result ImageInfo) {
				// Should detect nodejs requirement
				assert.Equal(t, "mcs/code-server-python-node:latest", result.Image)
				assert.Equal(t, "Dockerfile.python-node", result.Dockerfile)
			},
		},
		{
			name:     "Component with empty requirements",
			language: "go",
			components: []components.Component{
				{
					Name:     "Basic Tool",
					Selected: true,
					Requires: []string{},
				},
			},
			validate: func(t *testing.T, result ImageInfo) {
				// Should use base Go image
				assert.Equal(t, "mcs/code-server-go:latest", result.Image)
				assert.Equal(t, "Dockerfile.go", result.Dockerfile)
			},
		},
		{
			name:     "Component with nil requirements",
			language: "node",
			components: []components.Component{
				{
					Name:     "Basic Tool",
					Selected: true,
					Requires: nil,
				},
			},
			validate: func(t *testing.T, result ImageInfo) {
				// Should use Node image (already has Node.js)
				assert.Equal(t, "mcs/code-server-node:latest", result.Image)
				assert.Equal(t, "Dockerfile.node", result.Dockerfile)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetImageInfo(tt.language, tt.components)
			tt.validate(t, result)
		})
	}
}

// TestGenerateEnvFile tests the GenerateEnvFile function
func TestGenerateEnvFile(t *testing.T) {
	tests := []struct {
		name           string
		config         ComposeConfig
		expectedSubstr []string
	}{
		{
			name: "Basic environment file",
			config: ComposeConfig{
				CodespaceName: "test-project",
				Password:      "secret123",
				Environment: map[string]string{
					"NODE_ENV": "development",
					"DEBUG":    "true",
				},
			},
			expectedSubstr: []string{
				"# MCS Codespace Environment",
				"CODESPACE_NAME=test-project",
				"PASSWORD=secret123",
				"NODE_ENV=development",
				"DEBUG=true",
			},
		},
		{
			name: "Empty environment variables",
			config: ComposeConfig{
				CodespaceName: "empty-env",
				Password:      "pass123",
				Environment:   map[string]string{},
			},
			expectedSubstr: []string{
				"# MCS Codespace Environment",
				"CODESPACE_NAME=empty-env",
				"PASSWORD=pass123",
			},
		},
		{
			name: "No environment variables",
			config: ComposeConfig{
				CodespaceName: "no-env",
				Password:      "noenv123",
			},
			expectedSubstr: []string{
				"# MCS Codespace Environment",
				"CODESPACE_NAME=no-env",
				"PASSWORD=noenv123",
			},
		},
		{
			name: "Special characters in values",
			config: ComposeConfig{
				CodespaceName: "special-chars",
				Password:      "p@ss!w0rd#",
				Environment: map[string]string{
					"API_URL":    "https://api.example.com/v1",
					"SECRET_KEY": "sk_test_123!@#$%^&*()",
				},
			},
			expectedSubstr: []string{
				"CODESPACE_NAME=special-chars",
				"PASSWORD=p@ss!w0rd#",
				"API_URL=https://api.example.com/v1",
				"SECRET_KEY=sk_test_123!@#$%^&*()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateEnvFile(tt.config)
			assert.NotEmpty(t, result)

			resultStr := string(result)

			// Check that it ends with newline
			assert.True(t, strings.HasSuffix(resultStr, "\n"), "Env file should end with newline")

			// Check expected substrings
			for _, expected := range tt.expectedSubstr {
				assert.Contains(t, resultStr, expected, "Expected substring not found: %s", expected)
			}

			// Verify structure
			lines := strings.Split(strings.TrimSpace(resultStr), "\n")
			assert.True(t, len(lines) >= 3, "Env file should have at least header, codespace name, and password")

			// First line should be comment
			assert.True(t, strings.HasPrefix(lines[0], "#"), "First line should be a comment")
		})
	}
}

// TestLanguageImages tests the language image mappings
func TestLanguageImages(t *testing.T) {
	// Test that language image maps are properly defined
	assert.Contains(t, languageImages, "python")
	assert.Contains(t, languageImages, "node")
	assert.Contains(t, languageImages, "go")
	assert.Contains(t, languageImages, "generic")

	assert.Contains(t, languageImagesWithNode, "python")
	assert.Contains(t, languageImagesWithNode, "node")
	assert.Contains(t, languageImagesWithNode, "go")
	assert.Contains(t, languageImagesWithNode, "generic")

	// Test that images are defined for supported languages
	supportedLanguages := []string{"python", "node", "go", "rust", "java", "php", "ruby", "generic"}
	
	for _, lang := range supportedLanguages {
		assert.Contains(t, languageImages, lang, "Language %s should have base image", lang)
		assert.Contains(t, languageImagesWithNode, lang, "Language %s should have Node.js image", lang)
		assert.NotEmpty(t, languageImages[lang], "Base image for %s should not be empty", lang)
		assert.NotEmpty(t, languageImagesWithNode[lang], "Node.js image for %s should not be empty", lang)
	}

	// Test that node language uses the same image for both cases
	assert.Equal(t, languageImages["node"], languageImagesWithNode["node"], "Node language should use same image in both cases")
}

// TestDockerComposeTemplate tests the docker-compose template structure
func TestDockerComposeTemplate(t *testing.T) {
	// Verify template contains essential sections
	assert.Contains(t, dockerComposeTemplate, "services:")
	assert.Contains(t, dockerComposeTemplate, "{{ .ContainerName }}:")
	assert.Contains(t, dockerComposeTemplate, "image: {{ .Image }}")
	assert.Contains(t, dockerComposeTemplate, "container_name: {{ .ContainerName }}")
	assert.Contains(t, dockerComposeTemplate, "restart: unless-stopped")
	assert.Contains(t, dockerComposeTemplate, "PASSWORD={{ .Password }}")
	assert.Contains(t, dockerComposeTemplate, "ports:")
	assert.Contains(t, dockerComposeTemplate, "volumes:")
	assert.Contains(t, dockerComposeTemplate, "labels:")
	assert.Contains(t, dockerComposeTemplate, "working_dir:")
	assert.Contains(t, dockerComposeTemplate, "networks:")
	assert.Contains(t, dockerComposeTemplate, "mcs-network")
	
	// Verify conditional sections
	assert.Contains(t, dockerComposeTemplate, "{{- if .BuildContext }}")
	assert.Contains(t, dockerComposeTemplate, "build:")
	assert.Contains(t, dockerComposeTemplate, "context: {{ .BuildContext }}")
	assert.Contains(t, dockerComposeTemplate, "dockerfile: {{ .Dockerfile }}")
	assert.Contains(t, dockerComposeTemplate, "{{- end }}")
	
	assert.Contains(t, dockerComposeTemplate, "{{- if .Components }}")
	assert.Contains(t, dockerComposeTemplate, "./components:/home/coder/.components:ro")
	assert.Contains(t, dockerComposeTemplate, "./init:/docker-entrypoint-initdb.d:ro")
}

// TestInitScriptTemplate tests the init script template structure
func TestInitScriptTemplate(t *testing.T) {
	// Verify template contains essential sections
	assert.Contains(t, initScriptTemplate, "#!/bin/bash")
	assert.Contains(t, initScriptTemplate, "# Component installation script")
	assert.Contains(t, initScriptTemplate, "set -e")
	assert.Contains(t, initScriptTemplate, "🚀 Installing MCS components...")
	assert.Contains(t, initScriptTemplate, "✨ Component installation complete!")
	
	// Verify component iteration
	assert.Contains(t, initScriptTemplate, "{{- range .Components }}")
	assert.Contains(t, initScriptTemplate, "{{- if .Selected }}")
	assert.Contains(t, initScriptTemplate, "echo \"📦 Installing {{ .Name }}...\"")
	assert.Contains(t, initScriptTemplate, "/home/coder/.components/{{ .Installer }} install")
	assert.Contains(t, initScriptTemplate, "echo \"✅ {{ .Name }} installed successfully\"")
	
	// Verify directory creation and PATH setup
	assert.Contains(t, initScriptTemplate, "mkdir -p /home/coder/.local/bin /home/coder/.local/share")
	assert.Contains(t, initScriptTemplate, "mkdir -p /home/coder/.npm-global/bin")
	assert.Contains(t, initScriptTemplate, "export PATH=")
	assert.Contains(t, initScriptTemplate, "export NPM_PREFIX=")
}

// Benchmark tests for performance
func BenchmarkGenerateDockerCompose(b *testing.B) {
	config := ComposeConfig{
		ContainerName: "benchmark-test",
		CodespaceName: "benchmark",
		Image:         "mcs/code-server:latest",
		Password:      "benchpass",
		Ports: map[string]string{
			"8080": "8080",
			"3000": "3000",
			"5432": "5432",
		},
		Environment: map[string]string{
			"NODE_ENV": "production",
			"DEBUG":    "false",
			"API_URL":  "https://api.example.com",
		},
		Labels: map[string]string{
			"version":     "1.0.0",
			"maintainer":  "test@example.com",
			"environment": "production",
		},
		Components: []components.Component{
			{Name: "Node.js", Selected: true, Installer: "install-node.sh"},
			{Name: "Python", Selected: true, Installer: "install-python.sh"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateDockerCompose(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateInitScript(b *testing.B) {
	components := []components.Component{
		{Name: "Node.js", Selected: true, Installer: "install-node.sh"},
		{Name: "Python", Selected: true, Installer: "install-python.sh"},
		{Name: "Go", Selected: true, Installer: "install-go.sh"},
		{Name: "Docker", Selected: true, Installer: "install-docker.sh"},
		{Name: "Git", Selected: true, Installer: "install-git.sh"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateInitScript(components)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetImageInfo(b *testing.B) {
	components := []components.Component{
		{Name: "ESLint", Selected: true, Requires: []string{"nodejs"}},
		{Name: "Python Tools", Selected: true, Requires: []string{"python"}},
		{Name: "Docker", Selected: false, Requires: []string{"docker"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetImageInfo("python", components)
	}
}

// Test helper functions
func TestGenerateDockerCompose_NilMaps(t *testing.T) {
	config := ComposeConfig{
		ContainerName: "nil-test",
		CodespaceName: "nil-test",
		Password:      "testpass",
		// Ports, Environment, Labels are nil
	}

	result, err := GenerateDockerCompose(config)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	resultStr := string(result)
	assert.Contains(t, resultStr, "mcs.managed: \"true\"")
	assert.Contains(t, resultStr, "mcs.codespace: \"nil-test\"")
}

func TestGenerateEnvFile_EmptyConfig(t *testing.T) {
	config := ComposeConfig{} // All fields empty

	result := GenerateEnvFile(config)
	assert.NotEmpty(t, result)

	resultStr := string(result)
	assert.Contains(t, resultStr, "# MCS Codespace Environment")
	assert.Contains(t, resultStr, "CODESPACE_NAME=")
	assert.Contains(t, resultStr, "PASSWORD=")
}