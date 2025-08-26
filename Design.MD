# JobWorker Design Doc 

## Design approach 
Three components: CLI, Server, and Worker Library.
The server will wrap the Worker Library to communicate with the CLI:

`CLI -> <- Server <- [Library]`

### Component Details

1. **Worker Library**
    - Data Structures: 
        - `Job` struct with various fields (Job ID, command, user, etc). Will contain private fields for Context, Cmd and CancelFunc (from `exec` and `context` packages) to handle stopping commands and deal with timeouts for tasks
        - `jobMap`, `nextID`, `jobLock`: Three variables: first holds a mapping from `ID` to `Job` struct `map[string]*Job`, second is an integer to produce unique job ID's, and the third is a read/write lock `sync.RWMutex` to prevent concurrent access to these variables. 
    - Functions to perform tasks from the API (all public):
        - `Start()`, `Stop()`, `Status()`, `Output()`
    - Notes:
        - Will use incrementing ID's to uniquely identify Jobs rather than UUID's (This is to avoid importing Google's UUID package, but I do think UUID's would be the choice for production. In the interest of simplicity and readability, I went with an integer counter).
        - When a command is started via `Start()` it will be given a context for timeout/deadline and termination signal from the `Stop()` function. The default timeout will be set to 1 minute and will not be adjustable from the CLI
    - Tests:
        - Simple unit tests for all functions, may also need to write a private `clear()` function to reset the map of jobs `jobMap` between tests. 
2. **Server**
    -  API:
        - Simple API with the prefix `"/jobs/"`. So if the server is running on localHost:8443, the endpoint for start would be `localHost:8443/jobs/start`
        - Four endpoints: `/jobs/start`, `/jobs/stop/{id}`, `/jobs/status/{id}`, `/jobs/output/{id}`
        - `"/jobs/"` prefix included to delineate this API from other possible future tasks on server.
    -  Data Structures: 
        -  `User` struct with fields for `Username`, `Password`, and `Role`
        -  `userMap map[string]User` a hardcoded map of `User` structs that are set to default values for this project, 
        -  `rolePermissions map[string]map[string]bool` a hardcoded map to authorize commands based on `User.Role` field. 3 roles, `admin`, `user`, `viewer` with varying priveleges. This data structure will map roles to a map of actions. 
        -  Server file may have various structs for sending and receiving JSON to and from client
        -  `mux` a multiplexer to route API using `http.NewServeMux()`
    - Functions: 
        - `handleStartJob(w http.ResponseWriter, r *http.Request)`, `handleStopJob()`, `handleGetJobStatus()`, `handleGetJobOutput()` Helper functions to handle the job library functions. Authentication and authorization will be done within these functions. The handler functions are only called by the multiplexer based on the URL, and errors are handled using the `ResponseWriter` 
        - `hashPassword(password string) (string, bool)` Cryptographic password hash function (using crypto library `bcrypt`) 
        - `authenticate(r *http.Request) (string, bool)` Authenticate function using the `bcrypt` library which checks if user exists, hashes incoming http password, compares hashed password to stored hashed password, and returns (username, ok)
        - `authorize(username, task string) bool` Authorization function that runs after the authentication function (authenticate returns a username string). Performs a simple check which compares the user's role to the allowed roles in the permission: `userMap[username].Role` 
        - `setUpRoutes()` Helper function to set up routes for multiplexer 
    -  Authentication:
        - This project will use Basic authentication by checking the HTTP header's Basic field for a username, finding the `User` in `userMap`, passing the password through a cryptographic hash function using bcrypt library, and comparing that to the stored user's password hash
        - Ideally username/password would be stored in a database with the password salted + hashed, but in this project demo users password hashes will be premade and hardcoded (to match the user's hashed password).
        - To further clarify previous point, let's say admin is a type `User` with fields `{Username = "admin", Password = hashedPassword, Role = "admin"}`. When admin tries to authenticate, their username and literal password will be sent in the header. The username is found in the `userMap` and the string password is hashed and compared to the stored hashed password. 
        - Authentication will be called before Authorization in each `handle...()` function, like `handleStartJob()`
    - Authorization:
        - The server will maintain a hardcoded `rolePermissions` map in the format `{Permission Level : [List of allowed functions]}`, for example `{"admin" : ["start", "stop", "status", "output"]}, {"user" : ["start", "status", "output"]}`. 
        - Every issued command will be authorized via a helper function based on the user's username: retrieve `User.Permission` field from `userMap[username]` and check if the command to be executed is in the `permissions` map  
    - HTTPS:
        - Generate and hard code self signed certificate and a key and put that in the repo, the client will set `tls.Config.InsecureSkipVerify` to `true`
    - Auth Workflow: After TLS handshake, client makes request to server with Basic Authentication credentials in HTTP header, server calls appropriate helper function for API endpoint, helper function authenticates client via `authenticate()` function, helper function , server executes command and returns response to client. 
    - Error handling: Errors will be descriptive and passed back to the client.
3. **Client**
## Scope 

## Proposed API 

## Security considerations,

## CLI UX