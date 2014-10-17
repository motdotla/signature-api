package main

import (
	"github.com/go-martini/martini"
	"github.com/handshakejs/handshakejserrors"
	"github.com/joho/godotenv"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"github.com/motdotla/signaturelogic"
	"log"
	"net/http"
	"os"
)

const (
	LOGIC_ERROR_CODE_UNKNOWN = "unknown"
)

var (
	CARVE_ROOT             string
	SIGNATURE_CATCHER_ROOT string
	SIGNATURE_SHARE_ROOT   string
	SIGNATURE_API_ROOT     string
	ORCHESTRATE_API_KEY    string
)

type Document struct {
	Url string `form:"url" json:"url"`
}

type Signing struct {
	DocumentUrl string `form:"document_url" json:"document_url"`
}

type SignatureElement struct {
	X          string `form:"x" json:"x"`
	Y          string `form:"y" json:"y"`
	Url        string `form:"url" json:"url"`
	PageNumber string `form:"page_number" json:"page_number"`
	SigningId  string `form:"signing_id" json:"signing_id"`
}

type TextElement struct {
	X          string `form:"x" json:"x"`
	Y          string `form:"y" json:"y"`
	Content    string `form:"content" json:"content"`
	PageNumber string `form:"page_number" json:"page_number"`
	SigningId  string `form:"signing_id" json:"signing_id"`
}

func main() {
	loadEnvs()

	signaturelogic.Setup(ORCHESTRATE_API_KEY)

	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Content-Length", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	m.Use(render.Renderer())

	m.Any("/api/v0/documents/create.json", binding.Bind(Document{}), DocumentsCreate)
	m.Any("/api/v0/documents/:id.json", DocumentsShow)
	m.Any("/api/v0/signings/create.json", binding.Bind(Signing{}), SigningsCreate)
	m.Any("/api/v0/signings/:id.json", SigningsShow)
	m.Any("/api/v0/signature_elements/create.json", binding.Bind(SignatureElement{}), SignatureElementsCreate)
	m.Any("/api/v0/signature_elements/:id/update.json", binding.Bind(SignatureElement{}), SignatureElementsUpdate)
	m.Any("/api/v0/text_elements/create.json", binding.Bind(TextElement{}), TextElementsCreate)
	//m.Any("/api/v0/text_elements/:id/update.json", binding.Bind(TextElement{}), TextElementsUpdate)

	m.Run()
}

func ErrorPayload(logic_error *handshakejserrors.LogicError) map[string]interface{} {
	error_object := map[string]interface{}{"code": logic_error.Code, "field": logic_error.Field, "message": logic_error.Message}
	errors := []interface{}{}
	errors = append(errors, error_object)
	payload := map[string]interface{}{"errors": errors}

	return payload
}

func DocumentsPayload(document map[string]interface{}) map[string]interface{} {
	document["share_url"] = SIGNATURE_SHARE_ROOT + "?document_url=" + SIGNATURE_API_ROOT + "/api/v0/documents/" + document["id"].(string) + ".json"
	documents := []interface{}{}
	documents = append(documents, document)
	payload := map[string]interface{}{"documents": documents}

	return payload
}

func SigningsPayload(signing map[string]interface{}) map[string]interface{} {
	signings := []interface{}{}
	signings = append(signings, signing)
	payload := map[string]interface{}{"signings": signings}

	return payload
}

func SignatureElementsPayload(signature_element map[string]interface{}) map[string]interface{} {
	signature_elements := []interface{}{}
	signature_elements = append(signature_elements, signature_element)
	payload := map[string]interface{}{"signature_elements": signature_elements}

	return payload
}

func TextElementsPayload(text_element map[string]interface{}) map[string]interface{} {
	text_elements := []interface{}{}
	text_elements = append(text_elements, text_element)
	payload := map[string]interface{}{"text_elements": text_elements}

	return payload
}

func DocumentsCreate(document Document, req *http.Request, r render.Render) {
	_url := document.Url

	params := map[string]interface{}{"url": _url}
	result, logic_error := signaturelogic.DocumentsCreate(params)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		postscript := result["id"].(string)
		requestCarve(_url, postscript)

		payload := DocumentsPayload(result)
		r.JSON(200, payload)
	}
}

func DocumentsShow(params martini.Params, req *http.Request, r render.Render) {
	id := params["id"]
	result, logic_error := signaturelogic.DocumentsShow(id)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := DocumentsPayload(result)
		r.JSON(200, payload)
	}
}

func SigningsShow(params martini.Params, req *http.Request, r render.Render) {
	id := params["id"]
	result, logic_error := signaturelogic.SigningsShow(id)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := SigningsPayload(result)
		r.JSON(200, payload)
	}
}
func SigningsCreate(signing Signing, req *http.Request, r render.Render) {
	document_url := signing.DocumentUrl

	params := map[string]interface{}{"document_url": document_url}
	result, logic_error := signaturelogic.SigningsCreate(params)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := SigningsPayload(result)
		r.JSON(200, payload)
	}
}

func SignatureElementsCreate(signature_element SignatureElement, req *http.Request, r render.Render) {
	x := signature_element.X
	y := signature_element.Y
	_url := signature_element.Url
	page_number := signature_element.PageNumber
	signing_id := signature_element.SigningId

	params := map[string]interface{}{"x": x, "y": y, "url": _url, "page_number": page_number, "signing_id": signing_id}
	result, logic_error := signaturelogic.SignatureElementsCreate(params)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := SignatureElementsPayload(result)
		r.JSON(200, payload)
	}
}

func SignatureElementsUpdate(params martini.Params, signature_element SignatureElement, req *http.Request, r render.Render) {
	id := params["id"]
	x := signature_element.X
	y := signature_element.Y

	payload := map[string]interface{}{"x": x, "y": y, "id": id}
	result, logic_error := signaturelogic.SignatureElementsUpdate(payload)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := SignatureElementsPayload(result)
		r.JSON(200, payload)
	}
}

func TextElementsCreate(text_element TextElement, req *http.Request, r render.Render) {
	x := text_element.X
	y := text_element.Y
	content := text_element.Content
	page_number := text_element.PageNumber
	signing_id := text_element.SigningId

	params := map[string]interface{}{"x": x, "y": y, "content": content, "page_number": page_number, "signing_id": signing_id}
	result, logic_error := signaturelogic.TextElementsCreate(params)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := TextElementsPayload(result)
		r.JSON(200, payload)
	}
}

func requestCarve(document_url string, postscript string) {
	webhook_url := SIGNATURE_CATCHER_ROOT + "/webhook/v0/documents/processed.json"
	carve_url := CARVE_ROOT + "/api/v0/documents/create.json?url=" + document_url + "&webhook=" + webhook_url + "&postscript=" + postscript

	log.Println(carve_url)

	res, err := http.Get(carve_url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
}

func determineStatusCodeFromLogicError(logic_error *handshakejserrors.LogicError) int {
	code := 400
	if logic_error.Code == LOGIC_ERROR_CODE_UNKNOWN {
		code = 500
	}

	return code
}

func loadEnvs() {
	godotenv.Load()

	CARVE_ROOT = os.Getenv("CARVE_ROOT")
	SIGNATURE_CATCHER_ROOT = os.Getenv("SIGNATURE_CATCHER_ROOT")
	SIGNATURE_SHARE_ROOT = os.Getenv("SIGNATURE_SHARE_ROOT")
	SIGNATURE_API_ROOT = os.Getenv("SIGNATURE_API_ROOT")
	ORCHESTRATE_API_KEY = os.Getenv("ORCHESTRATE_API_KEY")
}
