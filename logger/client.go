package logger

import (
	"github.com/Alonza0314/dp-tcp/constant"
	loggergo "github.com/Alonza0314/logger-go/v2"
	loggergoModel "github.com/Alonza0314/logger-go/v2/model"
	loggergoUtil "github.com/Alonza0314/logger-go/v2/util"
)

type ClientLogger struct {
	*loggergo.Logger

	CfgLog    loggergoModel.LoggerInterface
	ClientLog loggergoModel.LoggerInterface
	Tcp1Log   loggergoModel.LoggerInterface
	Tcp2Log   loggergoModel.LoggerInterface
	TunLog    loggergoModel.LoggerInterface
}

func NewClientLogger(level loggergoUtil.LogLevelString, filePath string, debugMode bool) *ClientLogger {
	logger := loggergo.NewLogger(filePath, debugMode)
	logger.SetLevel(level)

	return &ClientLogger{
		Logger:    logger,
		CfgLog:    logger.WithTags(constant.CLIENT_TAG, constant.CONFIG_TAG),
		ClientLog: logger.WithTags(constant.CLIENT_TAG, constant.CLIENT_TAG),
		Tcp1Log:   logger.WithTags(constant.CLIENT_TAG, constant.TCP_1_TAG),
		Tcp2Log:   logger.WithTags(constant.CLIENT_TAG, constant.TCP_2_TAG),
		TunLog:    logger.WithTags(constant.CLIENT_TAG, constant.TUN_TAG),
	}
}
