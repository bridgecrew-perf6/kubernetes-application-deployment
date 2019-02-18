package utils

import (
	"github.com/urfave/cli"
	"log"

	"kubernetes-services-deployment/constants"
	"os"
)

func InitFlags() error {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port",
			Usage:       "port for the application. default:8089",
			Destination: &constants.ServicePort,
			EnvVar:      "PORT",
		},
		cli.StringFlag{
			Name:        "logging_engine_url",
			Usage:       "logging ip:port",
			Destination: &constants.LoggingURL,
			EnvVar:      "LOGGING_ENGINE_URL",
		},
		cli.StringFlag{
			Name:        "cluster_engine_url",
			Usage:       "cluster ip:port ",
			Destination: &constants.ClusterAPI,
			EnvVar:      "CLUSTER_ENGINE_URL",
		},
		cli.StringFlag{
			Name:        "kubernetes_engine_url",
			Usage:       "kubernetes ip:port ",
			Destination: &constants.KubernetesEngineURL,
			EnvVar:      "KUBERNETES_ENGINE_URL",
		},
		cli.StringFlag{
			Name:        "environment_engine_url",
			Usage:       "Environment ip:port ",
			Destination: &constants.EnvironmentEngineURL,
			EnvVar:      "ENVIRONMENT_ENGINE_URL",
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
