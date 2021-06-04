# statusrepeater

[![Build Status](https://img.shields.io/github/workflow/status/olvlvl/go-status-repeater/test)](https://github.com/olvlvl/go-status-repeater/actions?query=workflow%3Atest)
[![Coverage Status](https://coveralls.io/repos/github/olvlvl/go-status-repeater/badge.svg?branch=main)](https://coveralls.io/github/olvlvl/go-status-repeater?branch=main)

A middleware that repeats the status code of a matching request without processing the request again.

The middleware can be useful when indelicate clients keep querying your API for resources that don't exist. After the
first 404, similar requests will get a 404 right away from the middleware, for a given duration. The middleware can be
used to repeat any status code, but you probably want to limit it to 400 and 404.

Here's an example:

```go
handler := statusrepeater.Handler(
    next,                            // The handler you want to decorate
    http.StatusNotFound,             // The status code you want to repeat
    statusrepeater.DefaultDuration,  // The duration of the repeat
    statusrepeater.DefaultFormatKey, // A function that creates a key from a request
)
```

## License

This software is under the [BSD 3-Clause](LICENSE). 
