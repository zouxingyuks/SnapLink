package shortLink

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	model "go-ssas/model/shortLink"
	"go-ssas/service/dao/db"
	"gorm.io/gorm"
	"log"
	"net/url"
	"time"
)

type Param struct {
	OriginURL   string `json:"origin_url"`
	GID         string `json:"gid"`
	ValidTime   string `json:"valid_time"`
	ValidType   string `json:"valid_type"`
	Description string `json:"description"`
}

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
	//0. 获取参数
	param := Param{}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		//todo log  的设置
		log.Println(errors.Wrap(err, " 参数绑定错误"))
		c.JSON(400, gin.H{
			"msg": "参数错误",
		})
		return
	}

	u, err := url.Parse(param.OriginURL)
	if err != nil {
		log.Println(errors.Wrap(err, "url格式错误"))
		c.JSON(400, gin.H{
			"msg": "url格式错误",
		})
		return
	}
	sLink := model.ShortLink{
		Model:       gorm.Model{},
		Clicks:      0,
		Enable:      false,
		Domain:      u.Host,
		OriginURL:   u.String(),
		Gid:         param.GID,
		Description: param.Description,
	}

	sLink.ValidTime, err = time.Parse("2006-01-02 15:04:05", param.ValidTime)
	if err != nil {
		log.Println(errors.Wrap(err, "时间格式错误"))
		c.JSON(400, gin.H{
			"msg": "时间格式错误,请使用 YYYY-MM-DD HH:mm:ss 格式",
		})
		return
	}

	//2. 生成hash
	sLink.URI = ShortLinkToHash(sLink.Domain, u.Path)
	//3. 保存到数据库
	// 对布隆过滤器误判的情况进行判断
	tdb := db.DB().Save(&sLink)
	if tdb.Error != nil {
		c.JSON(500, gin.H{
			"msg": "保存失败",
		})
		return
	}
	fullShortURL := makeFullShortURL(sLink.URI)
	//4. 返回短链接
	c.JSON(200, gin.H{
		"url": fullShortURL,
		"msg": "ok",
	})

}
