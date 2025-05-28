package types

type ScreenshotRequest struct {
	URL string `json:"url" binding:"required"`
	//Width    int    `json:"width"`
	//Height   int    `json:"height"`
	Device   string `json:"device"`
	FullPage bool   `json:"fullPage"`
	Quality  int    `json:"quality"`
}
