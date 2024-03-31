package elasticsearch

import (
	"SnapLink/internal/config"
	"context"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"
)

var instance struct {
	es   *ES
	once sync.Once
}

func esInstance() *ES {
	instance.once.Do(
		func() {
			esConfig := config.Get().Elasticsearch
			cfg := elasticsearch.Config{
				Addresses:                esConfig.Addresses,
				Username:                 esConfig.Username,
				Password:                 esConfig.Password,
				CloudID:                  esConfig.CloudID,
				APIKey:                   esConfig.APIKey,
				ServiceToken:             esConfig.ServiceToken,
				CertificateFingerprint:   esConfig.CertificateFingerprint,
				Header:                   nil,
				CACert:                   nil,
				RetryOnStatus:            esConfig.RetryOnStatus,
				DisableRetry:             esConfig.DisableRetry,
				MaxRetries:               esConfig.MaxRetries,
				RetryOnError:             nil,
				CompressRequestBody:      esConfig.CompressRequestBody,
				CompressRequestBodyLevel: esConfig.CompressRequestBodyLevel,
				DiscoverNodesOnStart:     esConfig.DiscoverNodesOnStart,
				DiscoverNodesInterval:    esConfig.DiscoverNodesInterval,
				EnableMetrics:            esConfig.EnableMetrics,
				EnableDebugLogger:        esConfig.EnableDebugLogger,
				EnableCompatibilityMode:  esConfig.EnableCompatibilityMode,
				DisableMetaHeader:        esConfig.DisableMetaHeader,
				RetryBackoff:             nil,
				Transport:                nil,
				Logger:                   nil,
				Selector:                 nil,
				ConnectionPoolFunc:       nil,
				Instrumentation:          nil,
			}
			var err error
			instance.es, err = NewES(&cfg)
			if err != nil {
				logger.Panic(errors.Wrap(err, "init default es failed").Error())
			}
		})
	return instance.es
}

// Search ES 搜索 API
func Search(ctx context.Context, index string, body any, parser ResponseParse) (map[string]any, error) {
	return esInstance().Search(ctx, index, body, parser)
}
