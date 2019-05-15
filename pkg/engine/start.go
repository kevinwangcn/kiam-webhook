package engine

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/openlab-red/kiam-webhook/pkg/kubernetes"
)

func Start() {
	var engine = gin.New()

	kubernetes.InitLogrus(engine)

	engine.GET("/health", health)

	hook(engine)

	engine.RunTLS(":"+viper.GetString("port"), "/var/run/secrets/kubernetes.io/certs/tls.crt", "/var/run/secrets/kubernetes.io/certs/tls.key")

	shutdown(engine)
}

func hook(engine *gin.Engine) {

	config := kubernetes.WebHookConfig{}
	kubernetes.Load("/var/run/secrets/kubernetes.io/config/kiam-config.yaml", &config)

	wk := kubernetes.WebHook{
		Config: &config,
	}

	engine.POST("/mutate", wk.Mutate)

}
