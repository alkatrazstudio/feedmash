// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package src

import (
	"feedmash/util"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

// use var instead of const so these values can be overriden.
// see example in build-unix.sh
var appId = "feedmash"
var appTitle = "FeedMash"
var appVersion = "0.0.0"
var appBuildTimestamp = "0"
var appBuildGitHash = ""
var copyleftAuthor = "Alexey Parfenov (a.k.a. ZXED)"
var copyleftAuthorEmail = "zxed@alkatrazstudio.net"
var copyleftLicense = "AGPLv3"
var appHomepage = "https://github.com/alkatrazstudio/feedmash"
var authorHomepage = "https://alkatrazstudio.net"

type Config struct {
	filename         string
	appId            string
	appTitle         string
	serverAddr       string
	outFeedFilename  string
	outFeedId        string
	outFeedTitle     string
	outFeedSelfLink  string
	sources          []string
	userAgent        string
	maxOutItems      int
	initialPauseSecs int
	minIntervalMins  int
	maxIntervalMins  int
}

func getString(v *viper.Viper, key string, def string) string {
	v.SetDefault(key, def)
	return v.GetString(key)
}

func getStringSlice(v *viper.Viper, key string, def []string) []string {
	v.SetDefault(key, def)
	return v.GetStringSlice(key)
}

func getInt(v *viper.Viper, key string, def int) int {
	v.SetDefault(key, def)
	return v.GetInt(key)
}

func dataRootDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	homeDir := usr.HomeDir

	switch osName := runtime.GOOS; osName {
	case "linux":
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome != "" {
			return xdgDataHome
		}
		return filepath.Join(homeDir, ".local", "share")

	case "windows":
		appDataDir := os.Getenv("APPDATA")
		if appDataDir != "" {
			return appDataDir
		}
		return filepath.Join(homeDir, "AppData", "Roaming")

	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support")

	default:
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		errStr := fmt.Sprintf("Unsupported OS: %s; Using config dir as data dir: %s", osName, configDir)
		_, _ = fmt.Fprintln(os.Stderr, errStr)
		return configDir
	}
}

func configFromFile(configFilename string) Config {
	file, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}

	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(file)
	if err != nil {
		panic(err)
	}

	outFeedFilename := getString(v, "outFeedFilename", "")
	if outFeedFilename == "" {
		dataDir := dataRootDir()
		outFeedFilename = filepath.Join(dataDir, appId, appId+".xml")
	}

	cfg := Config{
		filename:         configFilename,
		appId:            appId,
		appTitle:         appTitle,
		serverAddr:       getString(v, "serverAddr", "127.0.0.1:13742"),
		outFeedFilename:  outFeedFilename,
		outFeedTitle:     getString(v, "outFeedTitle", appTitle),
		sources:          getStringSlice(v, "sources", []string{}),
		userAgent:        getString(v, "userAgent", appTitle),
		maxOutItems:      getInt(v, "maxOutItems", 666),
		initialPauseSecs: getInt(v, "initialPauseSecs", 1),
		minIntervalMins:  getInt(v, "minIntervalMins", 3*60),
		maxIntervalMins:  getInt(v, "maxIntervalMins", 4*60),
	}

	defaultOutFeedSelfLink := "http://" + cfg.serverAddr + "/" + cfg.appId + ".xml"
	cfg.outFeedSelfLink = getString(v, "outFeedSelfLink", defaultOutFeedSelfLink)

	defaultOutFeedId := cfg.appId
	cfg.outFeedId = getString(v, "outFeedId", defaultOutFeedId)

	if len(cfg.sources) == 0 {
		panic(
			fmt.Sprintf(
				"No sources specified. Add sources to your config file (%s), in the \"sources\" array.",
				cfg.filename,
			),
		)
	}

	return cfg
}

func handleCli(callback func(Config), exampleYaml string) {
	var printExampleConfig = false

	ts, err := strconv.ParseInt(appBuildTimestamp, 10, 64)
	if err != nil {
		panic(err)
	}
	t := time.Unix(ts, 0)
	tStr := t.Format("January 02, 2006")

	versionStr := fmt.Sprintf("v%s (%s) [git hash: %s]", appVersion, tStr, appBuildGitHash)

	var rootCmd = &cobra.Command{
		Use: fmt.Sprintf("%s <config-file>", appId),
		Short: "Monitors multiple RSS/Atom/JSON feeds,\n" +
			"combines them into one Atom feed\n" +
			"and serves the resulting feed via HTTP.\n" +
			"\n" +
			"Project homepage: " + appHomepage + "\n" +
			"License: " + copyleftLicense + "\n" +
			"Build date: " + tStr + "\n" +
			"Git commit: " + appBuildGitHash + "\n" +
			"Author: " + copyleftAuthor + " <" + copyleftAuthorEmail + ">\n" +
			"Author's homepage: " + authorHomepage,
		Version:               versionStr,
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(_ *cobra.Command, args []string) {
			if printExampleConfig {
				util.LogInfo(exampleYaml)
				return
			}

			if len(args) == 0 {
				util.LogWarn("Specify a path to the config file. Use --help flag for more instructions.")
				return
			}

			cfgFilename := args[0]
			cfg := configFromFile(cfgFilename)
			callback(cfg)
		},
		ValidArgs: []string{"CFG_FILE"},
		Example: "  1) Get an example config (which also contains further instructions):\n" +
			"\n" +
			"    " + appId + " --print-example-config\n" +
			"\n" +
			"  2) Use that config to create your own config file and then pass it to FeedMash:\n" +
			"\n" +
			"    " + appId + " /path/to/your/config.yaml",
	}

	rootCmd.Flags().BoolVar(&printExampleConfig, "print-example-config", false, "print an example config file")

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
