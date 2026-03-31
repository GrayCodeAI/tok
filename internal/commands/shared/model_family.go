package shared

import "github.com/GrayCodeAI/tokman/internal/utils"

// GetModelFamily extracts the model family from a model name.
// Delegates to utils.GetModelFamily for single source of truth.
func GetModelFamily(modelName string) string {
	return utils.GetModelFamily(modelName)
}
