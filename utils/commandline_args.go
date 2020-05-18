package utils

import (
	"github.com/urfave/cli/v2"
	"log"

	"bitbucket.org/cloudplex-devs/kubernetes-services-deployment/constants"
	"os"
)

func InitFlags() error {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "port",
			Usage:       "port for the application. default:8089",
			Destination: &constants.ServicePort,
			EnvVars:     []string{"PORT"},
		},
		&cli.StringFlag{
			Name:        "logging_engine_url",
			Usage:       "logging ip:port",
			Destination: &constants.LoggingURL,
			EnvVars:     []string{"LOGGING_ENGINE_URL"},
		},
		&cli.StringFlag{
			Name:        "cluster_engine_url",
			Usage:       "cluster ip:port ",
			Destination: &constants.ClusterAPI,
			EnvVars:     []string{"CLUSTER_ENGINE_URL"},
		},
		&cli.StringFlag{
			Name:        "kubernetes_engine_url",
			Usage:       "kubernetes ip:port ",
			Destination: &constants.KubernetesEngineURL,
			EnvVars:     []string{"KUBERNETES_ENGINE_URL"},
		},
		&cli.StringFlag{
			Name:        "environment_engine_url",
			Usage:       "Environment ip:port ",
			Destination: &constants.EnvironmentEngineURL,
			EnvVars:     []string{"ENVIRONMENT_ENGINE_URL"},
		},
		&cli.StringFlag{
			Name:        "robin_url",
			Usage:       "robin ip:port ",
			Destination: &constants.VaultURL,
			EnvVars:     []string{"ROBIN_URL"},
		},
		&cli.StringFlag{
			Name:        "rbac_url",
			EnvVars:     []string{"RBAC_URL"},
			Destination: &constants.RbacURL,
		},
		&cli.StringFlag{
			Name:        "woodpecker_url",
			EnvVars:     []string{"WOODPECKER_URL"},
			Destination: &constants.WoodpeckerURL,
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
