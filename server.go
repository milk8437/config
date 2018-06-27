package config

import (
	"time"
	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
)

type Server struct {
	Port int
}


//Logger gin用的日志
//记录请求耗费时间
func LoggerForGin() gin.HandlerFunc {
	l := LOG.Named("GIN")
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		// Process request
		c.Next()
		end := time.Now()
		//延迟
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		l.Info("request running time",
			zap.String("path", path),
			zap.String("method", method),
			zap.String("clientIP", clientIP),
			zap.String("latency", fmt.Sprintf("%+7v", latency)),
			zap.Int("status", statusCode),
		)
	}
}

//Recovery 异常恢复
//发生panic及时恢复并且记录日志，返回固定resp
//panic会导致整个应用崩溃
func RecoveryForGin() gin.HandlerFunc {
	l := LOG.Named("GIN.RECOVERY")
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				path := c.Request.URL.Path
				clientIP := c.ClientIP()
				method := c.Request.Method
				switch x := err.(type) {
				case error:
					//error 会有stacktrace信息打印
					l.Error("[Recovery] panic recovered",
						zap.String("path", path),
						zap.String("clientIP", clientIP),
						zap.String("method", method),
						zap.Error(x),
					)
					break
				case string:

					l.Error("[Recovery] panic recovered",
						zap.String("path", path),
						zap.String("clientIP", clientIP),
						zap.String("method", method),
						zap.String("error", x),
					)
					break
				default:
					l.Error("[Recovery] panic recovered",
						zap.String("path", path),
						zap.String("clientIP", clientIP),
						zap.String("method", method),
						zap.String("error", "known error"),
					)
					break
				}
				c.JSON(http.StatusInternalServerError, &gin.H{
					"message": "error",
					"code":    "5000",
				})
				//这个会强制中断链路调用
				c.Abort()
			}
		}()
		c.Next()
	}
}
