package shortLink

import (
	"crypto/sha1"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	model "go-ssas/model/shortLink"
	"go-ssas/service/dao/db"
	"gorm.io/gorm"
	"net/url"
	"time"
)

var hash = sha1.New()

// Create 创建短链接
// @Summary 创建短链接
// @Description 创建短链接
// @Tags 短链接
// @Accept application/json
// @Produce application/json
// @Param origin_url body string true "原始链接"
// @Param gid body string false "组id"
// @Param valid_time body string true "有效时间"
// @Param valid_type body int false "有效类型"
// @Param description body string false "描述"
// @Success 200 {string} string "{"code":200,"data":{},"msg":"ok"}"
// @Router /shortLink/ [post]
func Create(c *gin.Context) {
	sLink := model.ShortLink{
		Model:       gorm.Model{},
		Clicks:      0,
		Gid:         "",
		Enable:      false,
		CreateType:  "",
		ValidTime:   time.Time{},
		ValidType:   0,
		Description: "",
	}
	//0. 获取参数
	err := c.ShouldBind(&sLink)
	if sLink.OriginURL == "" {
		c.JSON(200, gin.H{
			"code": 400,
			"msg":  "原始链接不能为空",
		})
		return
	}

	//1. 分割原始链接
	u, err := url.Parse(sLink.OriginURL)

	if err != nil {
		c.JSON(200, gin.H{
			"code": 400,
			"msg":  "原始链接格式错误",
		})
		return
	}
	sLink.Domain = u.Host
	//2. 生成hash
	sLink.URI = ShortLinkToHash(sLink.Domain, u.Path)
	fmt.Println(sLink)
	//3. 保存到数据库
	db.DB().Save(&sLink)
	fullShortURL := makeFullShortURL(sLink.URI)
	//4. 返回短链接
	c.JSON(200, gin.H{
		"code": 200,
		"url":  fullShortURL,
		"msg":  "ok",
	})

}

// makeFullShortURL 生成完整的短链接
func makeFullShortURL(uri string) string {
	//此处配置从配置文件中获取
	u := url.URL{
		Scheme: "http",
		Host:   "anubis.cafe",
		Path:   uri,
	}
	return u.String()
}

// ShortLinkToHash 短链接转hash
func ShortLinkToHash(domain, shortLink string) string {
	// 生成 hash
	// 1. 生成 hash
	uri := GenerateHash(shortLink)
	for {
		// 同一域名下的短链接不能重复
		if existHash(domain, uri) {
			//todo 如何取降低冲突
			GenerateHash(shortLink)
		}
		break
	}
	return uri
}
func GenerateHash(uri string) string {
	return uuid.New().String()
}

// 检查 hash 是否存在
func existHash(Domain, hash string) bool {
	return false
}
