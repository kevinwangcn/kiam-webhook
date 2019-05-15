package engine

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"
	"github.com/openlab-red/kiam-webhook/pkg/kubernetes"
)

var log = kubernetes.Log()

func shutdown(engine *gin.Engine) {
	srv := &http.Server{
		Addr:    viper.GetString("port"),
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}
