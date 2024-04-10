package dao

import (
	"SnapLink/internal/model"
	"golang.org/x/sync/singleflight"
)

var fixDaoInstance = new(fixDao)

func FixDao() *fixDao {
	fixDaoInstance.once.Do(func() {
		fixDaoInstance.db = model.GetDB()
		fixDaoInstance.sfg = new(singleflight.Group)
	})
	return fixDaoInstance
}

var shortLinkInstance = new(shortLinkDao)

func ShortLinkDao() *shortLinkDao {
	shortLinkInstance.once.Do(func() {
		shortLinkInstance.db = model.GetDB()
		shortLinkInstance.sfg = new(singleflight.Group)
	})
	return shortLinkInstance
}

var redirectInstance = new(redirectsDao)

func RedirectDao() *redirectsDao {
	redirectInstance.once.Do(func() {
		redirectInstance.db = model.GetDB()
		shortLinkInstance.sfg = new(singleflight.Group)
	})
	return redirectInstance
}

var tUserDaoInstance = new(tUserDao)

func TUserDao() *tUserDao {
	tUserDaoInstance.once.Do(func() {
		tUserDaoInstance.db = model.GetDB()
		tUserDaoInstance.sfg = new(singleflight.Group)
	})
	return tUserDaoInstance

}
