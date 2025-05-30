package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"middleman/pkg/config"
	"middleman/pkg/middleware"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	server *http.Server
	router *gin.Engine
}

func NewHttpServer() *HttpServer {
	conf := config.GetConf()
	r := gin.Default()

	r.POST("register/", handleRegister)

	g := r.Group("middleman/")
	g.Use(middleware.AccessKeyMiddleware())
	g.GET("slave-nodes/", getSlaveNodes)

	g.Use(middleware.DatabaseMiddleware())
	g.GET("resources/", getResources)
	g.POST("resources/", saveResources)
	g.DELETE("resources/:id/", deleteResources)

	return &HttpServer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%v", conf.Port),
			Handler: r,
		},
		router: r,
	}
}

func (s *HttpServer) Start() error {
	log.Printf("HTTP服务器启动，监听端口 %v\n", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *HttpServer) Stop() {
	log.Println("正在关闭HTTP服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
}

func RunForever() {
	httpServer := NewHttpServer()
	go func() {
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-gracefulStop
	httpServer.Stop()
	log.Println("所有服务已关闭")
}
