package usecase

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"ssweb/internal/core/entity"
	"time"
)

func TakeScreenshot(req entity.ScreenshotRequest) ([]byte, error) {
	// Setup ChromeDP dengan options yang lebih robust
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-web-security", true), // untuk bypass CORS
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel1 := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel1()

	ctx, cancel2 := chromedp.NewContext(allocCtx)
	defer cancel2()

	timeoutCtx, cancel3 := context.WithTimeout(ctx, 20*time.Second)
	defer cancel3()

	var buf []byte
	tasks := []chromedp.Action{
		chromedp.Navigate(req.URL),
		chromedp.Sleep(2 * time.Second), // wait for initial load
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
			// Wait for page to be more ready
			chromedp.WaitVisible("body", chromedp.ByQuery),

			// Advanced content loading strategy
			chromedp.Evaluate(`
                (function() {
                    // Disable animations safely with TrustedHTML check
                    try {
                        const style = document.createElement('style');
                        const css = '*, *::before, *::after { animation-duration: 0.01s !important; animation-delay: 0s !important; transition-duration: 0.01s !important; transition-delay: 0s !important; } body { scroll-behavior: auto !important; }';
                        
                        if (window.trustedTypes && window.trustedTypes.createPolicy) {
                            const policy = window.trustedTypes.createPolicy('screenshot-policy', {
                                createHTML: (string) => string
                            });
                            style.innerHTML = policy.createHTML(css);
                        } else {
                            style.innerHTML = css;
                        }
                        
                        document.head.appendChild(style);
                    } catch(e) {
                        console.log('Animation disable failed:', e);
                    }

                    // Remove sticky/fixed positioning that causes navbar issues
                    try {
                        const elements = document.querySelectorAll('*');
                        elements.forEach(el => {
                            const computedStyle = window.getComputedStyle(el);
                            if (computedStyle.position === 'fixed' || computedStyle.position === 'sticky') {
                                el.style.position = 'static !important';
                            }
                        });
                    } catch(e) {
                        console.log('Position fix failed:', e);
                    }
                })();
            `, nil),

			chromedp.Sleep(1*time.Second),

			// Smart scrolling to load content progressively
			chromedp.Evaluate(`
                (function() {
                    return new Promise((resolve) => {
                        const loadContent = async () => {
                            const viewportHeight = window.innerHeight;
                            const documentHeight = Math.max(
                                document.body.scrollHeight,
                                document.body.offsetHeight,
                                document.documentElement.clientHeight,
                                document.documentElement.scrollHeight,
                                document.documentElement.offsetHeight
                            );
                            
                            let currentPosition = 0;
                            const scrollStep = viewportHeight * 0.8; // 80% of viewport
                            
                            while (currentPosition < documentHeight) {
                                window.scrollTo(0, currentPosition);
                                
                                // Wait for lazy loading
                                await new Promise(r => setTimeout(r, 300));
                                
                                // Trigger any lazy loading by checking for common lazy load attributes
                                try {
                                    const lazyElements = document.querySelectorAll('[data-src], [loading="lazy"], .lazy');
                                    lazyElements.forEach(el => {
                                        if (el.dataset.src && !el.src) {
                                            el.src = el.dataset.src;
                                        }
                                        // Trigger intersection observer
                                        el.dispatchEvent(new Event('scroll'));
                                    });
                                } catch(e) {
                                    console.log('Lazy load trigger failed:', e);
                                }
                                
                                currentPosition += scrollStep;
                                
                                // Update document height as content might have loaded
                                const newHeight = Math.max(
                                    document.body.scrollHeight,
                                    document.body.offsetHeight,
                                    document.documentElement.clientHeight,
                                    document.documentElement.scrollHeight,
                                    document.documentElement.offsetHeight
                                );
                                
                                if (newHeight > documentHeight) {
                                    documentHeight = newHeight;
                                }
                            }
                            
                            // Final scroll to bottom to ensure everything is loaded
                            window.scrollTo(0, documentHeight);
                            await new Promise(r => setTimeout(r, 500));
                            
                            // Scroll back to top for screenshot
                            window.scrollTo(0, 0);
                            await new Promise(r => setTimeout(r, 300));
                            
                            resolve();
                        };
                        
                        loadContent();
                    });
                })();
            `, nil),

			chromedp.Sleep(1*time.Second),
			chromedp.FullScreenshot(&buf, 25),
		)
	} else {
		tasks = append(tasks,
			chromedp.WaitVisible("body", chromedp.ByQuery),
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
