package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/model"
	"context"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

type LinkAccessStatisticDao interface {
	GetStatistic(ctx context.Context, uri string, startDatetime, endDatetime string, pageNum, pageSize uint64) ([]model.LinkAccessStatistic, error)
	GetRecord(ctx context.Context, uri string, startDatetime, endDatetime string, pageNum, pageSize uint64) ([]model.LinkAccessRecord, error)
	SaveToDB(ctx context.Context, uri string, date string, hour int) error
	GetStatisticByDay(ctx context.Context, uri string, date string, order string) (model.LinkAccessStatisticDay, error)
}
type LinkAccessStatisticHandler struct {
	iDao LinkAccessStatisticDao
}

func NewLinkAccessStatisticHandler() *LinkAccessStatisticHandler {
	h := &LinkAccessStatisticHandler{
		iDao: dao.NewLinkAccessStatisticDao(
			cache.NewLinkStatsCache(model.GetCacheType())),
	}
	return h
}

// GetStatistic 获取基础访问统计(PV,UV,UIP)
// @Summary 获取基础访问统计(PV,UV,UIP)
// @Description 获取单日单小时的访问统计
// @Tags LinkAccessStatistic
// @Accept json
// @Produce json
// @Param uri query string true "uri"
// @Param startDatetime query string true "开始日期,format:2006-01-02 15:04:05"
// @Param endDatetime query string false "结束日期,format:2006-01-02 15:04:05,默认为开始日期"
// @Param options query string false "查询选项,可选有 region,device,默认为全部"
// @Param pageNum query int true "页码"
// @Param pageSize query int true "每页数量"
// @Success 200 {object} model.LinkAccessStatistic
// @Router /linkAccessStatistic [get]
func (h *LinkAccessStatisticHandler) GetStatistic(c *gin.Context) {
	//参数获取与校验
	uri := c.Query("uri")
	startDatetime := c.Query("startDatetime")
	endDatetime := c.Query("endDatetime")
	page, _ := strconv.ParseUint(c.Query("pageNum"), 10, 64)
	pageSize, _ := strconv.ParseUint(c.Query("pageSize"), 10, 64)
	if uri == "" || startDatetime == "" || page == 0 || pageSize == 0 {
		//todo 日志设计
		c.JSON(400, gin.H{"error": "uri,startDatetime,page and pageSize is required"})
		return
	}
	if endDatetime == "" {
		endDatetime = startDatetime
	}
	//获取数据
	stats, err := h.iDao.GetStatistic(c, uri, startDatetime, endDatetime, page, pageSize)
	//查询时间范围内的数据
	if err != nil {
		c.JSON(500, gin.H{"error": err})
	}

	//返回数据
	c.JSON(200, stats)
}

// GetRecords 获取单次访问详情
// @Summary 获取单次访问详情
// @Description 获取单次访问详情
// @Tags LinkAccessStatistic
// @Accept json
// @Produce json
// @Param uri query string true "uri"
// @Param startDatetime query string true "开始日期,format:2006-01-02 15:04:05"
// @Param endDatetime query string false "结束日期,format:2006-01-02 15:04:05,默认为开始日期"
// @Param pageNum query int true "页码"
// @Param pageSize query int true "每页数量"
// @Success 200 {object} model.LinkAccessRecord
// @Router /linkAccessStatistic/detail [get]
func (h *LinkAccessStatisticHandler) GetRecords(c *gin.Context) {
	//参数获取与校验
	uri := c.Query("uri")
	startDatetime := c.Query("startDatetime")
	endDatetime := c.Query("endDatetime")
	page, _ := strconv.ParseUint(c.Query("pageNum"), 10, 64)
	pageSize, _ := strconv.ParseUint(c.Query("pageSize"), 10, 64)
	if uri == "" || startDatetime == "" || page == 0 || pageSize == 0 {
		//todo 日志设计
		c.JSON(400, gin.H{"error": "uri,startDatetime,page and pageSize is required"})
		return
	}
	if endDatetime == "" {
		endDatetime = startDatetime
	}
	//获取数据
	record, err := h.iDao.GetRecord(c, uri, startDatetime, endDatetime, page, pageSize)
	if err != nil {
		//todo 日志设计
		c.JSON(500, gin.H{"error": err})
		return
	}
	//todo 查询数据
	//返回数据
	c.JSON(200, record)
}

// RefreshStatistic 立刻更新最新的访问统计数据
// @Summary 立刻更新最新的访问统计数据
// @Description 立刻更新最新的访问统计数据
// @Tags LinkAccessStatistic
// @Accept json
// @Produce json
// @Param uri query string true "uri"
// @Success 200 {string} string "success"
// @Router /linkAccessStatistic/update [post]
func (h *LinkAccessStatisticHandler) RefreshStatistic(c *gin.Context) {
	//参数获取与校验
	uri := c.Query("uri")
	if uri == "" {
		//todo 日志设计
		c.JSON(400, gin.H{"error": "uri is required"})
		return
	}
	date := time.Now().Format("2006-01-02")
	hour := time.Now().Hour()
	err := h.iDao.SaveToDB(c, uri, date, hour)
	if err != nil {
		//todo 日志设计
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	//返回数据
	c.JSON(200, "success")
}

// GetStatisticByDay
// @Summary 获取单日的访问统计
// @Description 获取单日的访问统计
// @Tags LinkAccessStatistic
// @Accept json
// @Produce json
// @Param uri query string true "uri"
// @Param date query string true "开始日期,format:2006-01-02 "
// @Param order query string false "排序方式,可选有 asc,desc,默认为desc"
// @Success 200 {object} model.LinkAccessStatisticDay
// @Router /linkAccessStatistic/day [get]
func (h *LinkAccessStatisticHandler) GetStatisticByDay(c *gin.Context) {
	uri := c.Query("uri")
	if uri == "" {
		c.JSON(400, gin.H{
			"err": "uri param err",
		})
		return
	}
	date, err := time.Parse("2006-01-02", c.Query("date"))
	if err != nil {
		c.JSON(400, gin.H{
			"err": "time param format err",
		})
		return
	}
	order := c.Query("order")
	if order == "" {
		order = "date desc"
	}
	data, err := h.iDao.GetStatisticByDay(c, uri, date.Format("2006-01-02"), order)
	if err != nil {
		c.JSON(500, gin.H{
			"err": err.Error(),
		})
		return
	}
	c.JSON(200, data)
}
