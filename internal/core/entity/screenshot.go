package entity

type ScreenshotRequest struct {
	URL      string `json:"url" binding:"required"`
	Device   string `json:"device"`
	FullPage bool   `json:"fullPage"`
}
