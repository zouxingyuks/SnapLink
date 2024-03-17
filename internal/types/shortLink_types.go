package types

// CreateShortLinkRequest 创建短链接请求参数
type CreateShortLinkRequest struct {
	OriginUrl string `json:"originUrl" binding:"required"`
	Gid       string `json:"gid" binding:"required"`
	// 0 为 api 创建,1 为控制台创建
	CreatedType int    `json:"createdType"`
	ValidDate   string `json:"validDate"`
	// 0 为 永不过期,1 为指定时间过期
	ValidDateType int    `json:"validDateType"`
	Description   string `json:"describe" binding:"required"`
}

// ShortLinkRecord 短链接详情
type ShortLinkRecord struct {
	CreatedAt     string `json:"createTime"`
	OriginUrl     string `json:"originUrl"`
	ShortUrl      string `json:"shortUrl"`
	ValidDateType int    `json:"validDateType"`
	ValidDate     string `json:"validDate"`
	Describe      string `json:"describe"`
}

// ListShortLinkResponse 短链接列表响应
type ListShortLinkResponse struct {
	Total    int                `json:"total"`
	Size     int                `json:"size"`
	Current  int                `json:"current"`
	OrderTag string             `json:"orderTag"`
	Records  []*ShortLinkRecord `json:"records"`
}
