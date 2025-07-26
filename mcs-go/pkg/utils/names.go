package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	adjectives = []string{
		"happy", "clever", "brave", "calm", "eager", "fancy", "gentle", "jolly",
		"kind", "lively", "nice", "proud", "silly", "witty", "zealous", "cosmic",
		"electric", "melodic", "quantum", "serene", "vibrant", "whimsical",
		"dazzling", "groovy", "mystical", "radiant", "stellar", "tranquil",
	}
	
	nouns = []string{
		"panda", "koala", "otter", "penguin", "dolphin", "eagle", "falcon",
		"giraffe", "hamster", "iguana", "jaguar", "kitten", "lemur", "monkey",
		"narwhal", "octopus", "parrot", "quokka", "rabbit", "sloth", "turtle",
		"unicorn", "viper", "walrus", "xerus", "yak", "zebra",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateFunName generates a fun name like "happy-panda"
func GenerateFunName(base string) string {
	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	return fmt.Sprintf("%s-%s", adj, noun)
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