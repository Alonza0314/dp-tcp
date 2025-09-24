package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Alonza0314/dp-tcp/logger"
	"github.com/Alonza0314/dp-tcp/model"
	"github.com/Alonza0314/dp-tcp/server"
	"github.com/Alonza0314/dp-tcp/util"
	loggergo "github.com/Alonza0314/logger-go/v2"
	loggergoUtil "github.com/Alonza0314/logger-go/v2/util"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Start the server",
	Example: "dp-tcp server",
	Run:     serverFunc,
}

func init() {
	serverCmd.Flags().StringP("config", "c", "config/server.yaml", "config file")
	if err := serverCmd.MarkFlagRequired("config"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(serverCmd)
}

func serverFunc(cmd *cobra.Command, args []string) {
	if os.Geteuid() != 0 {
		loggergo.Error("SERVER", "This program requires root privileges")
		return
	}

	serverConfigFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}

	serverConfig := &model.ServerConfig{}
	if err := util.LoadFromYaml(serverConfigFilePath, serverConfig); err != nil {
		panic(err)
	}

	serverLogger := logger.NewServerLogger(loggergoUtil.LogLevelString(serverConfig.LoggerIE.Level), "", true)

	server := server.NewDpTcpServer(serverConfig, serverLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		return
	}
	defer server.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	cancel()
}
