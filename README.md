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

