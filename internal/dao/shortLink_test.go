package dao

import (
	"context"
	"testing"
	"time"

	"SnapLink/internal/cache"
	"SnapLink/internal/model"

	"github.com/zhufuyi/sponge/pkg/gotest"
	"github.com/zhufuyi/sponge/pkg/mysql/query"
	"github.com/zhufuyi/sponge/pkg/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func newShortLinkDao() *gotest.Dao {
	testData := &model.ShortLink{}
	testData.ID = 1
	testData.CreatedAt = time.Now()
	testData.UpdatedAt = testData.CreatedAt

	// init mock cache
	//c := gotest.NewCache(map[string]interface{}{"no cache": testData}) // to test mysql, disable caching
	c := gotest.NewCache(map[string]interface{}{utils.Uint64ToStr(uint64(testData.ID)): testData})
	c.ICache = cache.NewShortLinkCache(&model.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})

	// init mock dao
	d := gotest.NewDao(c, testData)
	d.IDao = NewShortLinkDao(c.ICache.(cache.ShortLinkCache))

	return d
}

func Test_shortLinkDao_Create(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ShortLinkDao).Create(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_shortLinkDao_DeleteByID(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	testData.DeletedAt = gorm.DeletedAt{
		Time:  time.Now(),
		Valid: false,
	}

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ShortLinkDao).DeleteByID(d.Ctx, uint64(testData.ID))
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(ShortLinkDao).DeleteByID(d.Ctx, 0)
	assert.Error(t, err)
}

func Test_shortLinkDao_DeleteByIDs(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	testData.DeletedAt = gorm.DeletedAt{
		Time:  time.Now(),
		Valid: false,
	}

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ShortLinkDao).DeleteByID(d.Ctx, uint64(testData.ID))
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(ShortLinkDao).DeleteByIDs(d.Ctx, []uint64{0})
	assert.Error(t, err)
}

func Test_shortLinkDao_UpdateByID(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ShortLinkDao).UpdateByID(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(ShortLinkDao).UpdateByID(d.Ctx, &model.ShortLink{})
	assert.Error(t, err)

}

func Test_shortLinkDao_GetByID(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(testData.ID, testData.CreatedAt, testData.UpdatedAt)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID).
		WillReturnRows(rows)

	_, err := d.IDao.(ShortLinkDao).GetByID(d.Ctx, uint64(testData.ID))
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// notfound error
	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(2).
		WillReturnRows(rows)
	_, err = d.IDao.(ShortLinkDao).GetByID(d.Ctx, 2)
	assert.Error(t, err)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(3, 4).
		WillReturnRows(rows)
	_, err = d.IDao.(ShortLinkDao).GetByID(d.Ctx, 4)
	assert.Error(t, err)
}

func Test_shortLinkDao_GetByCondition(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(testData.ID, testData.CreatedAt, testData.UpdatedAt)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID).
		WillReturnRows(rows)

	_, err := d.IDao.(ShortLinkDao).GetByCondition(d.Ctx, &query.Conditions{
		Columns: []query.Column{
			{
				Name:  "id",
				Value: testData.ID,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// notfound error
	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(2).
		WillReturnRows(rows)
	_, err = d.IDao.(ShortLinkDao).GetByCondition(d.Ctx, &query.Conditions{
		Columns: []query.Column{
			{
				Name:  "id",
				Value: 2,
			},
		},
	})
	assert.Error(t, err)
}

func Test_shortLinkDao_GetByIDs(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(testData.ID, testData.CreatedAt, testData.UpdatedAt)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID).
		WillReturnRows(rows)

	_, err := d.IDao.(ShortLinkDao).GetByIDs(d.Ctx, []uint64{uint64(testData.ID)})
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.IDao.(ShortLinkDao).GetByIDs(d.Ctx, []uint64{111})
	assert.Error(t, err)

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_shortLinkDao_GetByLastID(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(testData.ID, testData.CreatedAt, testData.UpdatedAt)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, err := d.IDao.(ShortLinkDao).GetByLastID(d.Ctx, 0, 10, "")
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, err = d.IDao.(ShortLinkDao).GetByLastID(d.Ctx, 0, 10, "unknown-column")
	assert.Error(t, err)
}

func Test_shortLinkDao_GetByColumns(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(testData.ID, testData.CreatedAt, testData.UpdatedAt)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, _, err := d.IDao.(ShortLinkDao).GetByColumns(d.Ctx, &query.Params{
		Page: 0,
		Size: 10,
		Sort: "ignore count", // ignore test count(*)
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, _, err = d.IDao.(ShortLinkDao).GetByColumns(d.Ctx, &query.Params{
		Page: 0,
		Size: 10,
		Columns: []query.Column{
			{
				Name:  "id",
				Exp:   "<",
				Value: 0,
			},
		},
	})
	assert.Error(t, err)

	// error test
	dao := &shortLinkDao{}
	_, _, err = dao.GetByColumns(context.Background(), &query.Params{Columns: []query.Column{{}}})
	t.Log(err)
}

func Test_shortLinkDao_CreateByTx(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	_, err := d.IDao.(ShortLinkDao).CreateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_shortLinkDao_DeleteByTx(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	testData.DeletedAt = gorm.DeletedAt{
		Time:  time.Now(),
		Valid: false,
	}

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ShortLinkDao).DeleteByTx(d.Ctx, d.DB, uint64(testData.ID))
	if err != nil {
		t.Fatal(err)
	}
}

func Test_shortLinkDao_UpdateByTx(t *testing.T) {
	d := newShortLinkDao()
	defer d.Close()
	testData := d.TestData.(*model.ShortLink)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ShortLinkDao).UpdateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}
