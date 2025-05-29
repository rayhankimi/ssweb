package utils

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"ssweb/types"
	"time"
)

func TakeScreenshot(req types.ScreenshotRequest) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var buf []byte
	tasks := []chromedp.Action{
		chromedp.Navigate(req.URL),
		chromedp.Sleep(1 * time.Second),
	}

	// Device emulation
	if req.Device == "mobile" {
		tasks = append([]chromedp.Action{
			chromedp.Emulate(device.IPhone12),
		}, tasks...)
	} else if req.Device == "tab" || req.Device == "tablet" {
		tasks = append([]chromedp.Action{
			chromedp.EmulateViewport(1366, 768),
		}, tasks...)
	} else {
		tasks = append([]chromedp.Action{
			chromedp.EmulateViewport(1366, 768),
		}, tasks...)
	}

	if req.FullPage {
		tasks = append(tasks,
			// Disable CSS animations
			chromedp.Evaluate(`
                const style = document.createElement('style');
                style.innerHTML = '*, *::before, *::after { animation-duration: 0s !important; animation-delay: 0s !important; transition-duration: 0s !important; transition-delay: 0s !important; }';
                document.head.appendChild(style);
                document.documentElement.style.scrollBehavior = 'auto';
            `, nil),

			// Auto scroll - FIXED JavaScript
			chromedp.Evaluate(`
                (async function() {
                    let totalHeight = 0;
                    const distance = 100;
                    const scrollHeight = document.body.scrollHeight;
                    
                    while (totalHeight < scrollHeight) {
                        window.scrollBy(0, distance);
                        totalHeight += distance;
                        await new Promise(resolve => setTimeout(resolve, 50));
                    }
                    
                    // Scroll back to top
                    window.scrollTo(0, 0);
                    await new Promise(resolve => setTimeout(resolve, 500));
                })();
            `, nil),

			chromedp.Sleep(1*time.Second),
			chromedp.FullScreenshot(&buf, 25),
		)
	} else {
		tasks = append(tasks,
			// Disable animations untuk viewport screenshot
			chromedp.Evaluate(`
                const style = document.createElement('style');
                style.innerHTML = '*, *::before, *::after { animation-duration: 0s !important; animation-delay: 0s !important; transition-duration: 0s !important; transition-delay: 0s !important; }';
                document.head.appendChild(style);
            `, nil),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.CaptureScreenshot(&buf),
		)
	}

	err := chromedp.Run(timeoutCtx, tasks...)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
