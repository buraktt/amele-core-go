# amele-core-go

A Go package for handling communication in amele agents. It supports both TCP and shared memory (shmem) protocols for accepting inputs, calling functions, and responding to orchestrators.

## Installation

```bash
go get github.com/buraktt/amele-core-go
```

## Usage

Set environment variables based on the communication protocol:

- For TCP: Set `COMMUNICATION_PROTOCOL=tcp` and `AMELE_TCP_PORT=<port>`
- For shmem: Set `COMMUNICATION_PROTOCOL=shmem`, `AMELE_INBOX_FILE=<path>`, and `AMELE_OUTBOX_FILE=<path>`

```go
import "github.com/buraktt/amele-core-go"

func main() {
    // Accept inputs from the orchestrator
    inputs, err := amelecore.Accept()
    if err != nil {
        panic(err)
    }

    // Get the stored context
    ctx := amelecore.Context()

    // Call another function (TCP mode only)
    result, err := amelecore.CallFunction("someFunction", map[string]any{"key": "value"})
    if err != nil {
        // Handle error
    }

    // Respond with updated context
    err = amelecore.Respond(map[string]any{"status": "done"})
    if err != nil {
        panic(err)
    }
}
```

## License

[Add license information here]
