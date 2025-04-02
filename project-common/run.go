package common

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动HTTP服务并监听指定端口，处理优雅的启动和停止
// 参数:
//
//	r *gin.Engine: Gin的HTTP路由引擎
//	srvName string: 服务的名称，用于日志等
//	addr string: 服务监听的地址和端口
//	stop func(): 服务停止前需要执行的清理函数
func Run(r *gin.Engine, srvName string, addr string, stop func()) {
	// 创建HTTP服务实例
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 使用goroutine启动服务，避免阻塞主goroutine
	go func() {
		log.Printf("%s running in %s \n", srvName, srv.Addr)
		// 监听并服务HTTP请求，发生错误时记录日志
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	// 创建一个用于接收终止信号的通道
	quit := make(chan os.Signal)
	// 监听系统中断和终止信号，当收到信号时，将信号发送到quit通道
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞等待信号，当收到信号时继续执行下面的代码
	<-quit
	log.Printf("Shutting Down project %s... \n", srvName)

	// 创建一个带有超时的context，用于优雅地关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 如果提供了清理函数，则执行清理函数
	if stop != nil {
		stop()
	}
	// 优雅地关闭HTTP服务
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("%s Shutdown, cause by : %v", srvName, err)
	}
	// 等待context超时，确保服务有足够的时间进行清理
	select {
	case <-ctx.Done():
		log.Println("wait timeout....")
	}
	log.Printf("%s stop success... \n", srvName)
}
