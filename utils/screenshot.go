package utils

import (
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"ssweb/types"
	"time"
)

func TakeScreenshot(req types.ScreenshotRequest) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second) // increase timeout
	defer cancel()

	var buf []byte
	tasks := []chromedp.Action{
		chromedp.Navigate(req.URL),
		chromedp.Sleep(1 * time.Second), // tunggu initial load
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
		// Disable animations dan auto-scroll fot lazy loading
		tasks = append(tasks,
			// Disable CSS animations
			chromedp.Evaluate(`
                // Disable all animations
                const style = document.createElement('style');
                style.innerHTML = '*, *::before, *::after { animation-duration: 0s !important; animation-delay: 0s !important; transition-duration: 0s !important; transition-delay: 0s !important; }';
                document.head.appendChild(style);
                
                // Disable smooth scrolling
                document.documentElement.style.scrollBehavior = 'auto';
            `, nil),

			// Auto scroll to trigger lazy loading
			chromedp.Evaluate(`
                const scrollToBottom = () => {
                    return new Promise((resolve) => {
                        let totalHeight = 0;
                        const distance = 100;
                        const timer = setInterval(() => {
                            const scrollHeight = document.body.scrollHeight;
                            window.scrollBy(0, distance);
                            totalHeight += distance;

                            if(totalHeight >= scrollHeight){
                                clearInterval(timer);
                                // Scroll back to top
                                window.scrollTo(0, 0);
                                resolve();
                            }
                        }, 50); // 50ms interval
                    });
                };
                return scrollToBottom();
            `, nil),

			chromedp.Sleep(1*time.Second), // Wait after scroll
			chromedp.FullScreenshot(&buf, 25),
		)
	} else {
		tasks = append(tasks,
			// Disable animations for viewport screenshot
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
