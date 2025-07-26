package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mathrand "math/rand"
	"strings"
	"time"
)

var (
	adjectives = []string{
		"happy", "clever", "brave", "calm", "eager", "fancy", "gentle", "jolly",
		"kind", "lively", "nice", "proud", "silly", "witty", "zealous", "cosmic",
		"electric", "melodic", "quantum", "serene", "vibrant", "whimsical",
		"dazzling", "groovy", "mystical", "radiant", "stellar", "tranquil",
		"fierce", "noble", "swift", "wise", "bold", "cheerful", "dynamic",
		"elegant", "friendly", "graceful", "heroic", "inspired", "joyful",
		"luminous", "magical", "nimble", "peaceful", "quirky", "resilient",
		"spirited", "thoughtful", "unique", "valiant", "wonderful", "zesty",
		"brilliant", "charming", "delightful", "energetic", "fabulous", "glorious",
		"harmonious", "incredible", "jubilant", "kinetic", "legendary", "magnificent",
	}
	
	nouns = []string{
		"panda", "koala", "otter", "penguin", "dolphin", "eagle", "falcon",
		"giraffe", "hamster", "iguana", "jaguar", "kitten", "lemur", "monkey",
		"narwhal", "octopus", "parrot", "quokka", "rabbit", "sloth", "turtle",
		"unicorn", "viper", "walrus", "xerus", "yak", "zebra", "alpaca",
		"badger", "cheetah", "dragon", "elephant", "flamingo", "gecko",
		"hedgehog", "impala", "jellyfish", "kangaroo", "llama", "meerkat",
		"newt", "owl", "platypus", "quail", "raccoon", "seahorse", "toucan",
		"urchin", "vulture", "wombat", "fox", "bear", "wolf", "lynx",
		"moose", "bison", "crane", "dove", "elk", "ferret", "gazelle",
		"heron", "ibis", "jackal", "kiwi", "lobster", "mantis", "nightingale",
	}
)

func init() {
	mathrand.Seed(time.Now().UnixNano())
}

// GenerateFunName generates a fun name like "happy-panda"
func GenerateFunName(base string) string {
	adj := adjectives[mathrand.Intn(len(adjectives))]
	noun := nouns[mathrand.Intn(len(nouns))]
	return fmt.Sprintf("%s-%s", adj, noun)
}

// GenerateCodespaceName creates a name from repository owner and name
func GenerateCodespaceName(owner, repoName string) string {
	// Sanitize owner and repo name
	owner = SanitizeName(owner)
	repoName = SanitizeName(repoName)
	
	// Create base name from owner-repo format
	return fmt.Sprintf("%s-%s", owner, repoName)
}

// GenerateUniqueCodespaceName creates a unique name with collision detection
func GenerateUniqueCodespaceName(owner, repoName string, checkExists func(string) bool) string {
	baseName := GenerateCodespaceName(owner, repoName)
	
	// If base name doesn't exist, use it
	if !checkExists(baseName) {
		return baseName
	}
	
	// Try with funny suffixes
	maxAttempts := 50
	for i := 0; i < maxAttempts; i++ {
		funnySuffix := GenerateFunName("")
		uniqueName := fmt.Sprintf("%s-%s", baseName, funnySuffix)
		
		if !checkExists(uniqueName) {
			return uniqueName
		}
	}
	
	// Fallback: use timestamp
	return fmt.Sprintf("%s-%d", baseName, time.Now().Unix())
}

// SanitizeName ensures a name is valid for use as a container/directory name
func SanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	
	// Replace invalid characters with hyphens
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		} else if len(result) > 0 && result[len(result)-1] != '-' {
			result = append(result, '-')
		}
	}
	
	// Trim hyphens from start and end
	name = strings.Trim(string(result), "-")
	
	// Ensure it's not empty
	if name == "" {
		name = "codespace"
	}
	
	return name
}

// GenerateSecurePassword generates a secure 16-character password
func GenerateSecurePassword() string {
	// Generate 12 random bytes (will become 16 chars in base64)
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to math/rand if crypto/rand fails
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result := make([]byte, 16)
		for i := range result {
			result[i] = charset[mathrand.Intn(len(charset))]
		}
		return string(result)
	}
	
	// Encode to base64 and clean up
	password := base64.URLEncoding.EncodeToString(b)
	// Remove padding and special chars, take first 16 chars
	password = strings.ReplaceAll(password, "=", "")
	password = strings.ReplaceAll(password, "-", "")
	password = strings.ReplaceAll(password, "_", "")
	
	// Ensure we have at least 16 chars
	for len(password) < 16 {
		password += string([]byte{byte(mathrand.Intn(26) + 65)}) // Add random uppercase letter
	}
	
	return password[:16]
}