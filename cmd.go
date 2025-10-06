package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/imroc/req/v3"
	"github.com/spf13/cobra"
)

const UA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"

var (
	globalClient = req.C().SetUserAgent(UA)

	enableDebug bool
	enableProxy bool
	bindAddr    string
	cfgPath     string
	rootCmd     = &cobra.Command{
		Use:     filepath.Base(os.Args[0]),
		Short:   "一个用于加速下载github和其他限速网站资源的代理服务",
		Version: VERSION,
		Run:     runServer,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if enableDebug {
				globalClient.EnableDebugLog()
				globalClient.EnableDumpEachRequest()
			}

			if !enableProxy {
				globalClient.SetProxy(nil)
			}

			globalClient.EnableHTTP3()
			globalClient.EnableInsecureSkipVerify()
			globalClient.SetRedirectPolicy(req.NoRedirectPolicy())
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&enableDebug, "debug", "d", false, "enable debug log level (default not enable)")
	rootCmd.PersistentFlags().BoolVarP(&enableProxy, "proxy", "p", false, "enable use proxy to fetch resource (default not enable)")
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "rule", "r", "", "rules file path (default name is "+RULE_FILE_NAME+")")
	rootCmd.PersistentFlags().StringVarP(&bindAddr, "bind", "b", ":3000", "tcp address that web server listen (default is :3000)")
}

func runServer(cmd *cobra.Command, args []string) {
	var err error

	if err = searchConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cfgPath == "" {
		fmt.Fprintf(os.Stderr, "Error: rule file not found\n\n")

		cmd.Help()
		os.Exit(1)
	}

	cfg, err = ParseConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	startWeb()
}
