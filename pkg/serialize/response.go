package serialize

import (
	"SnapLink/internal/config"
	"SnapLink/internal/ecode"
	"github.com/gin-gonic/gin"
)

// Response 基础序列化器
type Response struct {
	Code  int         `json:"-"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg,omitempty"`
	Error string      `json:"error,omitempty"`
}

func (r Response) ToJSON(c *gin.Context) {
	c.JSON(r.Code, r)
}
func NewResponse(Code int, opts ...ResponseOption) Response {
	o := new(Response)
	//在此处设置默认值
	for _, opt := range opts {
		opt.apply(o)
	}
	o.Code = Code
	return *o
}
func NewResponseWithErrCode(errcode ecode.ErrCode, opts ...ResponseOption) Response {
	o := new(Response)
	//在此处设置默认值
	for _, opt := range opts {
		opt.apply(o)
	}
	o.Code = errcode.Code
	o.Msg = errcode.ToString()
	return *o
}

// ResponseOption 定义一个接口类型
type ResponseOption interface {
	apply(*Response)
}

// funcOption 定义funcOption类型，实现 IOption 接口
type funcOption struct {
	f func(*Response)
}

func newFuncOption(f func(option *Response)) ResponseOption {
	return &funcOption{
		f: f,
	}
}
func (fo funcOption) apply(o *Response) {
	fo.f(o)
}

// WithData 定义一个函数，用于设置 Data
func WithData(data interface{}) ResponseOption {
	return newFuncOption(func(o *Response) {
		o.Data = data
	})
}

// WithErr 将应用error携带标准库中的error
func WithErr(err error) ResponseOption {
	return newFuncOption(func(o *Response) {
		// 生产环境隐藏底层报错
		if err != nil && config.Get().App.Env == config.EnvProd {
			o.Error = err.Error()
		}
	})
}

func WithMsg(Msg string) ResponseOption {
	return newFuncOption(func(o *Response) {
		o.Msg = Msg
	})
}
