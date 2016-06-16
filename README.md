# transfig
A Go package that provides the ability to load a JSON config settings file into a complex object graph, and apply transformations to specific settings based on the environment (dev, staging, live etc.), similar to ASP.NET Web.Config files. Also supports live reloading.

## Installation

    go get github.com/sironfoot/transfig

## Setup

Create a ```config.json``` configuration file in the root of your project. This will be the **primary** config file:

```json
{
    "database": {
        "driverName": "postgres",
        "connectionString": "user=user dbname=myDB sslmode=disable"
    },

    "appSettings": {
        "encryptionKey": "DONKEY_RHUBARB13",
        "recordsPerPage": 20
    },

    "emailSettings": {
        "testMode": false,
        "testEmail": "info@example.com",
        "smtpUsername": "sendgrid_username",
        "smtpPassword": "sendgrid_password"
    }
}
```

Create the following structs in Go:

```go
type Configuration struct {
    Database      DatabaseConfig `json:"database"`
    AppSettings   AppConfig      `json:"appSettings"`
    EmailSettings EmailConfig    `json:"emailSettings"`
}

type DatabaseConfig struct {
    DriverName       string `json:"driverName"`
    ConnectionString string `json:"connectionString"`
}

type AppConfig struct {
    EncryptionKey  string `json:"encryptionKey"`
    RecordsPerPage int    `json:"recordsPerPage"`
}

type EmailConfig struct {
    TestMode     bool   `json:"testMode"`
    TestEmail    string `json:"testEmail"`
    SMTPUsername string `json:"smtpUsername"`
    SMTPPassword string `json:"smtpPassword"`
}
```

Now lets create an **environment** specific config.json file. By convention, these must be in the format ```whatever.environment.json```, so create a file called ```config.dev.json```

```json
{
    "database": {
        "connectionString": "user=username password=password dbname=devDb sslmode=disable"
    },

    "emailSettings": {
        "testMode": true,
        "testEmail": "jane.smith@example.com"
    }
}
```

Make sure to add ```config.*.json``` to your ```.gitignore``` file.

The idea is that individual developers can have their own dev environment config file with settings specific to their machine/environment, without accidentally committing these into source control. You could also create ```config.staging.json``` and ```config.live.json``` environment config files that sit on the staging and live servers, and avoid having sensitive live/production config settings (such as encryption keys) accidentally committed into source control.

The environment specific config file needs to use the same structure as the primary config file, but only needs to include the properties it wants to change. If a property is missing from the environment config file, the **primary** one is used.

Given the **primary** and **environment** config files listed above, the final output would be:

```json
{
    "database": {
        "driverName": "postgres",
        "connectionString": "user=username password=password dbname=devDb sslmode=disabl"
    },

    "appSettings": {
        "encryptionKey": "DONKEY_RHUBARB13",
        "recordsPerPage": 20
    },

    "emailSettings": {
        "testMode": true,
        "testEmail": "jane.smith@example.com",
        "smtpUsername": "sendgrid_username",
        "smtpPassword": "sendgrid_password"
    }
}
```

FInally, in your ```main.go``` file:

```go
import github.com/sironfoot/transfig

func main() {
    environment := "dev" // populate this dynamically (e.g. using 'flags')

    var config Configuration
    err := transfig.Load("config.json", environment, &config)
}
```

## Live Reloading

transfig supports caching and live reloading of configuration files, so you can update the configuration file without having to restart the Go program.

```go
var config Configuration
err := transfig.LoadWithCaching("config.json", environment, &config)
```

transfig checks for changes to the config files every 5 seconds.

LoadWithCaching is thread-safe, and each call to LoadWithCaching gets it's own copy of the config object, so different threads could edit the properties of their own config copy without affecting each other.

## Use In a Web Application

The best place to use transfig is in middleware so your config object is available to all your HTTP Handlers. Here is an example from goji.io:

```go
router.UseC(func(next goji.Handler) goji.Handler {
    return goji.HandlerFunc(func(ctx context.Context, res http.ResponseWriter, req *http.Request) {

        var config Configuration
        err := config.LoadWithCaching("config.json", environment, &config)
        if err != nil {
            panic(err)
        }

        ctx = context.WithValue(ctx, "config", &config)

        next.ServeHTTPC(ctx, res, req)
    })
})

router.HandleC(pat.New("/users"), func(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	config := ctx.Value("config").(*Configuration))

    // TODO: load Users (snip)...
})
```
