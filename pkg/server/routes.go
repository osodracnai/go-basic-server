package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) SendMessage(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"ok": "ok"})
}
