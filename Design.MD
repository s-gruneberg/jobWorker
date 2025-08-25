# JobWorker Design Doc 

## Design approach 
Three components: CLI, Server, and Worker Library.
The server will wrap the Worker Library to communicate with the CLI:
```
CLI -> <- Server [Library]
```
### Component Details
1. **Worker Library**
    - Defines a job struct with various fields (Job ID, command, user, etc)
    - Instatiates two private variables: a map to hold jobs `{JobID -> JobStruct}` and a read/write write lock for the map `sync.RWMutex`
    - Has four public functions corresponding to the necessary API calls: Start, Stop, Status, Output
2. **Server**
    - 
## Scope 

## Proposed API 

## Security considerations,

## CLI UX