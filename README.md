# di

Minimalistic, reflection-based Dependency Injection (DI) framework for GoLang.

Inspired by https://github.com/ivankorobkov/go-di

## Features

* ✅ Constructor-based registration
* ✅ Automatic dependency resolution via reflection
* ✅ Support for interfaces (implementation matched automatically)
* ✅ Manual argument injection for primitives or configs
* ❌ No lazy-loading of constructors — instances created when first resolved

## Installation

```bash
go get github.com/rom8726/di
```

## Usage

### 1. Define your interfaces and implementations

```go
type DBClient interface {
	Exec() (string, error)
}

type Repo interface {
	Find() (string, error)
}

type RepoImpl struct {
	db DBClient
}

func (r *RepoImpl) Find() (string, error) {
	return r.db.Exec()
}

func NewRepo(db DBClient) *RepoImpl {
	return &RepoImpl{db: db}
}
```

### 2. Register constructors

```go
c := di.New()
c.Provide(NewRepo)
c.Provide(NewDBClient)
```

### 3. Resolve instances

```go
var repo Repo
err := c.Resolve(&repo)
if err != nil {
	log.Fatal(err)
}
```

### 4. Inject configuration

```go
params := &MyServiceParams{ParamInt: 42, ParamStr: "hello", ParamBool: true}
c.Provide(NewMyService).Arg(params)
```

Or for primitives:

```go
c.Provide(NewMyService2).Args(123, true)
```

## Example

See example in unit tests.

## Design Principles

* **Simplicity**: No code generation, no additional interfaces to implement.
* **Reflection**: Uses `reflect` to resolve dependencies at runtime.
* **Predictability**: Always construct dependencies top-down, respecting constructor order.
* **Safety**: Panics on duplicate provider types.

## Limitations

* Only supports singleton resolution (no scoped or transient lifetimes).
* No automatic scanning — you must explicitly register each constructor.

## Testing

The package includes tests that cover:

* Interface resolution
* Manual parameter injection
* Complex constructor trees

Run:

```bash
go test ./...
```

## App

Using the `App` struct, you can create an application that manages the lifecycle of services within a `Container`.
Services that implement the `Servicer` interface are automatically detected and handled.
This enables clean and consistent startup and shutdown sequences for your application.

### Servicer Interface
```go
type Servicer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

### Example Usage

Here's how you can integrate `App` into your application:

```go
package main

import (
    "fmt"
    "log"

    "github.com/rom8726/di"
)

// Define a service that implements the Servicer interface
type MyService struct {
    // Dependencies
}

func NewMyService() *MyService {
    return &MyService{}
}

func (s *MyService) Start() error {
    fmt.Println("MyService started")
    return nil
}

func (s *MyService) Stop() error {
    fmt.Println("MyService stopped")
    return nil
}

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Create a new di Container
    c := di.New()

    // Provide your services to the Container
    c.Provide(NewMyService)

    // Create a new App with the Container
    app := di.NewApp(c)

    // Start the application
    if err := app.Run(ctx); err != nil {
        log.Fatalf("Failed to start app: %v", err)
    }
}
