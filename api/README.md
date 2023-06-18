# **Documentation de l'API**

L'API fournit une interface pour interagir avec un "vault". Voici la structure de base du code de l'API :

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

La documentation ci-dessous détaille les différentes routes et fonctionnalités offertes par cette API.

## **Middlewares**

### **CorsMiddleware**

Le middleware **`CorsMiddleware`** permet de gérer les problèmes de CORS (Cross-Origin Resource Sharing) en ajoutant les en-têtes appropriés aux réponses. Il permet aux requêtes provenant de différents domaines d'accéder aux ressources de l'API.

### **AuthMiddleware**

Le middleware **`AuthMiddleware`** est responsable de l'authentification des clients avec le "vault". Il utilise les variables d'environnement **`ADDRESS_VPS_VAULT`** et **`TOKEN_VPS_VAULT`** pour se connecter au "vault". Une fois authentifié, il stocke le client "vault" dans le contexte de la requête pour une utilisation ultérieure.

## **Routes**

### Body (JSON)

```json
{
	"Value":"Remplacez par la valeur attendu"
}
```

### **GET /project/**

Cette route permet de récupérer la liste des projets depuis le "vault".

### **POST /project/**

Cette route permet de créer un nouveau projet dans le "vault". Les détails du projet doivent être fournis dans le corps de la requête au format JSON.

### **PATCH /project/:project/**

Cette route permet de renommer un projet existant dans le "vault". Le nom du projet doit être spécifié dans l'URL et les détails de mise à jour doivent être fournis dans le corps de la requête au format JSON.

### **DELETE /project/:project/**

Cette route permet de supprimer un projet du "vault". Le nom du projet doit être spécifié dans l'URL.

### **GET /project/:project/**

Cette route permet de récupérer les détails d'un projet spécifique à partir du "vault". Le nom du projet doit être spécifié dans l'URL.

### **GET /project/:project/var/**

Cette route permet de récupérer la liste des variables associées à un projet spécifique à partir du "vault". Le nom du projet doit être spécifié dans l'URL.

### **POST /project/:project/var/:variable**

Cette route permet d'ajouter une nouvelle variable à un projet spécifique dans le "vault". Le nom du projet et le nom de la variable doivent être spécifiés dans l'URL, et les détails de la variable doivent être fournis dans le corps de la requête au format JSON.

### **PATCH /project/:project/var/:variable**

Cette route permet de modifier une variable existante dans un projet spécifique du "vault". Le nom du projet et le nom de la variable doivent être spécifiés dans l'URL, et les détails de mise à jour doivent être fournis dans le corps de la requête au format JSON.

### **DELETE /project/:project/var/:variable**

Cette route permet de supprimer une variable d'un projet spécifique dans le "vault". Le nom du projet et le nom de la variable doivent être spécifiés dans l'URL.

Veuillez noter que pour accéder à ces routes, vous devez vous assurer que l'authentification est effectuée avec succès via le middleware **`AuthMiddleware`**.

Ceci conclut la documentation de l'API. Pour toute question supplémentaire, veuillez contacter le développeur.