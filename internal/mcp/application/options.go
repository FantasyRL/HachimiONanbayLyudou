package application

type ChatOptions struct {
	EnableWebSearch bool
	Model           string
	Temperature     *float64
	TopP            *float64
	TopK            *int
	MaxTokens       *int
}