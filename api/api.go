package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imafaz/b9m/controller"
)

type API struct {
	bindManager *controller.BindManager
	apiKey      string
}

func NewAPI(bindManager *controller.BindManager, apiKey string) *API {
	return &API{bindManager: bindManager, apiKey: apiKey}
}

func (api *API) authMiddleware(c *gin.Context) {
	if c.GetHeader("Authorization") != "Bearer "+api.apiKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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
	router.DELETE("/domains/:domain/records/:name", api.DeleteRecord)
	router.GET("/domains/:domain/records", api.GetAllRecords)
	router.POST("/reload", api.ReloadBind)
	router.POST("/restart", api.RestartBind)
	router.POST("/stop", api.StopBind)
	router.POST("/start", api.StartBind)
	router.GET("/status", api.StatusBind)
	router.GET("/stats", api.GetStats)
}
func (api *API) ReloadBind(c *gin.Context) {
	err := api.bindManager.ReloadBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bind reloaded successfully"})
}

func (api *API) RestartBind(c *gin.Context) {
	err := api.bindManager.RestartBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bind restarted successfully"})
}

func (api *API) StopBind(c *gin.Context) {
	err := api.bindManager.StopBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bind stopped successfully"})
}

func (api *API) StartBind(c *gin.Context) {
	err := api.bindManager.StartBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bind started successfully"})
}

func (api *API) StatusBind(c *gin.Context) {
	status, err := api.bindManager.StatusBind()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

func (api *API) GetStats(c *gin.Context) {
	stats, err := api.bindManager.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
func (api *API) AddDomain(c *gin.Context) {
	var input struct {
		Domain string `json:"domain" binding:"required"`
		NS1    string `json:"ns1" binding:"required"`
		NS2    string `json:"ns2" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := api.bindManager.AddDomain(input.Domain, input.NS1, input.NS2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Domain added successfully"})
}

func (api *API) DeleteDomain(c *gin.Context) {
	domain := c.Param("domain")
	err := api.bindManager.DeleteDomain(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Domain deleted successfully"})
}

func (api *API) AddRecord(c *gin.Context) {
	var input struct {
		Name  string                `json:"name" binding:"required"`
		Type  controller.RecordType `json:"type" binding:"required"`
		Value string                `json:"value" binding:"required"`
		TTL   int                   `json:"ttl" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := c.Param("domain")
	err := api.bindManager.AddRecord(domain, input.Type, input.Name, input.Value, input.TTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Record added successfully"})
}

func (api *API) DeleteRecord(c *gin.Context) {
	domain := c.Param("domain")
	name := c.Param("name")
	err := api.bindManager.DeleteRecord(domain, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Record deleted successfully"})
}

func (api *API) GetAllRecords(c *gin.Context) {
	domain := c.Param("domain")
	records, err := api.bindManager.GetAllRecords(domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func StartServer(port string, apiKey string) {
	bindManager := controller.NewBindManager("/etc/bind/zones", "/etc/bind/named.conf.local")
	api := NewAPI(bindManager, apiKey)
	router := gin.Default()
	api.SetupRoutes(router)

	log.Printf("Starting API server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
