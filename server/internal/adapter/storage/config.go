package storage

import (
	"fmt"
	"net/url"

	"jww/utils/configuration"
)

// Config describes a private S3-compatible object store. It is
// purpose-agnostic: one connection (bucket + endpoint + credentials) can back
// many features, each namespacing its own objects under a folder in the key.
// Bucket and region are required; endpoint/credentials are optional so the same
// adapter targets AWS (default credential chain) or a local MinIO.
//
// No byte-size limit is configured anywhere: that is deliberately left to the
// caller/purpose.
type Config struct {
	Bucket          string `env:"OBJECT_STORAGE_S3_BUCKET" bind:"required"`
	Region          string `env:"OBJECT_STORAGE_S3_REGION" bind:"required"`
	Endpoint        string `env:"OBJECT_STORAGE_S3_ENDPOINT" bind:"optional" default:""`
	ForcePathStyle  bool   `env:"OBJECT_STORAGE_S3_FORCE_PATH_STYLE" bind:"optional" default:"false"`
	AccessKeyID     string `env:"OBJECT_STORAGE_S3_ACCESS_KEY_ID" bind:"optional" default:""`
	SecretAccessKey string `env:"OBJECT_STORAGE_S3_SECRET_ACCESS_KEY" bind:"optional" default:""`
	SessionToken    string `env:"OBJECT_STORAGE_S3_SESSION_TOKEN" bind:"optional" default:""`
}

func LoadConfig() (Config, error) {
	cfg, err := configuration.Load[Config]()
	if err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate fails fast on a misconfiguration that would otherwise surface only
// as an opaque runtime error on the first upload: a half-configured static
// credential pair, a malformed custom endpoint, or a blank bucket.
func (c Config) Validate() error {
	if (c.AccessKeyID == "") != (c.SecretAccessKey == "") {
		return fmt.Errorf("storage: OBJECT_STORAGE_S3_ACCESS_KEY_ID and OBJECT_STORAGE_S3_SECRET_ACCESS_KEY must be set together")
	}
	if c.Endpoint != "" {
		u, err := url.Parse(c.Endpoint)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("storage: OBJECT_STORAGE_S3_ENDPOINT must be an absolute URL, got %q", c.Endpoint)
		}
	}
	if c.Bucket == "" {
		return fmt.Errorf("storage: OBJECT_STORAGE_S3_BUCKET must not be blank")
	}
	return nil
}
