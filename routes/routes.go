package routes

import (
	"multitenant/handlers"
	"github.com/gorilla/mux"
)

// InitializeRoutes initializes all the routes for the application
func InitializeRoutes() *mux.Router {
	router := mux.NewRouter()

	// Define routes for login and other endpoints
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/create-manager", handlers.CreateManagerHandler).Methods("POST")
	router.HandleFunc("/delete-manager", handlers.RemoveManagerHandler).Methods("DELETE")
    router.HandleFunc("/create-group", handlers.CreateGroupHandler).Methods("POST")
	router.HandleFunc("/add-user", handlers.AddUserHandler).Methods("POST")
	router.HandleFunc("/remove-user", handlers.RemoveUserHandler).Methods("DELETE")
	router.HandleFunc("/create-user", handlers.CreateUserHandler).Methods("POST")
	router.HandleFunc("/delete-user", handlers.DeleteUserHandler).Methods("DELETE")
    router.HandleFunc("/add-budget", handlers.AddBudgetHandler).Methods("POST")
    router.HandleFunc("/update-budget", handlers.UpdateBudgetHandler).Methods("PUT")
    router.HandleFunc("/list-groups", handlers.ListGroupsHandler).Methods("GET")
	router.HandleFunc("/check-user-group", handlers.CheckUserGroupHandler).Methods("GET")

	// routes for AWS service creation
	router.HandleFunc("/create-ec2-instance", handlers.CreateEC2InstanceHandler).Methods("POST")

	router.HandleFunc("/create-s3-bucket", handlers.CreateS3BucketHandler).Methods("POST")

	router.HandleFunc("/create-lambda-function", handlers.CreateLambdaFunctionHandler).Methods("POST")
	router.HandleFunc("/get-lambda-runtimes", handlers.GetLambdaRuntimesHandler).Methods("GET")

	router.HandleFunc("/create-rds-instance", handlers.CreateRDSInstanceHandler).Methods("POST")

	router.HandleFunc("/create-dynamodb-table", handlers.CreateDynamoDBTableHandler).Methods("POST")

	router.HandleFunc("/create-cloudfront-distribution", handlers.CreateCloudFrontDistributionHandler).Methods("POST")
	
	router.HandleFunc("/create-vpc", handlers.CreateVPCHandler).Methods("POST")

	return router
}
