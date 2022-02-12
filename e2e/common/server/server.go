package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Routes interface {
	RegisterRoutes(r gin.IRoutes)
}

// Server handles incoming http requests
type Server struct {
	routes   []Routes
	stopOnce sync.Once
	stop     chan struct{}
}

// NewServer returns a new server
func NewServer(routes ...Routes) *Server {
	return &Server{
		routes: routes,
		stop:   make(chan struct{}),
	}
}

// Serve starts the server
func (s *Server) Serve() error {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(Logger())
	c := cors.DefaultConfig()
	c.AllowAllOrigins = true
	c.AllowHeaders = append(c.AllowHeaders, "Authorization")
	r.Use(cors.New(c))

	for i := range s.routes {
		s.routes[i].RegisterRoutes(r)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	}

	go func() {
		<-s.stop
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	return srv.ListenAndServe()
}

// Stop stops the server
func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
}

// Logger forces gin to use our logger
// Adapted from gin.Logger
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Info().Fields(map[string]interface{}{
			"path":    path,
			"method":  method,
			"status":  statusCode,
			"latency": latency,
			"comment": comment,
		}).Msgf("API Request")
	}
}
