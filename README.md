# Account API

## Overview
This is a small Microservice example of simple Shopping Account API endpoint to be used with [Shipped](http://shipped-cisco.com).

## Getting Started
1. Simple clone the repo or download the repo and cd into the repo and do `go run account.go` to start the server on localhost.
2. Then run this simple curl command to make sure the service is up and running `curl -i http://localhost:8000` or view it on your web browser

```
Expected Result

HTTP/1.1 200 OK
Date: Tue, 19 Jan 2016 15:52:02 GMT
Content-Length: 160
Content-Type: text/html; charset=utf-8

<html>
  <head>
    <title>An example layout</title>
  </head>
  <body>

<p>Account is up and running.</p>
<p>Try This command</p>
<p></p>

  </body>
</html>

```

## Endpoints

- **POST /v1/account/**
```
curl -i -H "Content-Type: application/json" -X POST -d '{"username":"nick","password":"123"}' http://localhost:8000/v1/account/
Expected Result:
{
    "success":true,
    "token": 12325343,
    "message":"Successfully logged in"
}
or curl -H "Content-Type: application/json" -X POST -d @loginExamplePost.json http://localhost:8000/v1/account/
```
- **GET /v1/account/**
```
curl -i -H "Content-Type: application/json" -X GET http://localhost:8000/v1/account/
Expected Result:
{
    "success":true,
    "message":"Successfully logged out in"
}
```

## Requirements
* [Go](https://github.com/golang/example)

## Credits
- [Nick Hayward](https://github.com/nehayward)
