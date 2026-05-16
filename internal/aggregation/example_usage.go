// Package aggregation provides data aggregation services for the dinnerclub application.
// It includes restaurant data aggregation, popular item calculation, rating averages,
// and comprehensive restaurant analytics functionality.
package aggregation

// This file contains example usage of the RestaurantAggregationService
// It demonstrates how to integrate the service with controllers

/*
Example usage in a controller:

import (
    "github.com/gin-gonic/gin"
    "github.com/toozej/dinnerclub/internal/aggregation"
    "github.com/toozej/dinnerclub/pkg/database"
)

func RestaurantByNameHandler(c *gin.Context) {
    restaurantName := c.Param("name")

    // Initialize the aggregation service
    db := database.GetDB() // Assuming you have a database getter
    aggregationService := aggregation.NewRestaurantAggregationService(db)

    // Get comprehensive restaurant data
    aggregateData, err := aggregationService.GetRestaurantAggregateData(restaurantName)
    if err != nil {
        c.JSON(404, gin.H{"error": "Restaurant not found"})
        return
    }

    // Use the data in your template or JSON response
    c.HTML(200, "restaurant.html", gin.H{
        "restaurant":     aggregateData.Restaurant,
        "entryCount":     aggregateData.EntryCount,
        "popularItems":   aggregateData.PopularItems,
        "averageRatings": aggregateData.AverageRatings,
        "recentEntries":  aggregateData.RecentEntries,
    })
}

// For restaurant listing with entry counts:
func RestaurantListHandler(c *gin.Context) {
    db := database.GetDB()
    aggregationService := aggregation.NewRestaurantAggregationService(db)

    var restaurants []models.Restaurant
    db.Find(&restaurants)

    type RestaurantWithCount struct {
        models.Restaurant
        EntryCount int64
    }

    var restaurantsWithCounts []RestaurantWithCount
    for _, restaurant := range restaurants {
        count, _ := aggregationService.GetEntryCountByRestaurant(restaurant.Name)
        restaurantsWithCounts = append(restaurantsWithCounts, RestaurantWithCount{
            Restaurant: restaurant,
            EntryCount: count,
        })
    }

    // Sort by entry count (ascending as per requirements)
    sort.Slice(restaurantsWithCounts, func(i, j int) bool {
        return restaurantsWithCounts[i].EntryCount < restaurantsWithCounts[j].EntryCount
    })

    c.HTML(200, "restaurants.html", gin.H{
        "restaurants": restaurantsWithCounts,
    })
}
*/
