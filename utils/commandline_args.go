package utils

import (
	"github.com/urfave/cli"
	"log"

	"os"
	"kubernetes-services-deployment/constants"
)

func InitFlags() error {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port",
			Usage:       "port for the application. default:8088",
			Destination: &constants.ServicePort,
			EnvVar:      "PORT",
		},
		cli.StringFlag{
			Name:        "istio_engine_url",
			Usage:       "ip:port",
			Destination: &constants.IstioEngineURL,
			EnvVar:      "ISTIO_ENGINE_URL",
		},
		cli.StringFlag{
			Name:        "logging_engine_url",
			Usage:       "logging ip:port",
			Destination: &constants.LoggingURL,
			EnvVar:      "LOGGING_ENGINE_URL",
		},
		cli.StringFlag{
			Name:        "knative_engine_url",
			Usage:       "knative ip:port ",
			Destination: &constants.KnativeEngineURL,
			EnvVar:      "KNATIVE_ENGINE_URL",
		},
	}
	app.Action = func(c *cli.Context) error {
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
