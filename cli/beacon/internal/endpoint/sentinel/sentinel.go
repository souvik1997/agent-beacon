package sentinel

import (
	"embed"
	"fmt"

	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/siempack"
)

//go:embed pack/*
var packFS embed.FS

const (
	DefaultLogPath   = "/var/log/beacon-agent/runtime.jsonl"
	DefaultOutputDir = "beacon-sentinel-pack"
)

func Files() []siempack.File {
	return []siempack.File{
		{Name: "README.md", Content: mustRead("pack/README.md")},
		{Name: "dcr-transform.kql", Content: DCRTransform()},
		{Name: "table-schema.json", Content: mustRead("pack/table-schema.json")},
		{Name: "dcr-template.json", Content: mustRead("pack/dcr-template.json"), TemplateLogPath: true},
		{Name: "queries.kql", Content: mustRead("pack/queries.kql")},
		{Name: "detections.kql", Content: mustRead("pack/detections.kql")},
		{Name: "sample-event.jsonl", Content: mustRead("pack/sample-event.jsonl")},
	}
}

func DCRTransform() string {
	return mustRead("pack/dcr-transform.kql")
}

func InstallPack(outputDir, logPath string) error {
	if outputDir == "" {
		outputDir = DefaultOutputDir
	}
	if logPath == "" {
		logPath = DefaultLogPath
	}
	return siempack.Install(outputDir, Files(), logPath)
}

func mustRead(path string) string {
	data, err := siempack.ReadFile(packFS, path)
	if err != nil {
		panic(fmt.Sprintf("sentinel pack asset %s: %v", path, err))
	}
	return data
}
