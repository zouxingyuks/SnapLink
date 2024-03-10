// Package types define the structure of request parameters and respond results in this package
package types

import (
	"time"
)

var _ time.Time

//todo 修正这些自动生成的部分
//
//// CreateTUserRequest request params
//type CreateTUserRequest struct {
//	Username     string    `json:"username" binding:""`
//	Password     string    `json:"password" binding:""`
//	RealName     string    `json:"realName" binding:""`
//	Phone        string    `json:"phone" binding:""`
//	Mail         string    `json:"mail" binding:""`
//	DeletionTime int64     `json:"deletionTime" binding:""`
//	CreateTime   time.Time `json:"createTime" binding:""`
//	UpdateTime   time.Time `json:"updateTime" binding:""`
//	DelFlag      int       `json:"delFlag" binding:""`
//}
//
//// UpdateTUserByIDRequest request params
//type UpdateTUserByIDRequest struct {
//	ID uint64 `json:"id" binding:""` // uint64 id
//
//	Username     string    `json:"username" binding:""`
//	Password     string    `json:"password" binding:""`
//	RealName     string    `json:"realName" binding:""`
//	Phone        string    `json:"phone" binding:""`
//	Mail         string    `json:"mail" binding:""`
//	DeletionTime int64     `json:"deletionTime" binding:""`
//	CreateTime   time.Time `json:"createTime" binding:""`
//	UpdateTime   time.Time `json:"updateTime" binding:""`
//	DelFlag      int       `json:"delFlag" binding:""`
//}
//
//// TUserObjDetail detail
//type TUserObjDetail struct {
//	ID string `json:"id"` // convert to string id
//
//	CreatedAt    time.Time `json:"createdAt"`
//	UpdatedAt    time.Time `json:"updatedAt"`
//	Username     string    `json:"username"`
//	Password     string    `json:"password"`
//	RealName     string    `json:"realName"`
//	Phone        string    `json:"phone"`
//	Mail         string    `json:"mail"`
//	DeletionTime int64     `json:"deletionTime"`
//	CreateTime   time.Time `json:"createTime"`
//	UpdateTime   time.Time `json:"updateTime"`
//	DelFlag      int       `json:"delFlag"`
//}
//
//// CreateTUserRespond only for api docs
//type CreateTUserRespond struct {
//	Code int    `json:"code"` // return code
//	Msg  string `json:"msg"`  // return information description
//	Data struct {
//		ID uint64 `json:"id"` // id
//	} `json:"data"` // return data
//}
//
//// UpdateTUserByIDRespond only for api docs
//type UpdateTUserByIDRespond struct {
//	Result
//}
//
//// GetTUserByIDRespond only for api docs
//type GetTUserByIDRespond struct {
//	Code int    `json:"code"` // return code
//	Msg  string `json:"msg"`  // return information description
//	Data struct {
//		TUser TUserObjDetail `json:"tUser"`
//	} `json:"data"` // return data
//}
//
//// DeleteTUserByIDRespond only for api docs
//type DeleteTUserByIDRespond struct {
//	Result
//}
//
//// DeleteTUsersByIDsRequest request params
//type DeleteTUsersByIDsRequest struct {
//	IDs []uint64 `json:"ids" binding:"min=1"` // id list
//}
//
//// DeleteTUsersByIDsRespond only for api docs
//type DeleteTUsersByIDsRespond struct {
//	Result
//}
//
//// GetTUserByConditionRequest request params
//type GetTUserByConditionRequest struct {
//	query.Conditions
//}
//
//// GetTUserByConditionRespond only for api docs
//type GetTUserByConditionRespond struct {
//	Code int    `json:"code"` // return code
//	Msg  string `json:"msg"`  // return information description
//	Data struct {
//		TUser TUserObjDetail `json:"tUser"`
//	} `json:"data"` // return data
//}
//
//// ListTUsersByIDsRequest request params
//type ListTUsersByIDsRequest struct {
//	IDs []uint64 `json:"ids" binding:"min=1"` // id list
//}
//
//// ListTUsersByIDsRespond only for api docs
//type ListTUsersByIDsRespond struct {
//	Code int    `json:"code"` // return code
//	Msg  string `json:"msg"`  // return information description
//	Data struct {
//		TUsers []TUserObjDetail `json:"tUsers"`
//	} `json:"data"` // return data
//}
//
//// ListTUsersRequest request params
//type ListTUsersRequest struct {
//	query.Params
//}
//
//// ListTUsersRespond only for api docs
//type ListTUsersRespond struct {
//	Code int    `json:"code"` // return code
//	Msg  string `json:"msg"`  // return information description
//	Data struct {
//		TUsers []TUserObjDetail `json:"tUsers"`
//	} `json:"data"` // return data
//}

// GetByUsernameDesensitizationRespond 脱敏数据返回
type GetByUsernameDesensitizationRespond struct {
	Username string `json:"username"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
	Mail     string `json:"mail"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username"  binding:"required,min=3,max=20"`
	Password string `json:"password"  binding:"required,min=6,max=50"`
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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
