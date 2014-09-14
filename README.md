<img src="https://raw.githubusercontent.com/motdotla/signature-api/master/signature-api.jpg" alt="signature-api" width="200" />

# signature-api

API platform for signing documents. Works in tandem with [signature-worker](https://github.com/motdotla/signature-worker).

```
curl http://signature-api.herokuapp.com/api/v0/documents/create.json?url=http://mot.la/assets/resume.pdf
```

View the [documentation](http://docs.signatureapi.apiary.io/).

## Installation

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

### Development

```
git clone https://github.com/motdotla/signature-api.git
cd signature-api
go get 
cp .env.example .env
go run app.go
```

Edit the contents of `.env`.
