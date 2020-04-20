package node

import (
	"fmt"
	"net/http"

	"github.com/SkycoinProject/cxo-2/pkg/errors"
	"github.com/SkycoinProject/cxo-2/pkg/model"
	"github.com/SkycoinProject/cxo-2/pkg/node/data"
	"github.com/gin-gonic/gin"
)

const (
	host = "127.0.0.1"
	port = 6421
)

type WebServer struct {
	Engine *gin.Engine
}

type Controller struct {
	Data data.Data
}

type ErrorResponse struct {
	Error string `json:"message"`
}

func InitServerAndController(data data.Data) *WebServer {
	server := &WebServer{
		Engine: gin.Default(),
	}

	ctrl := &Controller{Data: data}
	server.initRoutes(ctrl)
	return server
}

func (s *WebServer) Run() {
	if err := s.Engine.Run(fmt.Sprintf("%s:%v", host, port)); err != nil {
		panic(err.Error())
	}
}

func (s *WebServer) initRoutes(ctrl *Controller) {
	public := s.Engine.Group("/api/v1")
	public.POST("/registerApp", ctrl.registerApp)
}

func (ctrl *Controller) registerApp(c *gin.Context) {
	var app model.RegisterAppRequest
	if err := c.BindJSON(&app); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, ErrorResponse{Error: errors.ErrUnableToProcessRequest.Error()})
		return
	}

	if err := ctrl.Data.RegisterApp(app.Address, app.Name); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)
}
