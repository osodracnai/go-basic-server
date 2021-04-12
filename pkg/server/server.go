package server

import (
	"fmt"
	"github.com/Depado/ginprom"
	"github.com/osodracnai/go-basic-server/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
)

type Server struct {
	validate *validator.Validate
}

// New is method to get a new server instance
func New() (*Server, error) {
	s := Server{
		validate: validator.New(),
	}
	return &s, nil
}

//Create New Engine
func (s *Server) NewEngine() *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	if logrus.GetLevel() == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	p := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
	)

	r.Use(TracerMiddleware)
	if p != nil {
		r.Use(p.Instrument())
	}
	r.Use(Logger())
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	gin.SetMode(gin.ReleaseMode)
	if logrus.GetLevel() == logrus.DebugLevel || logrus.GetLevel() == logrus.TraceLevel {
		gin.SetMode(gin.DebugMode)
	}

	r.POST("/msg", s.SendMessage)

	return r
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Output: nil,
		Formatter: func(params gin.LogFormatterParams) string {
			fields := logrus.Fields{
				utils.KeyHTTPPath.String():       params.Path,
				utils.KeyHTTPMethod.String():     params.Method,
				utils.KeyHTTPStatusCode.String(): params.StatusCode,
				utils.KeyHTTPBodySize.String():   params.BodySize,
				utils.KeyHTTPClientIP.String():   params.ClientIP,
				utils.KeyHTTPLatency.String():    params.Latency.Milliseconds(),
				utils.KeyHTTPUrl.String():        params.Request.URL.String(),
			}
			for key, val := range params.Keys {
				fields[key] = val
			}
			if params.ErrorMessage != "" {
				fields[utils.KeyError.String()] = true
				logrus.WithFields(fields).Error(params.ErrorMessage)
			}
			logrus.WithFields(fields).Info("")
			return ""
		},
	})
}

func TracerMiddleware(c *gin.Context) {
	tr := opentracing.GlobalTracer()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	ctx, _ := tr.Extract(opentracing.HTTPHeaders, carrier)
	name := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
	sp := tr.StartSpan(name, ext.RPCServerOption(ctx))
	ext.HTTPMethod.Set(sp, c.Request.Method)
	ext.HTTPUrl.Set(sp, c.Request.URL.String())

	componentName := c.HandlerName()

	ext.Component.Set(sp, componentName)
	c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), sp))

	detectedErrors := c.Errors.ByType(gin.ErrorTypeAny)

	if len(detectedErrors) > 0 {
		ext.Error.Set(sp, true)
		err := detectedErrors[len(detectedErrors)-1].Err
		sp.SetTag(utils.KeyErrorMessage.String(), err.Error())
	} else {
		ext.Error.Set(sp, false)
	}
	ext.HTTPStatusCode.Set(sp, uint16(c.Writer.Status()))
	sp.Finish()
}

