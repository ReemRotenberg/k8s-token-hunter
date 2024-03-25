package internal

import (
	"fmt"
	"os"
)

// GenerateCombinationsToFile creates a file with all possible combinations of a fixed alphabet.
func GenerateCombinationsToFile(filepath string, combinationLength int) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer file.Close()

	if err := generateCombinationsToFile("", combinationLength, file); err != nil {
		return fmt.Errorf("error generating combinations: %w", err)
	}

	return nil
}

// generateCombinationsToFile is a helper function that recursively generates all possible combinations
func generateCombinationsToFile(prefix string, remainingLength int, file *os.File) error {
	const alphanums = "bcdfghjklmnpqrstvwxz2456789"

	if remainingLength == 0 {
		if _, err := file.WriteString(prefix + "\n"); err != nil {
			return err
		}
		return nil
	}

	for _, c := range alphanums {
		if err := generateCombinationsToFile(prefix+string(c), remainingLength-1, file); err != nil {
			return err
		}
	}
	return nil
}
