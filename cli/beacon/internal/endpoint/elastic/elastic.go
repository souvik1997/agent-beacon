package elastic

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed pack/*
var packFS embed.FS

const (
	DefaultLogPath   = "/var/log/beacon-agent/runtime.jsonl"
	DefaultOutputDir = "beacon-elastic-pack"
)

type File struct {
	Name    string
	Content string
}

func Files() []File {
	return []File{
		{Name: "README.md", Content: mustRead("pack/README.md")},
		{Name: "filebeat.yml", Content: InputSnippet(DefaultLogPath)},
		{Name: "elastic-agent-standalone.yml", Content: standaloneAgentConfig(DefaultLogPath)},
		{Name: "ilm-policy.json", Content: mustRead("pack/ilm-policy.json")},
		{Name: "component-template-mappings.json", Content: mustRead("pack/component-template-mappings.json")},
		{Name: "component-template-settings.json", Content: mustRead("pack/component-template-settings.json")},
		{Name: "index-template.json", Content: mustRead("pack/index-template.json")},
		{Name: "ingest-pipeline.json", Content: mustRead("pack/ingest-pipeline.json")},
		{Name: "kibana-assets.ndjson", Content: mustRead("pack/kibana-assets.ndjson")},
		{Name: "docker-compose.yml", Content: mustRead("pack/docker-compose.yml")},
		{Name: "sample-event.jsonl", Content: mustRead("pack/sample-event.jsonl")},
	}
}

func InputSnippet(logPath string) string {
	if logPath == "" {
		logPath = DefaultLogPath
	}
	return strings.ReplaceAll(mustRead("pack/filebeat.yml.tmpl"), "{{LOG_PATH}}", logPath)
}

func InstallPack(outputDir, logPath string) error {
	if outputDir == "" {
		outputDir = DefaultOutputDir
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	for _, file := range Files() {
		content := file.Content
		switch file.Name {
		case "filebeat.yml":
			content = InputSnippet(logPath)
		case "elastic-agent-standalone.yml":
			content = standaloneAgentConfig(logPath)
		}
		if err := os.WriteFile(filepath.Join(outputDir, file.Name), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func standaloneAgentConfig(logPath string) string {
	if logPath == "" {
		logPath = DefaultLogPath
	}
	return strings.ReplaceAll(mustRead("pack/elastic-agent-standalone.yml.tmpl"), "{{LOG_PATH}}", logPath)
}

func mustRead(path string) string {
	data, err := packFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("elastic pack asset %s: %v", path, err))
	}
	return string(data)
}
