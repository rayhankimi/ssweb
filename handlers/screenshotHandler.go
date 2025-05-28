package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ssweb/types"
	"ssweb/utils"
)

// Device presets from Chrome Dev tools
var devicePresets = map[string]struct {
	Width  int
	Height int
}{
	"iPhone SE":          {375, 667},
	"iPhone XR":          {414, 896},
	"iPhone 12 Pro":      {390, 844},
	"Pixel 5":            {393, 851},
	"Samsung Galaxy S21": {360, 800},
	"iPad Mini":          {768, 1024},
}

func ScreenshotHandler(c *gin.Context) {
	req := types.ScreenshotRequest{
		URL:      c.Query("url"),
		Device:   c.Query("device"),
		FullPage: c.Query("fullPage") == "true",
		Quality:  20,
	}

	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	imgBuf, err := utils.TakeScreenshot(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/png", imgBuf)
}
