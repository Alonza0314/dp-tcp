package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Alonza0314/dp-tcp/client"
	"github.com/Alonza0314/dp-tcp/logger"
	"github.com/Alonza0314/dp-tcp/model"
	"github.com/Alonza0314/dp-tcp/util"
	loggergo "github.com/Alonza0314/logger-go/v2"
	loggergoUtil "github.com/Alonza0314/logger-go/v2/util"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:     "client",
	Short:   "Start the client",
	Example: "dp-tcp client",
	Run:     clientFunc,
}

func init() {
	clientCmd.Flags().StringP("config", "c", "config/client.yaml", "config file")
	if err := clientCmd.MarkFlagRequired("config"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(clientCmd)
}

func clientFunc(cmd *cobra.Command, args []string) {
	if os.Geteuid() != 0 {
		loggergo.Error("CLIENT", "This program requires root privileges")
		return
	}

	clientConfigFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}

	clientConfig := &model.ClientConfig{}
	if err := util.LoadFromYaml(clientConfigFilePath, clientConfig); err != nil {
		panic(err)
	}

	clientLogger := logger.NewClientLogger(loggergoUtil.LogLevelString(clientConfig.LoggerIE.Level), "", true)

	client := client.NewDpTcpClient(clientConfig, clientLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Start(ctx); err != nil {
		return
	}
	defer client.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	cancel()
}
