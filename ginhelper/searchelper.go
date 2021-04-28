package ginhelper

import (
	"github.com/gin-gonic/gin"
	"time"
)

var MAX_ONE_PAGE_SIZE = 20

type PageSearchReq struct {
	Page int `form:"page" binding:"required"`
	Size int `form:"size" binding:"required"`
}

type TimeSearchReq struct {
	StartT int64 `form:"start_t"`
	EndT   int64 `form:"end_t"`
}

// 获取分页查询参数
func GetPageAndSize(c *gin.Context) (page, size, skip int, err error) {
	paramsJSON := PageSearchReq{}
	err = c.ShouldBind(&paramsJSON)
	if err != nil {
		return
	}
	page = paramsJSON.Page
	size = paramsJSON.Size

	if page < 1 {
		page = 1
	}

	if size > MAX_ONE_PAGE_SIZE {
		size = MAX_ONE_PAGE_SIZE
	}

	skip = (page - 1) * size

	return
}

// 时间查询
func GetSTAndETTime(c *gin.Context) (startT int64, endT int64, err error) {
	paramsJSON := TimeSearchReq{}
	err = c.ShouldBind(&paramsJSON)
	if err != nil {
		return
	}
	et := time.Now().Unix()
	st := et - 60*60*24*30

	if paramsJSON.StartT <= 0 {
		startT = st
	} else {
		startT = paramsJSON.StartT
	}
	if paramsJSON.EndT <= startT {
		endT = et
	} else {
		endT = paramsJSON.EndT
	}
	return
}
