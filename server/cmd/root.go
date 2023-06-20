package cmd

import (
	"context"
	"os"
	"strings"

	cbClient "cloud.google.com/go/cloudbuild/apiv1/v2"
	"cloud.google.com/go/compute/metadata"
	"github.com/AdamSLevy/flagbind"
	"github.com/chaitanyamaili/go-rpc-cloud-build/server/internal/build"
	"github.com/chaitanyamaili/go-rpc-cloud-build/server/internal/iam"
	"github.com/chaitanyamaili/go-rpc-cloud-build/server/internal/server"
	svc "github.com/chaitanyamaili/go-rpc-cloud-build/server/internal/server/v1"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/api/option"
)

// ServiceConfig holds the configuration for the service, needed to pass it to provision-client.
type ServiceConfig struct {
	Region    string `koanf:"region" env:"SERVICE_REGION" flag:"region;;Region of the service"`
	Name      string `koanf:"name" env:"SERVICE_NAME" flag:"name;;Name of the service"`
	Identity  string `koanf:"identity" env:"SERVICE_IDENTITY" flag:"identity;;Identity of the service"`
	ProjectID string `koanf:"projectid" env:"SERVICE_PROJECTID" flag:"projectid;;Project ID of the service"`
}

// AppConfig Application configuration.
type AppConfig struct {
	ListeningPort int           `koanf:"port" env:"PORT" flag:"port;8080;Port in which the service will run"`
	ListeningHost string        `koanf:"host" env:"HOST" flag:"host;0.0.0.0;Address in which the service will run"`
	Service       ServiceConfig `koanf:"service"`
	ClientImage   string        `koanf:"clientimage" env:"CLIENTIMAGE" flag:"clientimage;;Client image used in the cloud build."`
}

const (
	envPrefix = "RCB_"
	cmdName   = "rpc-cb-srv"
)

var (
	rootCmd = &cobra.Command{
		Use:     cmdName,
		Short:   "A lean implementation of a next-gen provisioning service.",
		PreRunE: setupConfig,
		RunE:    fireCommand,
	}
	cmdConfig     AppConfig
	koanfInstance *koanf.Koanf
	pFlagSet      *pflag.FlagSet
	// signalChan accepts OS messages.
	signalChan = make(chan os.Signal, 1)
)

// Execute returns the result of rootCmd.Execute.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	koanfInstance = koanf.New(".")

	pFlagSet = pflag.NewFlagSet(cmdName, pflag.ContinueOnError)
	cobra.CheckErr(flagbind.Bind(pFlagSet, &cmdConfig))

	rootCmd.PersistentFlags().AddFlagSet(pFlagSet)
	// TODO: Fix pflag parsing on cmdline
	cobra.CheckErr(rootCmd.MarkPersistentFlagRequired("port"))
}

// fireCommand would only reach if no validation errors are found on
// the init stage of the service.
func fireCommand(cmd *cobra.Command, args []string) error {
	// Read the config into the cmdConfig struct
	err := koanfInstance.Unmarshal("", &cmdConfig)
	if err != nil {
		return err
	}

	appLog := zerolog.New(os.Stderr)
	// Instantiate server and start it
	err = setupServer(cmd.Context(), &cmdConfig, &appLog)

	if err != nil {
		enrichedErr := errors.WithMessage(err, "Failed to instantiate HTTP server")
		return errors.WithStack(enrichedErr)
	}

	// If initialization goes well, return a nil error here
	return nil
}

// setupConfig should prepare and read configuration for our Application.
func setupConfig(cmd *cobra.Command, args []string) error {
	err := koanfInstance.Load(posflag.Provider(pFlagSet, ".", koanfInstance), nil)
	if err != nil {
		return err
	}

	err = koanfInstance.Load(env.ProviderWithValue(envPrefix, ".", func(k, v string) (string, interface{}) {
		var endVal interface{}

		endVal = v
		multiVal := strings.Split(v, " ")
		endKey := strings.Replace(
			strings.ToLower(
				strings.TrimPrefix(k, envPrefix),
			), "_", ".", -1)

		if len(multiVal) > 1 {
			endVal = multiVal
		}
		return endKey, endVal
	}), nil)
	if err != nil {
		return err
	}

	return nil
}

// setupServe should create a new server for decoupled-provision-service.
// This function is blocking, as the last instruction is a
// ListenAndServe.
func setupServer(ctx context.Context, appConfig *AppConfig, logger *zerolog.Logger) error {
	// Read structured config, already filled properly from koanf
	serverCfg := server.Config{
		Addr: appConfig.ListeningHost,
		Port: appConfig.ListeningPort,
	}

	var rawBuildClient *cbClient.Client
	var err error
	if !metadata.OnGCE() {
		// Create Raw Cloud Build API client
		cloudBuildScopes := []string{
			"https://www.googleapis.com/auth/cloud-platform",
		}
		ts, err := iam.GetImpersonateAccessToken(
			ctx,
			appConfig.Service.Identity,
			iam.RenewableToken,
			cloudBuildScopes...,
		)
		if err != nil {
			return err
		}

		rawBuildClient, err = cbClient.NewClient(ctx, option.WithTokenSource(ts))
		if err != nil {
			return err
		}
	} else {
		rawBuildClient, err = cbClient.NewClient(ctx)
		if err != nil {
			return err
		}
	}

	buildCfg := build.Config{
		Logger:            logger,
		UpperClient:       rawBuildClient,
		GcpBuilderProject: appConfig.Service.ProjectID,
		ClientImage:       appConfig.ClientImage,
	}
	buildHandler, err := build.NewBuilder(buildCfg)
	if err != nil {
		return err
	}

	// Service instantiation
	service := svc.NewService(logger, buildHandler)

	// Prepare new basic objects for our app
	// 1. An http server
	// 2. One or more gRPC services.
	serverInstance := server.NewServer(&serverCfg, logger, nil)
	serverInstance.RegisterGRPCService(service)

	// Return any error, if applies
	return serverInstance.ListenAndServe()
}
