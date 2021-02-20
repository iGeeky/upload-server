package controller

import (
	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
	"github.com/iGeeky/open-account/pkg/baselib/utils"

	"github.com/gin-gonic/gin"
)

// APIPing ping测试接口
func APIPing(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	data := gin.H{"buildTime": utils.BuildTime, "GitBranch": utils.GitBranch, "GitCommit": utils.GitCommit, "now": utils.Datetime()}
	ctx.JsonOk(data)
}
