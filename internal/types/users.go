package types

// GetByUsernameDesensitizationRespond 脱敏数据返回
type GetByUsernameDesensitizationRespond struct {
	Username string `json:"username"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Mail     string `json:"mail"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username"  binding:"required,min=3,max=11"`
	Password string `json:"password"  binding:"required,min=6,max=15"`
	RealName string `json:"realName"  binding:"required"`
	Phone    string `json:"phone"  binding:"required,e164"`
	Mail     string `json:"mail"  binding:"required,email"`
}

// RegisterRespond 用户注册返回
type RegisterRespond struct {
	Username string `json:"username"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Mail     string `json:"mail"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=11"`
	Password string `json:"password" binding:"required,min=6,max=15"`
}

// UpdateInfoRequest 用户修改请求
type UpdateInfoRequest struct {
	Password string `json:"password" binding:""`
	RealName string `json:"realName" binding:""`
	Phone    string `json:"phone" binding:""`
	Mail     string `json:"mail" binding:""`
}

// UpdateInfoRespond 用户修改返回
type UpdateInfoRespond struct {
	Username string `json:"username"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Mail     string `json:"mail"`
}
