package changelog

import (
	"fmt"
	"strings"
)

// ProcessorType definiert die verfügbaren Processor-Typen
type ProcessorType string

const (
	ProcessorTypeGitHub    ProcessorType = "github"
	ProcessorTypeGitLab    ProcessorType = "gitlab"
	ProcessorTypeBitbucket ProcessorType = "bitbucket"
)

// ProcessorFactory erstellt Processor-Instanzen basierend auf dem Typ
type ProcessorFactory struct{}

// NewProcessorFactory erstellt eine neue ProcessorFactory
func NewProcessorFactory() *ProcessorFactory {
	return &ProcessorFactory{}
}

// CreateProcessor erstellt einen Processor basierend auf dem Typ und Host
func (f *ProcessorFactory) CreateProcessor(processorType ProcessorType, host string) (Processor, error) {
	switch processorType {
	case ProcessorTypeGitHub:
		processor := &GitHubProcessor{
			Host: host,
		}
		return processor, nil
	case ProcessorTypeGitLab:
		processor := &GitLabProcessor{
			Host: host,
		}
		return processor, nil
	case ProcessorTypeBitbucket:
		processor := &BitbucketProcessor{
			Host: host,
		}
		return processor, nil
	default:
		return nil, fmt.Errorf("unknown processor type: %s", processorType)
	}
}

// CreateProcessorFromString erstellt einen Processor aus einem String
func (f *ProcessorFactory) CreateProcessorFromString(processorStr string) (Processor, error) {
	if processorStr == "" {
		return nil, nil
	}

	// Format: "type" oder "type:host"
	parts := strings.SplitN(processorStr, ":", 2)
	processorType := ProcessorType(strings.ToLower(parts[0]))

	host := ""
	if len(parts) > 1 {
		host = parts[1]
	}

	return f.CreateProcessor(processorType, host)
}

// GetAvailableProcessors gibt eine Liste der verfügbaren Processor-Typen zurück
func (f *ProcessorFactory) GetAvailableProcessors() []ProcessorType {
	return []ProcessorType{
		ProcessorTypeGitHub,
		ProcessorTypeGitLab,
		ProcessorTypeBitbucket,
	}
}

// GetDefaultHost gibt den Standard-Host für einen Processor-Typ zurück
func (f *ProcessorFactory) GetDefaultHost(processorType ProcessorType) string {
	switch processorType {
	case ProcessorTypeGitHub:
		return "https://github.com"
	case ProcessorTypeGitLab:
		return "https://gitlab.com"
	case ProcessorTypeBitbucket:
		return "https://bitbucket.org"
	default:
		return ""
	}
}
