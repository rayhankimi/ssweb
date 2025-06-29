package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	configs2 "ssweb/internal/configs"
	http2 "ssweb/internal/delivery/http"
	"ssweb/internal/middleware"
	"syscall"
	"time"
)

func init() {
	// Load Env
	configs2.LoadEnv()

	// Initialize ChromeDP allocator during start
	if err := configs2.InitializeChromeDPAllocator(); err != nil {
		log.Fatal("Failed to initialize ChromeDP allocator:", err)
	}

	log.Println("ChromeDP allocator initialized successfully")

	// Log environment detection
	if configs2.GetChromeDPConfig().NoSandbox {
		log.Println("Running in constrained environment (Docker/limited resources) - sandbox disabled")
	} else {
		log.Println("Running in normal environment - sandbox enabled")
	}
}

func main() {
	fmt.Println("Starting ssweb...")
	//ginMode := utils.GetEnv("GIN_MODE", "release")
	ginMode := "release"
	gin.SetMode(ginMode)

	router := gin.Default()

	if os.Getenv("GIN_MODE") == "release" {
		err := router.SetTrustedProxies([]string{"127.0.0.1", "10.0.0.0/8"})
		if err != nil {
			fmt.Println("Error setting trusted proxies: ", err)
			return
		}
	} else {
		err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"})
		if err != nil {
			fmt.Println("Error setting trusted proxies: ", err)
			return
		}
	}

	//err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	//if err != nil {
	//	fmt.Println("Error setting trusted proxies: ", err)
	//	return
	//}

	router.POST("/ping", func(c *gin.Context) {
		c.String(200, "Pong")
	})

	router.POST("/ssweb", middleware.TokenAuthMiddleware(), http2.ScreenshotHandler)

	port := os.Getenv("PORT")
	//port := "8080"
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	configs2.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
