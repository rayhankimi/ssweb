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
	"ssweb/conf"
	"ssweb/handlers"
	"ssweb/middleware"
	"ssweb/utils"
	"syscall"
	"time"
)

func init() {
	// Load Env
	utils.LoadEnv()

	// Initialize ChromeDP allocator during start
	if err := conf.InitializeChromeDPAllocator(); err != nil {
		log.Fatal("Failed to initialize ChromeDP allocator:", err)
	}

	log.Println("ChromeDP allocator initialized successfully")

	// Log environment detection
	if conf.GetChromeDPConfig().NoSandbox {
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

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "Pong")
	})

	router.GET("/ssweb", middleware.TokenAuthMiddleware(), handlers.ScreenshotHandler)

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

	conf.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
