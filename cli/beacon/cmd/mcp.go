package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/activity"
	endpointconfig "github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/config"
	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/dashboard"
	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/lifecycle"
	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/endpoint/writer"
	"github.com/asymptote-labs/agent-beacon/cli/beacon/internal/mcpserver"
)

const defaultMCPHTTPAddr = "127.0.0.1:8766"

var mcpOpts struct {
	userMode   bool
	systemMode bool
	logPath    string
	transport  string
	addr       string
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Expose local Beacon activity through MCP",
}

var mcpServeCmd = &cobra.Command{
	Use:          "serve",
	Short:        "Run the local Beacon MCP server",
	SilenceUsage: true,
	RunE:         runMCPServe,
}

var mcpDoctorCmd = &cobra.Command{
	Use:          "doctor",
	Short:        "Validate local Beacon MCP setup",
	SilenceUsage: true,
	RunE:         runMCPDoctor,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)
	mcpCmd.AddCommand(mcpDoctorCmd)
	for _, c := range []*cobra.Command{mcpServeCmd, mcpDoctorCmd} {
		c.Flags().BoolVar(&mcpOpts.userMode, "user", true, "Use per-user endpoint paths")
		c.Flags().BoolVar(&mcpOpts.systemMode, "system", false, "Use system endpoint paths")
		c.Flags().StringVar(&mcpOpts.logPath, "log-path", "", "Runtime JSONL log path")
		c.Flags().StringVar(&mcpOpts.transport, "transport", mcpserver.TransportStdio, "MCP transport: stdio or http")
		c.Flags().StringVar(&mcpOpts.addr, "addr", defaultMCPHTTPAddr, "Loopback HTTP listen address for --transport http")
	}
}

func runMCPServe(cmd *cobra.Command, args []string) error {
	if err := mcpserver.ValidateTransport(mcpOpts.transport); err != nil {
		return err
	}
	runtimeLog := resolveMCPRuntimeLog()
	server := mcpserver.New(mcpserver.Options{LogPath: runtimeLog.EffectiveLogPath, Stderr: os.Stderr})
	switch normalizedTransport() {
	case mcpserver.TransportStdio:
		return server.ServeStdio(cmd.Context(), os.Stdin, os.Stdout)
	case mcpserver.TransportHTTP:
		if err := dashboard.ValidateLoopbackAddr(mcpOpts.addr); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Beacon MCP server: http://%s/mcp\n", mcpOpts.addr)
		fmt.Fprintf(os.Stderr, "Runtime log: %s\n", runtimeLog.EffectiveLogPath)
		httpServer := &http.Server{
			Addr:              mcpOpts.addr,
			Handler:           server.HTTPHandler(),
			ReadHeaderTimeout: 5 * time.Second,
		}
		return httpServer.ListenAndServe()
	default:
		return fmt.Errorf("unsupported transport %q", mcpOpts.transport)
	}
}

func runMCPDoctor(cmd *cobra.Command, args []string) error {
	if err := mcpserver.ValidateTransport(mcpOpts.transport); err != nil {
		return err
	}
	runtimeLog := resolveMCPRuntimeLog()
	cfg := loadMCPConfig(runtimeLog.EffectiveUserMode, runtimeLog.EffectiveLogPath)
	fmt.Println("Beacon MCP doctor")
	fmt.Println()
	fmt.Printf("Transport: %s\n", normalizedTransport())
	if normalizedTransport() == mcpserver.TransportHTTP {
		if err := dashboard.ValidateLoopbackAddr(mcpOpts.addr); err != nil {
			return err
		}
		fmt.Printf("HTTP address: %s\n", mcpOpts.addr)
		if err := checkCanListen(mcpOpts.addr); err != nil {
			return err
		}
		fmt.Println("HTTP bind check: ok")
	}
	fmt.Printf("Runtime log: %s\n", runtimeLog.EffectiveLogPath)
	fmt.Printf("Runtime log source: %s\n", runtimeLog.Reason)
	if runtimeLog.Warning != "" {
		fmt.Printf("Runtime log warning: %s\n", runtimeLog.Warning)
	}
	fmt.Printf("Content retention: %s\n", cfg.ContentRetention)
	sampled, malformed, archives, err := activity.InspectLog(runtimeLog.EffectiveLogPath)
	if err != nil {
		return err
	}
	fmt.Printf("Sampled events: %d\n", sampled)
	fmt.Printf("Malformed lines: %d\n", malformed)
	fmt.Printf("Readable archives: %d\n", len(archives))
	server := mcpserver.New(mcpserver.Options{LogPath: runtimeLog.EffectiveLogPath, Stderr: os.Stderr})
	if err := server.HasExpectedTools(); err != nil {
		return err
	}
	fmt.Printf("MCP tools: %s\n", strings.Join(server.ToolNames(), ", "))
	if normalizedTransport() == mcpserver.TransportStdio {
		fmt.Println()
		fmt.Println("MCP client config:")
		fmt.Println(`{`)
		fmt.Println(`  "mcpServers": {`)
		fmt.Println(`    "beacon": {`)
		fmt.Println(`      "command": "beacon",`)
		fmt.Println(`      "args": ["mcp", "serve", "--transport", "stdio"]`)
		fmt.Println(`    }`)
		fmt.Println(`  }`)
		fmt.Println(`}`)
	}
	return nil
}

func resolveMCPRuntimeLog() lifecycle.RuntimeLogSource {
	return lifecycle.ResolveRuntimeLog(mcpUserMode(), mcpOpts.logPath)
}

func loadMCPConfig(userMode bool, logPath string) endpointconfig.Config {
	if cfg, err := endpointconfig.Load(userMode); err == nil {
		if logPath != "" {
			cfg.LogPath = logPath
		}
		return cfg
	}
	if logPath == "" {
		logPath = writer.DefaultPath(userMode)
	}
	return endpointconfig.Default(userMode, logPath)
}

func mcpUserMode() bool {
	return mcpOpts.userMode && !mcpOpts.systemMode
}

func normalizedTransport() string {
	if mcpOpts.transport == "" {
		return mcpserver.TransportStdio
	}
	return mcpOpts.transport
}

func checkCanListen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("HTTP bind check failed for %s: %w", addr, err)
	}
	return listener.Close()
}
