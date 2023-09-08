package config

import (
	"github.com/Fallen-Breath/etunnel/internal/constants"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

type cliFlags struct {
	// CLI stuffs
	Help       bool
	ConfigPath string

	// config common
	Mode  string
	Crypt string
	Key   string
	Cork  bool
	Debug bool

	// config - server
	Listen string

	// config - client
	Server  string
	Tunnels []string

	// config - tool
	Pid int
}

func CliEntry() *Config {
	var flags cliFlags
	var rootCmd = &cobra.Command{
		Use:     constants.Name,
		Short:   constants.Description,
		Version: constants.Version,
	}
	rootCmd.Flags().BoolVar(&flags.Help, "h", false, "Show help and exit")
	rootCmd.Flags().StringVar(&flags.ConfigPath, "conf", "", "Set the file to load config from. Arguments from command line will be ignored")

	addTunnelFlags := func(fs *pflag.FlagSet) {
		fs.StringVarP(&flags.Mode, "mode", "m", "", "The mode of etunnel. Options: client, server")
		fs.StringVarP(&flags.Crypt, "crypt", "c", Crypts[0], "The encryption method to use. Options: "+strings.Join(Crypts, ", "))
		fs.StringVarP(&flags.Key, "key", "k", "hidden secret", "The secret password for encryption")
		fs.BoolVar(&flags.Cork, "cork", false, "Enable tcp corking")
		fs.BoolVar(&flags.Debug, "debug", false, "Enable debug logging")
	}

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run etunnel as the server",
		Run: func(cmd *cobra.Command, args []string) {
			flags.Mode = ModeServer
		},
	}
	rootCmd.AddCommand(serverCmd)
	addTunnelFlags(serverCmd.Flags())
	serverCmd.Flags().StringVarP(&flags.Listen, "listen", "l", "127.0.0.1:12000", "(server) The address to listen to")

	var clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Run etunnel as the client",
		Run: func(cmd *cobra.Command, args []string) {
			flags.Mode = ModeClient
		},
	}
	rootCmd.AddCommand(clientCmd)
	addTunnelFlags(clientCmd.Flags())
	clientCmd.Flags().StringVarP(&flags.Server, "server", "s", "127.0.0.1:12000", "(client) The address of the etunnel server")
	clientCmd.Flags().StringSliceVarP(&flags.Tunnels, "tunnel", "t", nil, "(client) A list of encrypted tunnels")

	var toolCmd = &cobra.Command{
		Use:   "tool",
		Short: "Run etunnel as a tool",
		Run: func(cmd *cobra.Command, args []string) {
			flags.Mode = ModeTool
		},
	}
	rootCmd.AddCommand(toolCmd)
	toolCmd.Flags().IntVarP(&flags.Pid, "pid", "p", 0, "The pid of the etunnel process to interact with")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Execute command failed: %v", err)
	}
	if len(flags.Mode) == 0 {
		os.Exit(0)
	}

	return generateConfigOrDie(&flags)
}

func generateConfigOrDie(flags *cliFlags) *Config {
	var conf *Config
	var err error
	if len(flags.ConfigPath) > 0 {
		if conf, err = LoadConfigFromFile(flags.ConfigPath); err != nil {
			log.Fatalf("Read config file failed: %v", err)
		}
		log.Infof("Loaded config from %s", flags.ConfigPath)
	} else {
		if conf, err = LoadConfigFromFlags(flags); err != nil {
			log.Fatalf("Parse CLI flags failed: %v", err)
		}
	}

	if err := ValidateConfig(conf); err != nil {
		log.Fatalf("Validate config failed: %v", err)
	}
	return conf
}