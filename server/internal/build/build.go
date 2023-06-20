package build

import (
	"context"
	"errors"
	"os"
	"path"
	"strings"
	"sync"
	"unicode"

	client "cloud.google.com/go/cloudbuild/apiv1/v2"
	buildProtos "cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	"github.com/chaitanyamaili/go-rpc-cloud-build/server/internal/stringops"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Handler to use the build functions.
type Handler struct {
	apiClient      *client.Client
	builderProject string
	log            *zerolog.Logger
	clientImage    string
}

// Config to hold the Handler's configuration.
type Config struct {
	Logger            *zerolog.Logger
	UpperClient       *client.Client
	GcpBuilderProject string
	ClientImage       string
}

// NewBuilder return the instance of handler to use the build functions.
func NewBuilder(config Config) (*Handler, error) {
	builderLogger := config.Logger.With().Str("component", "builder").Logger()
	if config.UpperClient == nil {
		return nil, errors.New("no build logger found")
	}

	return &Handler{
		apiClient:      config.UpperClient,
		builderProject: config.GcpBuilderProject,
		log:            &builderLogger,
		clientImage:    config.ClientImage,
	}, nil
}

// CreateNewBuild function to create a new Cloud Build job.
func (b *Handler) CreateNewBuild(ctx context.Context, newGcsName string, region string) (*buildProtos.BuildOperationMetadata, error) {
	// Setup the paths and file names
	configYAML, err := os.ReadFile(path.Join(
		"/assets",
		"creategcs-cloudbuild.yaml",
	))
	if err != nil {
		b.log.Error().
			Str("component-path", "CreateNewBuild").
			Err(err).
			Msg("couldn't read builder YAML file")
		return nil, err
	}

	buildObj, err := b.YAMLToProtoBuild(configYAML)
	if err != nil {
		b.log.Error().
			Str("component-path", "CreateNewBuild").
			Err(err).
			Msg("couldn't parse cloud build config YAML to protoBuild object")
		return nil, err
	}

	// Last substitution to define the new deployment_name
	buildObj.Substitutions["_BASE_IMAGE"] = "gcr.io/cloud-builders/gsutil"
	buildObj.Substitutions["_GCS_REGION"] = region
	buildObj.Substitutions["_GCS_NAME"] = newGcsName

	req := buildProtos.CreateBuildRequest{
		Parent: strings.Join([]string{
			"projects",
			b.builderProject,
			"locations",
			"global",
		}, "/"),
		ProjectId: b.builderProject,
		Build:     buildObj,
	}

	buildOperation, err := b.apiClient.CreateBuild(ctx, &req)
	if err != nil {
		b.log.Error().
			Str("component-path", "CreateNewBuild").
			Err(err).
			Msg("Cloud Build API error")
		return nil, err
	}

	_, err = buildOperation.Metadata()
	if err != nil {
		for {
			_, err = buildOperation.Poll(ctx)
			if err == nil {
				break
			}
		}
	}

	return buildOperation.Metadata()
}

// GenerateGCSName returns a valid siteID (UUID base36 encoded) to be used in external packages.
func GenerateGCSName(ctx context.Context) string {
	wg := sync.WaitGroup{}

	gcsChan := make(chan string)
	done := make(chan struct{})

	// Keep cycling if first encoded char is a digit
	// 3 goroutines should be enough
	// first routine to exit takes it all
	for i := 0; i < 3; i++ {
		wg.Add(1)

		go func(ctx context.Context, gcsChan chan<- string) {
			defer wg.Done()
			foundID := ""
			for {
				id := uuid.New()
				name, err := stringops.Encode(id[:], stringops.EncoderBase36)
				if err != nil {
					continue
				}

				// Ensure string is in lowercase. We can convert it back to Uppercase to Decode from base36
				name = strings.ToLower(name)
				if !unicode.IsDigit([]rune(name)[0]) {
					foundID = name
					break
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			// Signal all goroutines to EOL
			case gcsChan <- foundID:
				close(done)
			}
		}(ctx, gcsChan)
	}

	go func() {
		wg.Wait()
		close(gcsChan)
	}()

	// Read output data
	gotID := <-gcsChan

	return gotID
}
