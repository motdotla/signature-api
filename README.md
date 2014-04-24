# signature-api

E-Signing API.

The signature-api is based on [CONTRA]() API design. [JSON](http://www.json.org) is returned in all responses from the API, including errors.

I've tried to make it as easy to use as possible, but if you have any feedback please [let me know](mailto:scott@scottmotte.com).

## Installation

### Development

```
git clone https://github.com/scottmotte/signature-api.git
cd signature-api
go get 
cp .env.example .env
go run app.go
```

## Summary

### API Endpoint

* [http://signature-api.herokuapp.com/api/v0](http://signature-api.herokuapp.com/api/v0)

### Documents

#### /documents/create

Create a signable document.

##### Definition

```
POST|GET http://signature-api.herokuapp.com/api/v0/documents/create.json?url=[url]
```

##### Parameters

* url*

##### Example Request

```
http://signature-api.herokuapp.com/api/v0/documents/create.json?url=http://scottmotte.com/assets/resume.pdf
```

##### Example Response

```
{
  "document": {
    "pages": [ ]
    "status": "unprocessed",
    "url": "http://scottmotte.com/assets/resume.pdf",
  },
  "success": true
}
```

##### Example Error

```
{
  "error": {
    "message": "This is the error message"
  }
  "success": false
}
```
