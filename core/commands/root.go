package commands

import (
	"errors"

	cmdenv "github.com/ipfs/go-ipfs/core/commands/cmdenv"
	name "github.com/ipfs/go-ipfs/core/commands/name"
	ocmd "github.com/ipfs/go-ipfs/core/commands/object"

	cmds "github.com/ipfs/go-ipfs-cmds"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("core/commands")

var ErrNotOnline = errors.New("this command must be run in online mode. Try running 'ipfs daemon' first")

const (
	ConfigOption  = "config"
	DebugOption   = "debug"
	OfflineOption = "offline"
	ApiOption     = "api"
)

var Root = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:  "Global p2p merkle-dag filesystem.",
		Synopsis: "ipfs [--config=<config> | -c] [--debug | -D] [--help] [-h] [--api=<api>] [--offline] [--cid-base=<base>] [--upgrade-cidv0-in-output] [--encoding=<encoding> | --enc] [--timeout=<timeout>] <command> ...",
		Subcommands: `
BASIC COMMANDS
  init          Initialize local IPFS configuration
  refs <ref>    List hashes of links from an object

ADVANCED COMMANDS
  daemon        Start a long-running daemon process
  resolve       Resolve any type of name
  name          Publish and resolve IPNS names
  key           Create and list IPNS name keypairs
  dns           Resolve DNS links
  stats         Various operational stats
  p2p           Libp2p stream mounting

NETWORK COMMANDS
  bootstrap     Add or remove bootstrap peers
  swarm         Manage connections to the p2p network
  dht           Query the DHT for values or peers
  ping          Measure the latency of a connection
  pubsub        Send and receive messages via pubsub

TOOL COMMANDS
  config        Manage configuration
  version       Show IPFS version information

Use 'ipfs <command> --help' to learn more about each command.

ipfs uses a repository in the local file system. By default, the repo is
located at ~/.ipfs. To change the repo location, set the $IPFS_PATH
environment variable:

  export IPFS_PATH=/path/to/ipfsrepo

EXIT STATUS

The CLI will exit with one of the following values:

0     Successful execution.
1     Failed executions.
`,
	},
	Options: []cmds.Option{
		cmds.StringOption(ConfigOption, "c", "Path to the configuration file to use."),
		cmds.BoolOption(DebugOption, "D", "Operate in debug mode."),
		cmds.BoolOption(cmds.OptLongHelp, "Show the full command help text."),
		cmds.BoolOption(cmds.OptShortHelp, "Show a short version of the command help text."),
		cmds.BoolOption(OfflineOption, "Run the command offline."),
		cmds.StringOption(ApiOption, "Use a specific API instance (defaults to /ip4/127.0.0.1/tcp/5001)"),

		// global options, added to every command
		cmdenv.OptionCidBase,
		cmdenv.OptionUpgradeCidV0InOutput,

		cmds.OptionEncodingType,
		cmds.OptionStreamChannels,
		cmds.OptionTimeout,
	},
}

var rootSubcommands = map[string]*cmds.Command{
	"pubsub":    PubsubCmd,
	"stats":     StatsCmd,
	"bootstrap": BootstrapCmd,
	"config":    ConfigCmd,
	"dht":       DhtCmd,
	"dns":       DNSCmd,
	"key":       KeyCmd,
	"name":      name.NameCmd,
	"object":    ocmd.ObjectCmd,
	"ping":      PingCmd,
	"p2p":       P2PCmd,
	"resolve":   ResolveCmd,
	"swarm":     SwarmCmd,
	"version":   VersionCmd,
}

// RootRO is the readonly version of Root
var RootRO = &cmds.Command{}

// RefsROCmd is `ipfs refs` command
var RefsROCmd = &cmds.Command{}

// VersionROCmd is `ipfs version` command (without deps).
var VersionROCmd = &cmds.Command{}

var rootROSubcommands = map[string]*cmds.Command{
	"dns": DNSCmd,
	"name": {
		Subcommands: map[string]*cmds.Command{
			"resolve": name.IpnsCmd,
		},
	},
	"object": {
		Subcommands: map[string]*cmds.Command{
			"data":  ocmd.ObjectDataCmd,
			"links": ocmd.ObjectLinksCmd,
			"get":   ocmd.ObjectGetCmd,
			"stat":  ocmd.ObjectStatCmd,
		},
	},
	"resolve": ResolveCmd,
}

func init() {
	Root.ProcessHelp()
	*RootRO = *Root

	// this was in the big map definition above before,
	// but if we leave it there lgc.NewCommand will be executed
	// before the value is updated (:/sanitize readonly refs command/)

	// sanitize readonly version command (no need to expose precise deps)
	*VersionROCmd = *VersionCmd
	VersionROCmd.Subcommands = map[string]*cmds.Command{}
	rootROSubcommands["version"] = VersionROCmd

	Root.Subcommands = rootSubcommands
	RootRO.Subcommands = rootROSubcommands
}

type MessageOutput struct {
	Message string
}
