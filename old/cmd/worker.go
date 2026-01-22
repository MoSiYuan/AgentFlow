package cmd

import (
	"fmt"
	"os"

	"github.com/jiangxiaolong/cpds-go/internal/config"
	"github.com/jiangxiaolong/cpds-go/internal/worker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start CPDS Worker client",
	Long: `Start the CPDS Worker client.

The Worker client is responsible for:
  - Registering with Master
  - Pulling and executing tasks
  - Reporting progress and results

Modes:
  - cloud: Worker runs continuously, calling Claude API for tasks
  - standalone: Worker exits after completing one task`,
	RunE: runWorker,
}

var (
	masterURL string
	workerName string
	autoMode bool
	oneShot bool
)

func init() {
	rootCmd.AddCommand(workerCmd)

	workerCmd.Flags().StringVarP(&masterURL, "master", "m", "http://localhost:8848", "Master server URL")
	workerCmd.Flags().StringVarP(&workerName, "name", "n", "", "Worker name (default: hostname)")
	workerCmd.Flags().BoolVarP(&autoMode, "auto", "a", false, "Automatic mode (pull tasks without prompting)")
	workerCmd.Flags().BoolVar(&oneShot, "oneshot", false, "Execute one task then exit (standalone mode)")

	viper.BindPFlag("worker.master", workerCmd.Flags().Lookup("master"))
	viper.BindPFlag("worker.name", workerCmd.Flags().Lookup("name"))
	viper.BindPFlag("worker.auto", workerCmd.Flags().Lookup("auto"))
	viper.BindPFlag("worker.oneshot", workerCmd.Flags().Lookup("oneshot"))
}

func runWorker(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with CLI flags
	if cmd.Flags().Changed("master") {
		cfg.Worker.MasterURL = masterURL
	}
	if cmd.Flags().Changed("name") {
		cfg.Worker.Name = workerName
	}
	if cmd.Flags().Changed("auto") {
		cfg.Worker.Auto = autoMode
	}
	if cmd.Flags().Changed("oneshot") {
		cfg.Worker.OneShot = oneShot
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
	fmt.Printf("║            🤖 CPDS Worker Client Started                 ║\n")
	fmt.Printf("╠═══════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Mode:       %-42s ║\n", mode)
	fmt.Printf("║  Master:     %-42s ║\n", cfg.Worker.MasterURL)
	if cfg.Worker.OneShot {
		fmt.Printf("║  One-Shot:   %-42s ║\n", "ENABLED")
	}
	fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n\n")

	// Run in appropriate mode
	if cfg.Worker.OneShot || cfg.IsStandaloneMode() {
		// One-shot mode (standalone)
		runner := worker.NewOneShotRunner(cfg, logger)
		return runner.Run()
	} else {
		// Continuous mode (cloud)
		return fmt.Errorf("continuous cloud mode not yet implemented")
	}
}
