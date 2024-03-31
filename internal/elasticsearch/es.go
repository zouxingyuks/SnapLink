package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
	"io"
)

type ES struct {
	client *elasticsearch.Client
}

// NewES cfg 的数据量较大,因此选择使用传地址
func NewES(cfg *elasticsearch.Config) (*ES, error) {
	client, err := elasticsearch.NewClient(*cfg)
	if err != nil {
		return nil, err
	}

	es := &ES{
		client: client,
	}
	return es, nil
}

// ResponseParse 响应解析函数
type ResponseParse func(body io.ReadCloser) (map[string]any, error)

// Search 搜索 API 的请求方法
func (es ES) Search(ctx context.Context, index string, body any, parser ResponseParse) (map[string]any, error) {
	// 获取 ES 实例
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, errors.Wrap(err, "Error encoding query")
	}
	data, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex(index),
		es.client.Search.WithBody(buf),
		es.client.Search.WithPretty(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "查询ES失败")
	}
	defer data.Body.Close()

	if data.IsError() {
		return nil, decodeErrorResponse(data.Body, data.Status())
	}
	return parser(data.Body)
}

// decodeErrorResponse 解析错误响应
func decodeErrorResponse(body io.ReadCloser, status string) error {
	var e map[string]interface{}
	if err := json.NewDecoder(body).Decode(&e); err != nil {
		return errors.Wrap(err, "解析错误响应失败")
	}
	return errors.New(fmt.Sprintf("[%s] %s: %s", status, e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"]))
}
