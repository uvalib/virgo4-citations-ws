package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	//"github.com/prometheus/client_golang/prometheus"
	//"github.com/prometheus/client_golang/prometheus/promhttp"
	//ginprometheus "github.com/zsais/go-gin-prometheus"
)

/*
 * Main entry point for the web service
 */
func main() {
	log.Printf("===> virgo4-citations-ws starting up <===")

	cfg := loadConfig()
	svc := initializeService(cfg)

	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	router := gin.Default()

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowCredentials = true
	corsCfg.AddAllowHeaders("Authorization")
	router.Use(cors.New(corsCfg))

	//
	// we are removing Prometheus support for now
	//
	//p := ginprometheus.NewPrometheus("gin")

	// roundabout setup of /metrics endpoint to avoid extra gzip of response
	//router.Use(p.HandlerFunc())
	//h := promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{DisableCompression: true}))

	//router.GET(p.MetricsPath, func(c *gin.Context) {
	//	h.ServeHTTP(c.Writer, c.Request)
	//})

	router.GET("/favicon.ico", svc.ignoreHandler)

	router.GET("/version", svc.versionHandler)
	router.GET("/healthcheck", svc.healthCheckHandler)

	if format := router.Group("/format"); format != nil {
		format.GET("/all", svc.allHandler)
		format.GET("/apa", svc.apaHandler)
		format.GET("/citeas", svc.citeAsHandler)
		format.GET("/cms", svc.cmsHandler)
		format.GET("/mla", svc.mlaHandler)
		format.GET("/lbb", svc.lbbHandler)
		format.GET("/ris", svc.risHandler)
	}

	router.GET("/unapi", svc.unapiHandler) // unAPI endpoint for Zotero

	portStr := fmt.Sprintf(":%s", svc.config.Port)
	log.Printf("[MAIN] listening on %s", portStr)

	log.Fatal(router.Run(portStr))
}
