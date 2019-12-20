package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// リクエストの型
type Request struct {
	Key      string
	ID       string
	Password string
}

// 結果
type Response struct {
	Status bool `xml:"status" json:"status"`
}

// PostするXML
type Post struct {
	XMLName     xml.Name `Request`
	Credentials struct {
		ID       string `xml:"id"`
		Password string `xml:"password"`
	} `xml:"Credentials"`
	Identity struct {
		Key string `xml:"key"`
	}
}

// テンプレート
var getTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<Request>
    <Credentials>
        <id>{{.ID}}</id>
        <password>{{.Password}}</password>
    </Credentials>
	<Identity>
        <key>{{.Key}}</key>
   </Identity>
</Request>`

// SOAP アクセス
func handler(w http.ResponseWriter, r *http.Request) {
	req := populateRequest()

	httpReq, err := generateSOAPRequest(req)
	if err != nil {
		log.Println("Some problem occurred in request generation")
	}
	log.Println(httpReq)
	response, err := soapCall(httpReq)
	if err != nil {
		log.Println("Problem occurred in making a SOAP call")
	}
	res, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

// リクエストの内容
func populateRequest() *Request {
	req := Request{}
	req.Key = "12345678"
	req.ID = "SENOUE"
	req.Password = "Password"
	return &req
}

// SOAP リクエストの作成
func generateSOAPRequest(req *Request) (*http.Request, error) {
	// テンプレートを使ってXMLを作成
	template, err := template.New("InputRequest").Parse(getTemplate)

	if err != nil {
		log.Printf("Error while marshling object. %s ", err.Error())
		return nil, err
	}

	doc := &bytes.Buffer{}
	err = template.Execute(doc, req)
	if err != nil {
		log.Printf("template.Execute error. %s ", err.Error())
		return nil, err
	}

	r, err := http.NewRequest(http.MethodPost, "http://localhost:8060/resp", doc)
	r.Header.Add("Content-Type", "application/xml; charset=utf-8")
	if err != nil {
		log.Printf("Error making a request. %s ", err.Error())
		return nil, err
	}

	return r, nil
}

// SOAP コール
func soapCall(req *http.Request) (*Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &Response{}
	err = xml.Unmarshal(body, &r)

	if err != nil {
		return nil, err
	}
	return r, nil
}

// テストのレスポンス
func respHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Status bool `xml:"status"`
	}

	xmldata := Response{
		Status: false,
	}

	body, err := ioutil.ReadAll(r.Body)
	post := &Post{}
	err = xml.Unmarshal(body, &post)

	// 適当なチェック
	if post.Credentials.ID != "SENOUE" {
		buf, err := xml.MarshalIndent(&xmldata, "", "    ")
		log.Println("Post Parse error:", err)
		w.Header().Set("Content-Type", "application/xml")
		w.Write(buf)
		return
	}

	xmldata.Status = true
	buf, err := xml.MarshalIndent(&xmldata, "", "    ")
	if err != nil {
		log.Println("XML Marshal error:", err)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Write(buf)
}

// MAIN
func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/resp", respHandler)
	http.ListenAndServe(":8060", nil)
}
