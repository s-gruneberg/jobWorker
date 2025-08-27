# JobWorker Design Doc
![Diagram](./images/Diagram.png "Design Approach Diagram")
## Proposed API
- Four endpoints: `/jobs/start`, `/jobs/stop/{id}`, `/jobs/status/{id}`, `/jobs/output/{id}`
- If the server is running on `https://localHost:8443`, the endpoint for start would be `https://localHost:8443/jobs/start`
- `"/jobs/"` prefix included to delineate this API from other possible future tasks on server

## Component Details
1. **Worker Library**
    - Data Structures: 
        - `Job` struct with various fields (Job ID, command, status, output, cancel, etc)
        - `jobMap`, `jobLock`: Mapping from `ID` to `Job` `map[string]*Job`, and a read/write lock `sync.RWMutex` 
    - Functions to perform tasks from the API (all public):
        - `Start()`, `Stop()`, `Status()`, `Output()`
    - Notes:
        - Will use a UUID package for the JOB ID. 
        - When a command is started via `Start()` it will be given a context for termination for the `Stop()` function. 
2. **Server**
    -  Data Structures: 
        -  `Token` struct with fields for `Token string` and `Role string`
        -  `tokenMap map[string]Token` a hardcoded map of `Token` structs that are set to default values
        -  `rolePermissions map[string]map[string]bool` a map to authorize commands based on `role`. 3 roles: `admin`, `user`, `viewer`
        -  `commandPermissions map[string][]string` maps `role` to a list of linux commands that role can run
        -  `mux` a multiplexer to route API using `http.NewServeMux()`
    - Functions: 
        - `handleStartJob(w http.ResponseWriter, r *http.Request)`, `handleStopJob()`, `handleGetJobStatus()`, `handleGetJobOutput()`
        - `authenticate(r *http.Request) (string, bool)` Authenticate function which extracts the token and checks if it is in the `tokenMap`
        - `authorizeAction(role, action string) bool` Authorizes if a user can perform an action (start, stop)
        - `authorizeCommand(role, command string) bool` Checks if a linux command is allowed for a given role
        - `authMiddleware(action string)` Authenticates user and authorizes action (start, stop, output)
    -  Authentication:
        - Use Bearer Token authentication by checking the HTTP header's Bearer field for a `Token` and finding the `Token` in `tokenMap`
    - Authorization:
        - Two separate authorizations: action (start, stop, status) and command (ls, grep, rm)
        - The server will maintain a `rolePermissions` map in the format `{Permission Level : {job : isAllowed}}` for task authorization
        - The linux command authorization will only be checked if a job is starting, so that will be authorized inside `handleStartJob()` and checked via map `comandPermissions`
    - Auth Middleware:
        - The authMiddleware function will be hardcoded with the appropriate `action` to validate if user can perform action. Shorthand example for route `/jobs/start` : `authMiddleware('start')(http.HandlerFunc(handleStartJob))`
    - HTTPS:
        - Generate and hard code self signed certificate and key and put that in the repo, the client will set `tls.Config.InsecureSkipVerify` to `true`
    - Auth Workflow Overview: 
        - After TLS handshake, client makes request to server with Bearer Token credentials in HTTP header, server multiplexer directs route and uses middleware to authenticate -> authorize -> authorize linux command -> execute command, then returns response to client. 
    - Notes:
        - Decided to include both role based authz (viewer cannot stop job or start job) as well as command based authz (regular user cannot run `rm` but can start other jobs). 
3. **Client**
    - Data Structures:
        - `Client` type that will hold the user's `token`, the `baseURL`, and a pointer to an `http.Client`
        - Possibly structs for parsing JSON to and from client
    - Functions: Most functions will act on the `Client` type as a method
        - `MakeClient` called in main with parsed command line arguments `-token` and `-server` to create a new client. This sets up bypassing TLS verification as well (since the server is using self-signed certs)
        - `(c *Client) StartJob(command string, args []string) (string, error)` starts a job with given parameters
        - `StopJob(id string)` stops the job with id 
        - `GetStatus(id string)` gets status of job with id
        - `GetOutput(id string)` gets output of job with id
        - `makeRequest()` helper function that builds and sends a request to the client 
    - Workflow: (Assumes server is running)
        - `main()` makes a new client, sets flagged variables or defaults to admin token and default server path `https://localhost:8443` (for demo)
        - Parses args
        - Uses a switch statement for the following 4 commands for args[0]: start, stop, status, output
        - Sends request to server based
        - Returns a confirmation to the user of some sort if there is not output

## Security Considerations
- Uses self signed certificates, and client bypasses verification (for the demo)
- Bearer tokens are stored unencrypted in memory, they never expire or rotate
- User can run any linux command that their command privilege level allows 
- There is no input validation or sanitization, potential for injection attacks.  
- No sandbox or containerization on the server, client can access server's file system
- No user logging (can't tell who ran what command -> no accountability)
- No job timeout or job resource limitation

## CLI UX   
```
            $~jobWorker start echo "Hello World"
            Job with ID 4c20a846-a780-4378-88bc-2447eb072811 started
            $~jobWorker start sleep 30
            Job with ID bf779987-12cf-41a8-b991-21045cd5d822 started
            $~jobworker stop bf779987-12cf-41a8-b991-21045cd5d822
            Job stopped
            $~jobWorker output 4c20a846-a780-4378-88bc-2447eb072811
            Hello World
            $~jobWorker status 4c20a846-a780-4378-88bc-2447eb072811
            Succeeded
            $~jobWorker stop bf779987-12cf-41a8-b991-21045cd5d822
            Error: Job is not running           
```

