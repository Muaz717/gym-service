package httpApp

import (
	personHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/person"
	personSubHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/person_sub"
	singleVisitHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/single_visit"
	statHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/statistics"
	subFreezeHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/sub_freeze"
	subscriptionHandler "github.com/Muaz717/gym_app/app/internal/http/handlers/subscription"
	"github.com/gin-gonic/gin"
)

func registerPersonRoutes(api *gin.RouterGroup, h *personHandler.PersonHandler, admin gin.HandlerFunc) {
	r := api.Group("/people")
	r.GET("", h.FindAllPeople)
	r.GET("/find", h.FindPersonByName)
	r.GET("/find/:id", h.FindPersonById)

	adminGroup := r.Group("")
	adminGroup.Use(admin)
	adminGroup.POST("/add", h.AddPerson)
	adminGroup.PUT("update/:id", h.UpdatePerson)
	adminGroup.DELETE("delete/:id", h.DeletePerson)
}

func registerSubscriptionRoutes(api *gin.RouterGroup, h *subscriptionHandler.SubscriptionHandler, admin gin.HandlerFunc) {
	r := api.Group("/subscription")
	r.GET("", h.FindAllSubscriptions)

	adminGroup := r.Group("")
	adminGroup.Use(admin)
	adminGroup.POST("/add", h.AddSubscription)
	adminGroup.PUT("update/:id", h.UpdateSubscription)
	adminGroup.DELETE("delete/:id", h.DeleteSubscription)
}

func registerPersonSubRoutes(api *gin.RouterGroup, h *personSubHandler.PersonSubHandler, admin gin.HandlerFunc) {
	r := api.Group("/person_sub")
	r.GET("", h.FindAllPersonSubs)
	r.GET("/find", h.FindPersonSubByPersonName)
	r.GET("/find/:number", h.FindPersonSubByNumber)
	r.GET("/find/id/:id", h.FindPersonSubByPersonId)

	adminGroup := r.Group("")
	adminGroup.Use(admin)
	adminGroup.POST("/add", h.AddPersonSub)
	adminGroup.DELETE("delete/:number", h.DeletePersonSub)
}

func registerFreezeRoutes(api *gin.RouterGroup, h *subFreezeHandler.SubFreezeHandler, admin gin.HandlerFunc) {
	r := api.Group("/freeze")
	r.GET("", h.GetAllActiveFreeze)

	adminGroup := r.Group("")
	adminGroup.Use(admin)
	adminGroup.POST("/add", h.FreezeSubscription)
	adminGroup.POST("/unfreeze", h.UnfreezeSubscription)
}

func registerSingleVisitRoutes(api *gin.RouterGroup, h *singleVisitHandler.SingleVisitHandler, admin gin.HandlerFunc) {
	r := api.Group("/single_visit")
	r.GET("", h.GetAllSingleVisits)
	r.GET("/:id", h.GetSingleVisitById)
	r.GET("/day", h.GetSingleVisitsByDay)
	r.GET("/period", h.GetSingleVisitsByPeriod)

	adminGroup := r.Group("")
	adminGroup.Use(admin)
	adminGroup.POST("/add", h.AddSingleVisit)
	adminGroup.DELETE("/delete/:id", h.DeleteSingleVisit)
}

func registerStatRoutes(api *gin.RouterGroup, h *statHandler.StatHandler) {
	r := api.Group("/statistics")
	r.GET("/total_clients", h.TotalClients)
	r.GET("/new_clients", h.NewClients)
	r.GET("/total_income", h.TotalIncome)
	r.GET("/income", h.Income)
	r.GET("/total_sold_subscriptions", h.TotalSoldSubscriptions)
	r.GET("/sold_subscriptions", h.SoldSubscriptions)
	r.GET("/monthly", h.MonthlyStatistics)
	r.GET("/total_single_visits", h.TotalSingleVisits)
	r.GET("/single_visits", h.SingleVisits)
	r.GET("/single_visits_income", h.SingleVisitsIncome)
}
