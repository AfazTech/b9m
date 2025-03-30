package api

import (
	"net/http"
	"strconv"

	"github.com/AfazTech/b9m/parser"
	"github.com/AfazTech/b9m/record"
	"github.com/AfazTech/b9m/servicemanager"
	"github.com/AfazTech/b9m/zone"
	"github.com/AfazTech/logger/v2"
	"github.com/gin-gonic/gin"
)

type API struct {
	apiKey string
}

func NewAPI(apiKey string) *API {
	return &API{apiKey: apiKey}
}

func (api *API) authMiddleware(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer "+api.apiKey {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "message": "Unauthorized"})
		c.Abort()
		return
	}
	c.Next()
}

func (api *API) SetupRoutes(router *gin.Engine) {
	router.Use(api.authMiddleware)
	router.POST("/domains", api.AddDomain)
	router.DELETE("/domains/:domain", api.DeleteDomain)
	router.POST("/domains/:domain/records", api.AddRecord)
	router.DELETE("/domains/:domain/records/:name/:type/:value", api.DeleteRecord)
	router.GET("/domains/:domain/records", api.GetAllRecords)
	router.GET("/domains", api.GetDomains)
	router.POST("/reload", api.ReloadBind)
	router.POST("/restart", api.RestartBind)
	router.POST("/stop", api.StopBind)
	router.POST("/start", api.StartBind)
	router.GET("/status", api.StatusBind)
}

func (api *API) ReloadBind(c *gin.Context) {
	err := servicemanager.ReloadBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Bind reloaded successfully"})
}

func (api *API) RestartBind(c *gin.Context) {
	err := servicemanager.RestartBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Bind restarted successfully"})
}

func (api *API) StopBind(c *gin.Context) {
	err := servicemanager.StopBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Bind stopped successfully"})
}

func (api *API) StartBind(c *gin.Context) {
	err := servicemanager.StartBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Bind started successfully"})
}

func (api *API) StatusBind(c *gin.Context) {
	status, err := servicemanager.StatusBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func (api *API) AddDomain(c *gin.Context) {
	var input struct {
		Domain string `json:"domain" binding:"required"`
		NS1    string `json:"ns1" binding:"required"`
		NS2    string `json:"ns2" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": err.Error()})
		return
	}

	err := zone.AddDomain(input.Domain, input.NS1, input.NS2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ok": true, "message": "Domain added successfully"})
}

func (api *API) DeleteDomain(c *gin.Context) {
	domain := c.Param("domain")
	err := zone.DeleteDomain(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Domain deleted successfully"})
}

func (api *API) AddRecord(c *gin.Context) {
	var input struct {
		Name  string            `json:"name" binding:"required"`
		Type  record.RecordType `json:"type" binding:"required"`
		Value string            `json:"value" binding:"required"`
		TTL   string            `json:"ttl" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": err.Error()})
		return
	}

	ttl, err := strconv.Atoi(input.TTL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "ttl must be a valid integer"})
		return
	}

	domain := c.Param("domain")
	err = record.AddRecord(domain, input.Type, input.Name, input.Value, ttl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ok": true, "message": "Record added successfully"})
}

func (api *API) DeleteRecord(c *gin.Context) {
	domain := c.Param("domain")
	name := c.Param("name")
	rType := c.Param("type")
	value := c.Param("value")
	err := record.DeleteRecord(domain, name, record.RecordType(rType), value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Record deleted successfully"})
}

func (api *API) GetAllRecords(c *gin.Context) {
	domain := c.Param("domain")
	records, err := record.GetAllRecords(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "records": records})
}

func (api *API) GetDomains(c *gin.Context) {
	domains, err := parser.GetDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "domains": domains})
}

func StartServer(port string, apiKey string) {
	api := NewAPI(apiKey)
	router := gin.Default()
	api.SetupRoutes(router)

	logger.Infof("Starting API server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		logger.Fatalf("Failed to run server: %v", err)
	}
}
