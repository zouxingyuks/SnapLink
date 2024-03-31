package routers

import (
	"SnapLink/pkg/serialize"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"net/http"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		thirdRouter(group)
	})
}

// todo 优化第三方接口调用,防止恶意调用
func thirdRouter(group *gin.RouterGroup) {
	group = group.Group("/")
	group.GET("/title", func(c *gin.Context) {
		url := c.Query("url")

		// 请求URL
		resp, err := http.Get(url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get URL"})
			return
		}
		defer resp.Body.Close()

		// 解析HTML
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse HTML"})
			return
		}

		// 获取<title>标签的内容
		title := doc.Find("title").Text()

		serialize.NewResponse(200, serialize.WithData(title)).ToJSON(c)
	})
}
