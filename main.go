package main

import (
	"embed"

	"github.com/luiccn/espresso-controller/cmd/espresso/cmdutil"
	"github.com/luiccn/espresso-controller/cmd/espresso/config"
	"github.com/luiccn/espresso-controller/cmd/espresso/log"
	"github.com/luiccn/espresso-controller/internal/espresso"
	serverLogger "github.com/luiccn/espresso-controller/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configKeys = []config.Key{
	{Path: "Port", ShortFlag: "p", Description: "Port on which the espresso server should listen", Default: "8080"},
	{Path: "HeatingElementRelayPin", ShortFlag: "r", Description: "The GPIO connected to the heating element relay", Default: 14},
	{Path: "PowerButtonPin", ShortFlag: "", Description: "The GPIO connected to the power button of the espresso machine", Default: 17},
	{Path: "PowerButtonRelayPin", ShortFlag: "", Description: "The GPIO connected to the power button relay", Default: 21},
	{Path: "PowerLedPin", ShortFlag: "", Description: "The GPIO connected to the power LED", Default: 16},
	{Path: "BoilerThermCsPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31865 chip select, aka chip enable", Default: 5},
	{Path: "BoilerThermClkPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31865 clock", Default: 11},
	{Path: "BoilerThermMisoPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31865 data output", Default: 9},
	{Path: "BoilerThermMosiPin", ShortFlag: "", Description: "The GPIO pin connected to the boiler thermometer's max31865 data input", Default: 10},
	{Path: "Verbose", ShortFlag: "v", Description: "verbose output", Default: false},
}

//go:embed ui/build/*
var buildFiles embed.FS

func newRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "espresso",
		Short: "Control and monitor an espresso machine",
		Long:  "Control and monitor an espresso machine",
		PreRun: func(cmd *cobra.Command, args []string) {
			// Bind config in PreRun() to avoid collisions with other commands' flags
			for _, k := range configKeys {
				if err := viper.BindPFlag(k.Path, cmd.Flags().Lookup(k.Flag())); err != nil {
					log.Fatal("Failed to bind flag to config: %+v", k)
				}
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info(cmdutil.Logo)
			log.Info("For more information, go to https://github.com/luiccn/espresso-controller\n")

			if verbose := viper.GetBool("Verbose"); verbose {
				serverLogger.UseDevLogger()
			} else {
				serverLogger.UseProdLogger(
					viper.GetString(config.KeyLogFilePath),
					viper.GetInt(config.KeyLogFileMaxSize),
					viper.GetInt(config.KeyLogFileMaxAge),
					viper.GetInt(config.KeyLogFileMaxBackups),
				)
			}

			c := espresso.Configuration{}
			if err := viper.Unmarshal(&c); err != nil {
				log.Fatal("Unmarshalling configuration: %s\n", err.Error())
			}

			server := espresso.New(c, buildFiles)
			return server.Run()
		},
	}

	for _, k := range configKeys {
		if k.Default != nil {
			viper.SetDefault(k.Path, k.Default)
		}
		k.BindFlag(&cmd)
	}
	for _, k := range configKeys {
		viper.BindEnv(k.Path, k.EnvKey())
	}
	return &cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
