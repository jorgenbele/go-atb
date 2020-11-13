// Date: 07.10.2020
// Author: Jørgen Bele Reinfjell
// Description:
//  Proxy API server with caching for
//  the undocumented TravelPlanner API for AtB

package atb

import (
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// CreateAPIEngine returns an instantiated gin router 
// with the api
func CreateAPIEngine(allowedAPIKey string) *gin.Engine {

	r := gin.Default()

	store := persistence.NewInMemoryStore(time.Second)

	v1 := r.Group("/v1")

	// Suggestions (not API limited)
	v1.GET("/stops/suggestions/:query", cache.CachePage(store, 24*time.Hour, func(c *gin.Context) {
		query := c.Param("query")

		v, err := GetSuggestions(query)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
				"error":  "unable to get suggestions",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"results": v,
			"cached":  time.Now().Unix(),
		})
	}))

	// Binding from JSON
	type DeparturesRequest struct {
		From          string `form:"from" json:"from" binding:"required"`
		To            string `form:"to" json:"to" binding:"required"`
		Time          string `form:"time" json:"time" binding:"required"`
		Date          string `form:"date" json:"date" binding:"required"`
		IsArrivalTime bool   `form:"is_arrival_time" json:"is_arrival_time" binding:"required"`
		// Not currently working, so disabled
		// IsRealtime    bool

		APIKey string `form:"api_key" json:"api_key" binding:"required"`
	}

	// Departures (API limited)
	v1.POST("/travel/departures", cache.CachePage(store, 10*time.Minute, func(c *gin.Context) {
		var json DeparturesRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.APIKey != allowedAPIKey {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "unauthorized",
				"error":  "invalid api key",
			})
			return
		}

		req := DepartureReq{json.From, json.To, json.Time, json.Date, json.IsArrivalTime, false}
		departures, err := GetDeparturesReq(req)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "unable to get departures",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"results": departures,
			"cached":  time.Now().Unix(),
		})
	}))
    return r
}
