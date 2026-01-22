package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jiangxiaolong/cpds-go/internal/config"
	"github.com/jiangxiaolong/cpds-go/internal/master"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Start CPDS Master server",
	Long: `Start the CPDS Master server.

The Master server is responsible for:
  - Task management and distribution
  - Worker registration and monitoring
  - Progress tracking and reporting

Modes:
  - cloud: Master runs continuously, workers call Claude API
  - standalone: Master shuts down after all tasks complete`,
	RunE: runMaster,
}

var (
	masterHost string
	masterPort int
	autoShutdown bool
)

func init() {
	rootCmd.AddCommand(masterCmd)

	masterCmd.Flags().StringVarP(&masterHost, "host", "H", "0.0.0.0", "Master server host")
	masterCmd.Flags().IntVarP(&masterPort, "port", "p", 8848, "Master server port")
	masterCmd.Flags().BoolVar(&autoShutdown, "auto-shutdown", false, "Auto-shutdown after all tasks complete (standalone mode)")

	viper.BindPFlag("master.host", masterCmd.Flags().Lookup("host"))
	viper.BindPFlag("master.port", masterCmd.Flags().Lookup("port"))
	viper.BindPFlag("master.auto_shutdown", masterCmd.Flags().Lookup("auto-shutdown"))
}

func runMaster(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with CLI flags
	if cmd.Flags().Changed("host") {
		cfg.Master.Host = masterHost
	}
	if cmd.Flags().Changed("port") {
		cfg.Master.Port = masterPort
	}
	if cmd.Flags().Changed("auto-shutdown") {
		cfg.Master.AutoShutdown = autoShutdown
	}

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	if cfg.Log.Level == "debug" {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Print startup banner
	mode := cfg.Mode.String()
	fmt.Printf("\n╔═══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║           🚀 CPDS Master Server Started                   ║\n")
	fmt.Printf("╠═══════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Mode:       %-42s ║\n", mode)
	fmt.Printf("║  Host:       %-42s ║\n", cfg.Master.Host)
	fmt.Printf("║  Port:       %-42d ║\n", cfg.Master.Port)
	if cfg.Master.AutoShutdown {
		fmt.Printf("║  Auto-Shutdown: %-37s ║\n", "ENABLED")
	}
	fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n\n")

	// Create server
	server, err := master.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigCh:
		logger.Info("Received shutdown signal")
	case err := <-errCh:
		if err != nil {
			logger.Errorf("Server error: %v", err)
			return err
		}
	}

	// Graceful shutdown
	logger.Info("Shutting down...")
	return nil
}
