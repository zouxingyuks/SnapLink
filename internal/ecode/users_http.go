package ecode

import "github.com/pkg/errors"

type ErrCode struct {
	Code  int
	ECode string
	Err   error
}

func newErrCode(code int, ecode, err string) ErrCode {
	return ErrCode{
		Code:  code,
		ECode: ecode,
		Err:   errors.New(err),
	}
}
func (e ErrCode) ToString() string {
	return e.ECode + " " + e.Err.Error()
}

var (
	//  ========== 一级宏观错误码 客户端错误 ==========
	ClientError = newErrCode(400, "A0001", "客户端错误")

	// ========== 二级宏观错误码 用户注册错误 ==========
	UserRegisterError             = newErrCode(400, "A000100", "用户注册错误")
	UserNameVerifyError           = newErrCode(400, "A000110", "用户名校验失败")
	UserNameExistError            = newErrCode(409, "A000111", "用户名已存在") // 409 Conflict 更适合表示资源冲突，如用户名已存在
	UserNameSensitiveError        = newErrCode(400, "A000112", "用户名包含敏感词")
	UserNameSpecialCharacterError = newErrCode(400, "A000113", "用户名包含特殊字符")
	PasswordVerifyError           = newErrCode(400, "A000120", "密码校验失败")
	PasswordShortError            = newErrCode(400, "A000121", "密码长度不够")
	PhoneVerifyError              = newErrCode(400, "A000151", "手机格式校验失败")

	// ========== 二级宏观错误码 系统请求缺少幂等Token ==========
	IdempotentTokenNullError   = newErrCode(401, "A000200", "幂等Token为空") // 缺少必要的请求参数，400 Bad Request 更适合
	IdempotentTokenDeleteError = newErrCode(401, "A000201", "幂等Token已被使用或失效")

	// ========== 二级宏观错误码 用户登录错误 ==========
	UserLoginError    = newErrCode(401, "A000300", "用户登录错误")
	UserNotExistError = newErrCode(401, "A000301", "用户不存在")
	UserPasswordError = newErrCode(401, "A000302", "密码错误")

	// ========== 一级宏观错误码 系统执行出错 ==========
	ServiceError = newErrCode(500, "B000001", "系统执行出错") // 500 Internal Server Error 表示服务器内部错误

	// ========== 二级宏观错误码 系统执行超时 ==========
	ServiceTimeoutError = newErrCode(504, "B000100", "系统执行超时") // 504 Gateway Timeout 更适合表示服务器处理超时

	// ========== 一级宏观错误码 调用第三方服务出错 ==========
	RemoteError = newErrCode(502, "C000001", "调用第三方服务出错") // 502 Bad Gateway 用于表示作为网关或代理的服务器，从上游服务器收到无效响应
)
