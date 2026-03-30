package controller

import (
	"time"
	"webssh/core"

	"github.com/gin-gonic/gin"
)

func S3List(c *gin.Context) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)

	if core.GlobalS3Manager == nil {
		responseBody.Msg = "S3 is not configured"
		return &responseBody
	}

	prefix := c.DefaultQuery("prefix", "")

	entries, err := core.GlobalS3Manager.ListObjects(prefix)
	if err != nil {
		responseBody.Msg = err.Error()
		return &responseBody
	}

	responseBody.Data = entries
	return &responseBody
}
