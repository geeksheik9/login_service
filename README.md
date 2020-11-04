# Login Service

- This application is designed to allow users to be registered, logged in, and view their information.
- Current application supports register, login, and get profile
- TODO: Add support for adding roles. Add support for deleting a user

## Deploy

### Local

```shell
go run ./main/main.go
```

- Will run the application locally at port 3000

### Local Docker Container

```shell
docker-compose up
```

- Will run the application in a docker container at port 3000
- Exit using `control(^) + c`

### TODO: Create deployment scripts for any location

## Config

- PORT
- USER_DATABASE
- USER_COLLECTION
- LOG_LEVEL

## Routes

### Health Information

- **GET** /ping

  - checks if the pod is running

- **GET** /health

  - check that the pod and its dependencies are running

- **POST** /register

  - function name: RegisterUser
  - Creates a user in the database
  - User information passed in the body:

    ```shell
        {
            "username":"user",
            "password":"pass",
            "firstName":"first",
            "lastName" :"last"
        }
    ```

- **POST** /login

  - function name: LoginUser
  - Compares information passed to database to log a user in
  - User information passed in the body:

    ```shell
    {
        "username":"user"
        "password":"pass"
    }
    ```

- **GET** /profile

  - function name: GetUserProfile
  - returns information in a user profile based on a JWT
  - JWT passed in at the authorization header level following format:
    - Authorization: Bearer {{token}}

### Swagger

- **GET** /swagger/

  - serves up swagger UI
