# **API Documentation**

The API provides an interface to interact with a vault. Here is the basic structure of the API code:

```go
func main() {
	router := gin.Default()
	router.Use(middlewares.CorsMiddleware())
	router.Use(middlewares.AuthMiddleware())
	routes.ApplyRoutes(router)
	router.Run(":8080")
	return
}
```

The documentation below details the different routes and functionalities offered by this API.

## **Middlewares**

### **CorsMiddleware**

The **`CorsMiddleware`** middleware handles Cross-Origin Resource Sharing (CORS) issues by adding appropriate headers to the responses. It allows requests from different domains to access the API's resources.

### **AuthMiddleware**

The **`AuthMiddleware`** middleware is responsible for client authentication with the vault. It uses the environment variables **`SECURE_ENV_TOKEN`** and **`SECURE_ENV_TOKEN`** to connect to the vault. Once authenticated, it stores the vault client in the request context for later use.

## **Routes**

### **Body (JSON)**

```
{
	"Value": "Replace with the expected value"
}
```

### **GET /project/**

This route retrieves the list of projects from the vault.

### **POST /project/**

This route creates a new project in the vault. The project details must be provided in the request body in JSON format.

### **PATCH /project/:project/**

This route renames an existing project in the vault. The project name must be specified in the URL, and the update details should be provided in the request body in JSON format.

### **DELETE /project/:project/**

This route deletes a project from the vault. The project name must be specified in the URL.

### **GET /project/:project/**

This route retrieves the details of a specific project from the vault. The project name must be specified in the URL.

### **GET /project/:project/var/**

This route retrieves the list of variables associated with a specific project from the vault. The project name must be specified in the URL.

### **POST /project/:project/var/:variable**

This route adds a new variable to a specific project in the vault. The project name and variable name must be specified in the URL, and the variable details should be provided in the request body in JSON format.

### **PATCH /project/:project/var/:variable**

This route modifies an existing variable in a specific project in the vault. The project name and variable name must be specified in the URL, and the update details should be provided in the request body in JSON format.

### **DELETE /project/:project/var/:variable**

This route deletes a variable from a specific project in the vault. The project name and variable name must be specified in the URL.

Please note that to access these routes, you need to ensure that authentication is successfully performed using the **`AuthMiddleware`** middleware.

This concludes the API documentation. For any further questions, please contact the developer.