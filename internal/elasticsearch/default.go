package elasticsearch

import (
	"SnapLink/internal/config"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"
)

var instance struct {
	es   *elasticsearch.Client
	once sync.Once
}

func Instance() *elasticsearch.Client {
	instance.once.Do(
		func() {
			cfg := elasticsearch.Config{
				Addresses:                config.Get().Elasticsearch.Addresses,
				Username:                 config.Get().Elasticsearch.Username,
				Password:                 config.Get().Elasticsearch.Password,
				CloudID:                  config.Get().Elasticsearch.CloudID,
				APIKey:                   config.Get().Elasticsearch.APIKey,
				ServiceToken:             config.Get().Elasticsearch.ServiceToken,
				CertificateFingerprint:   config.Get().Elasticsearch.CertificateFingerprint,
				Header:                   nil,
				CACert:                   nil,
				RetryOnStatus:            config.Get().Elasticsearch.RetryOnStatus,
				DisableRetry:             config.Get().Elasticsearch.DisableRetry,
				MaxRetries:               config.Get().Elasticsearch.MaxRetries,
				RetryOnError:             nil,
				CompressRequestBody:      config.Get().Elasticsearch.CompressRequestBody,
				CompressRequestBodyLevel: config.Get().Elasticsearch.CompressRequestBodyLevel,
				DiscoverNodesOnStart:     config.Get().Elasticsearch.DiscoverNodesOnStart,
				DiscoverNodesInterval:    config.Get().Elasticsearch.DiscoverNodesInterval,
				EnableMetrics:            config.Get().Elasticsearch.EnableMetrics,
				EnableDebugLogger:        config.Get().Elasticsearch.EnableDebugLogger,
				EnableCompatibilityMode:  config.Get().Elasticsearch.EnableCompatibilityMode,
				DisableMetaHeader:        config.Get().Elasticsearch.DisableMetaHeader,
				RetryBackoff:             nil,
				Transport:                nil,
				Logger:                   nil,
				Selector:                 nil,
				ConnectionPoolFunc:       nil,
				Instrumentation:          nil,
			}
			var err error
			instance.es, err = elasticsearch.NewClient(cfg)
			if err != nil {
				logger.Panic(errors.Wrap(err, "failed to create elasticsearch client").Error())
			}
		})
	return instance.es
}
