package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ssweb/types"
	"ssweb/utils"
)

func ScreenshotHandler(c *gin.Context) {
	var req types.ScreenshotRequest

	// Bind JSON body ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body! : " + err.Error(),
		})
		return
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL parameter is required",
		})
		return
	}

	imgBuf, err := utils.TakeScreenshot(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", imgBuf)
}
