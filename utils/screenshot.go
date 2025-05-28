package utils

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"ssweb/types"
	"time"
)

func TakeScreenshot(req types.ScreenshotRequest) ([]byte, error) {
	// Setup ChromeDP context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var buf []byte
	tasks := []chromedp.Action{
		chromedp.Navigate(req.URL),
		chromedp.Sleep(500 * time.Millisecond),
	}

	// Emulate mobile if needed
	if req.Device == "mobile" {
		tasks = append([]chromedp.Action{
			chromedp.Emulate(device.IPhone12),
		}, tasks...)
	} else {
		tasks = append([]chromedp.Action{
			chromedp.EmulateViewport(1920, 1080),
		}, tasks...)
	}

	if req.FullPage {
		tasks = append(tasks, chromedp.FullScreenshot(&buf, req.Quality))
	} else {
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
	}

	// Run tasks
	err := chromedp.Run(timeoutCtx, tasks...)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
