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
        - When a command is started via `Start()` it will be given a context for termination for the `Stop()` function. A default and configurable timeout for jobs would likely be a good feature to add in the future
    - Tests:
        - Simple unit tests for all functions, may also need to write a private `clear()` function to reset the map of jobs `jobMap` between tests. 
2. **Server**
    -  API:
        - Simple API with the prefix `"/jobs/"`. So if the server is running on `localHost:8443`, the endpoint for start would be `localHost:8443/jobs/start`
        - Four endpoints: `/jobs/start`, `/jobs/stop/{id}`, `/jobs/status/{id}`, `/jobs/output/{id}`
        - `"/jobs/"` prefix included to delineate this API from other possible future tasks on server
    -  Data Structures: 
        -  `Token` struct with fields for `Token string` and `Role string`
        -  `tokenMap map[string]Token` a hardcoded map of `Token` structs that are set to default values for this project
        -  `rolePermissions map[string]map[string]bool` a hardcoded map to authorize commands based on `Token.Role` field. 3 roles, `admin`, `user`, `viewer` with varying priveleges. This data structure will map roles to a map of `action string -> isAuthz bool` that can be performed by the role
        -  Server file may have various structs for sending and receiving JSON to and from client
        -  `mux` a multiplexer to route API using `http.NewServeMux()`
    - Functions: 
        - `handleStartJob(w http.ResponseWriter, r *http.Request)`, `handleStopJob()`, `handleGetJobStatus()`, `handleGetJobOutput()` Helper functions to handle the job library functions. Errors are handled using the `ResponseWriter` 
        - `authenticate(r *http.Request) (string, bool)` Authenticate function which extracts the token and checks if it is in the `tokenMap`
        - `authorize(username, task string) bool` Performs a check which compares the Token's `Role` field to the allowed roles in the `rolePermissions` map and compares that to the input variable `task` 
        - `authMiddleware(action string)` Extracts token from header, checks if it is authenticated, then checks if it is authorized to perform `action` and then calls the next route
    -  Authentication:
        - This project will use Bearer Token authentication by checking the HTTP header's Bearer field for a `Token` and finding the `Token` in `tokenMap`
    - Authorization:
        - The server will maintain a hardcoded `rolePermissions` map in the format `{Permission Level : {job : isAllowed}}`, for example `{"admin" : {"start" : true, "stop" : true, "status" : true, "output : true"}}, {"user" : {"start" : true, "stop" : false, "status" : true, "output" : true}}` 
        - Every issued command will be authorized via a helper function based on the `Token`, which will retrieve `Token.Role` field from `tokenMap` and check if the command to be executed is in the corresponding `rolePermissions` map  
    - Auth Middleware:
        - The multiplexer will run middleware on every request before routing it to the handleJob helper functions. There will be four routes (for each of the four endpoints) and the authMiddleware will be hardcoded with the appropriate `action` to validate if user can perform action. Shorthand example for route `/jobs/start` : `authMiddleware('start')(http.HandlerFunc(handleStartJob))`
        - To only parse the Token from the header once, the authorization and authentication will be combined in to one middleware function  
    - HTTPS:
        - Generate and hard code self signed certificate and key and put that in the repo, the client will set `tls.Config.InsecureSkipVerify` to `true`
    - Auth Workflow: After TLS handshake, client makes request to server with Bearer Token credentials in HTTP header, server multiplexer directs route and uses middleware to authenticate -> authorize -> execute command, then returns response to client. 
    - Error handling: Errors will be descriptive and passed back to the client using the http Writer
3. **Client**
    - Data Structures:
    - Functions:
    - UX:
## Scope 

## Proposed API 

## Security considerations,

## CLI UX