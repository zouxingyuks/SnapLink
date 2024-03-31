package model

const (
	// TUserShardingNum 用户表分表数量
	TUserShardingNum = 16
	// SLGroupShardingNum 短链接分组表分表数量
	SLGroupShardingNum = 16
	// ShortLinkShardingNum 短链接表分表数量
	ShortLinkShardingNum = 16
	// RedirectShardingNum 重定向信息表分表数量
	RedirectShardingNum = 16
	// LinkAccessRecordShardingNum 访问记录表分表数量
	LinkAccessRecordShardingNum = 16
	// LinkAccessStatisticShardingNum 访问统计表分表数量
	LinkAccessStatisticShardingNum = 16
)

const (
	//TUserPrefix TUser表前缀
	TUserPrefix = "t_user_"
	//SLGroupPrefix SLGroup表前缀
	SLGroupPrefix = "sl_group_"
	//ShortLinkPrefix ShortLink表前缀
	ShortLinkPrefix = "short_link_"
	//RedirectPrefix Redirect表前缀
	RedirectPrefix = "redirect_"
	//LinkAccessRecordPrefix LinkAccessRecord表前缀
	LinkAccessRecordPrefix = "link_access_record_"
	//LinkAccessStatisticPrefix LinkAccessStatistic表前缀
	LinkAccessStatisticPrefix = "link_access_statistic_"
)
