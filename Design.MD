# JobWorker Design Doc 

## Design approach 
Three components: CLI, Server, and Worker Library.
The server will wrap the Worker Library to communicate with the CLI:

`CLI -> <- Server <- [Library]`

### Component Details

1. **Worker Library**
    - Data Structures: 
        - `Job` struct with various fields (Job ID, command, status, output, etc). Will contain private fields for Context, Cmd and CancelFunc (from `exec` and `context` packages) to handle stopping commands and deal with timeouts for tasks
        - `jobMap`, `jobLock`: Two variables: first holds a mapping from `ID` to `Job` `map[string]*Job`, the second is a read/write lock `sync.RWMutex` to prevent concurrent access to the map. 
    - Functions to perform tasks from the API (all public):
        - `Start()`, `Stop()`, `Status()`, `Output()`
    - Notes:
        - Will use a UUID package for the JOB ID. 
        - When a command is started via `Start()` it will be given a context for timeout/deadline and termination signal for the `Stop()` function. The default timeout will be set to 1 minute and will not be adjustable from the CLI
    - Tests:
        - Simple unit tests for all functions, may also need to write a private `clear()` function to reset the map of jobs `jobMap` between tests. 
2. **Server**
    -  API:
        - Simple API with the prefix `"/jobs/"`. So if the server is running on `localHost:8443`, the endpoint for start would be `localHost:8443/jobs/start`
        - Four endpoints: `/jobs/start`, `/jobs/stop/{id}`, `/jobs/status/{id}`, `/jobs/output/{id}`
        - `"/jobs/"` prefix included to delineate this API from other possible future tasks on server
    -  Data Structures: 
        -  `User` struct with fields for `Username`, `Password`, and `Role`
        -  `userMap map[string]User` a hardcoded map of `User` structs that are set to default values for this project
        -  `rolePermissions map[string]map[string]bool` a hardcoded map to authorize commands based on `User.Role` field. 3 roles, `admin`, `user`, `viewer` with varying priveleges. This data structure will map roles to a map of actions
        -  Server file may have various structs for sending and receiving JSON to and from client
        -  `mux` a multiplexer to route API using `http.NewServeMux()`
    - Functions: 
        - `handleStartJob(w http.ResponseWriter, r *http.Request)`, `handleStopJob()`, `handleGetJobStatus()`, `handleGetJobOutput()` Helper functions to handle the job library functions. Errors are handled using the `ResponseWriter` 
        - `basicAuthenticate(r *http.Request) (string, bool)` Authenticate function using the `bcrypt` library which checks if user exists, hashes incoming http password, compares hashed password to stored hashed password, and returns (username, ok)
        - `authorize(username, task string) bool` Performs a simple check which compares the user's role to the allowed roles in the permission: `userMap[username].Role`  
        - `authenticateMiddleware(next http.Handler) http.Handler` Authenticates user and adds username as context to `http.Request` to be used by authorization middleware
        - `authorizationMiddleware(action string)` Authorizes user (username passed using http Request context) to perform action
    -  Authentication:
        - This project will use Basic authentication by checking the HTTP header's Basic field for a username, finding the `User` in `userMap`, passing the password through a cryptographic hash function using bcrypt library, and comparing that to the stored user's password hash
        - Ideally username/password would be stored in a database with the password salted + hashed, but in this project demo user's password hashes will be premade and hardcoded (to match the user's hashed password).
    - Authorization:
        - The server will maintain a hardcoded `rolePermissions` map in the format `{Permission Level : {job : isAllowed}}`, for example `{"admin" : {"start" : true, "stop" : true, "status" : true, "output : true"}}, {"user" : {"start" : true, "stop" : false, "status" : true, "output" : true}}` 
        - Every issued command will be authorized via a helper function based on the user's username: retrieve `User.Role` field from `userMap[username]` and check if the command to be executed is in the corresponding `rolePermissions` map  
    - Auth Middleware:
        - The multiplexer will run middleware on every request before routing it to the handleJob helper functions. There will be four routes (for each of the four endpoints) and the authorizationMiddleware will be hardcoded with the appropriate `action` to validate if user can perform action. Shorthand example  for route `/jobs/start` : `authenticateMiddleware(authorizeMiddleware('start')...handleStartJob)`
    - HTTPS:
        - Generate and hard code self signed certificate and a key and put that in the repo, the client will set `tls.Config.InsecureSkipVerify` to `true`
    - Auth Workflow: After TLS handshake, client makes request to server with Basic Authentication credentials in HTTP header, server multiplexer directs route and uses middleware to authenticate -> authorize -> execute command, then returns response to client. 
    - Error handling: Errors will be descriptive and passed back to the client using the http Writer
3. **Client**

## Scope 

## Proposed API 

## Security considerations,

## CLI UX