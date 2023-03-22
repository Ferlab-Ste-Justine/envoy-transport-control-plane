package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

func server(server string, port int64) {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"server": server,
		})
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: router,
	}
	go func() {
		srv.ListenAndServe()
	}()
	fmt.Printf("Started server %s\n", server)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	server("server1", 8081)
	server("server2", 8082)
	server("server3", 8083)
	server("server4", 8084)
	server("server5", 8085)
    for {
        time.Sleep(1 * time.Hour)
    }
}