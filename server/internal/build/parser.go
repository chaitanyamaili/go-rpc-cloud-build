package build

import (
	"encoding/json"

	buildProtos "cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	"google.golang.org/protobuf/encoding/protojson"
	yamlv3 "gopkg.in/yaml.v3"
)

// YAMLToProtoBuild reads from a cloudbuild.yaml file as bytes and parses it to a *buildProtos.Build struct.
func (b *Handler) YAMLToProtoBuild(configYAML []byte) (*buildProtos.Build, error) {
	var yamlRaw interface{}
	err := yamlv3.Unmarshal(configYAML, &yamlRaw)
	if err != nil {
		b.log.Error().
			Str("component-path", "YAMLToProtoBuild").
			Err(err).
			Msg("couldn't unmarshal cloudbuild yaml file to YAML object")
		return nil, err
	}

	configJSON, err := json.Marshal(yamlRaw)
	if err != nil {
		b.log.Error().
			Str("component-path", "YAMLToProtoBuild").
			Err(err).
			Msg("couldn't convert YAML to JSON object")
		return nil, err
	}

	buildObj := new(buildProtos.Build)

	err = protojson.Unmarshal(configJSON, buildObj)
	if err != nil {
		b.log.Error().
			Str("component-path", "YAMLToProtoBuild").
			Err(err).
			Msg("Can't unmarshal JSON object to Build proto")
		return nil, err
	}

	return buildObj, nil
}
