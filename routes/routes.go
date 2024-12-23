package routes
 
import (
    "multitenant/handlers"
 
    "github.com/gorilla/mux"
)
 
// InitializeRoutes initializes all the routes for the application
func InitializeRoutes() *mux.Router {
    router := mux.NewRouter()
 
    // Public routes (No authentication required)
    router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
 
    // Admin routes
    adminRouter := router.PathPrefix("/admin").Subrouter()
    adminRouter.Use(handlers.Authenticate)          // Middleware to verify JWT token
    adminRouter.Use(handlers.Authorize("admin"))   // Middleware to allow only Admin
    adminRouter.HandleFunc("/create-manager", handlers.CreateManagerHandler).Methods("POST")
    adminRouter.HandleFunc("/delete-manager", handlers.RemoveManagerHandler).Methods("DELETE")
 
    // Manager routes
    managerRouter := router.PathPrefix("/manager").Subrouter()
    managerRouter.Use(handlers.Authenticate)       // Middleware to verify JWT token
    managerRouter.Use(handlers.Authorize("manager")) // Middleware to allow only Managers
    managerRouter.HandleFunc("/create-group", handlers.CreateGroupHandler).Methods("POST")
    managerRouter.HandleFunc("/create-user", handlers.CreateUserHandler).Methods("POST")
    managerRouter.HandleFunc("/delete-user", handlers.DeleteUserHandler).Methods("DELETE")
    managerRouter.HandleFunc("/add-user", handlers.AddUserHandler).Methods("POST")
    managerRouter.HandleFunc("/remove-user", handlers.RemoveUserHandler).Methods("DELETE")
    managerRouter.HandleFunc("/list-groups", handlers.ListGroupsHandler).Methods("GET")
    managerRouter.HandleFunc("/check-user-group", handlers.CheckUserGroupHandler).Methods("GET")
    managerRouter.HandleFunc("/add-budget", handlers.AddBudgetHandler).Methods("POST")
    managerRouter.HandleFunc("/update-budget", handlers.UpdateBudgetHandler).Methods("PUT")
 
    // User routes
    userRouter := router.PathPrefix("/user").Subrouter()
    userRouter.Use(handlers.Authenticate)          // Middleware to verify JWT token
    userRouter.Use(handlers.Authorize("user"))     // Middleware to allow only Users
 
    //user session management
    userRouter.HandleFunc("/get-cloud-services", handlers.GetCloudServicesHandler).Methods("GET")
    userRouter.HandleFunc("/start-session", handlers.StartSessionHandler).Methods("GET")
    userRouter.HandleFunc("/update-session", handlers.UpdateSessionHandler).Methods("POST")
    userRouter.HandleFunc("/calculate-cost", handlers.CalculateCostHandler).Methods("POST")
    userRouter.HandleFunc("/complete-session", handlers.CompleteSessionHandler).Methods("POST")
 
    // userRouter.HandleFunc("/fetch-aws-price", handlers.FetchAWSServicePriceHandler).Methods("POST")
   
    // AWS Service routes
    userRouter.HandleFunc("/create-ec2-instance", handlers.CreateEC2InstanceHandler).Methods("POST")
    userRouter.HandleFunc("/create-s3-bucket", handlers.CreateS3BucketHandler).Methods("POST")
    userRouter.HandleFunc("/create-lambda-function", handlers.CreateLambdaFunctionHandler).Methods("POST")
    userRouter.HandleFunc("/create-rds-instance", handlers.CreateRDSInstanceHandler).Methods("POST")
    // userRouter.HandleFunc("/create-dynamodb-table", handlers.CreateDynamoDBTableHandler).Methods("POST")
    userRouter.HandleFunc("/create-cloudfront-distribution", handlers.CreateCloudFrontDistributionHandler).Methods("POST")
    userRouter.HandleFunc("/create-vpc", handlers.CreateVPCHandler).Methods("POST")
 
    // routes for GCP service creation
    userRouter.HandleFunc("/create-compute-engine", handlers.CreateComputeEngineHandler).Methods("POST")
    userRouter.HandleFunc("/create-cloud-storage", handlers.CreateCloudStorageHandler).Methods("POST")
    userRouter.HandleFunc("/create-GKE-cluster", handlers.CreateGKEClusterHandler).Methods("POST")
    userRouter.HandleFunc("/create-bigquery-dataset", handlers.CreateBigQueryDatasetHandler).Methods("POST")
    userRouter.HandleFunc("/create-cloud-SQL", handlers.CreateCloudSQLHandler).Methods("POST")
 
    // router.HandleFunc("/fetch-aws-price", handlers.FetchAWSServicePriceHandler).Methods("POST")
    // router.HandleFunc("/fetch-gcp-price", handlers.FetchGCPServicePriceHandler).Methods("POST")

	userRouter.HandleFunc("/delete-aws-service", handlers.DeleteAWSServiceHandler).Methods("POST")
    userRouter.HandleFunc("/delete-gcp-service", handlers.DeleteGCPServiceHandler).Methods("POST")

    userRouter.HandleFunc("/send-notification", handlers.SendNotificationHandler).Methods("POST")

    userRouter.HandleFunc("/fetch-service-cost", handlers.FetchServiceCostHandler).Methods("POST")

    return router
}