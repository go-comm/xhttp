package xhttp

import (
	"fmt"
	"io"
	"net/http"
)

func closeBody(body io.ReadCloser) error {
	if body != nil {
		return body.Close()
	}
	return nil
}

func WriteBytes(w http.ResponseWriter, status int, contentType string, b []byte) error {
	if len(contentType) > 0 {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	_, err := w.Write(b)
	return err
}

func WriteString(w http.ResponseWriter, status int, contentType string, s string) error {
	return WriteBytes(w, status, contentType, []byte(s))
}

func WriteHTML(w http.ResponseWriter, status int, b []byte) error {
	return WriteBytes(w, status, "text/html; charset=UTF-8", b)
}

func WriteError(w http.ResponseWriter, status int, err string) error {
	return WriteBytes(w, status, "text/plain; charset=utf-8", []byte(err))
}

func WriteErrorf(w http.ResponseWriter, status int, format string, a ...interface{}) error {
	return WriteError(w, status, fmt.Sprintf(format, a...))
}

func WriteBody(w http.ResponseWriter, status int, contentType string, en Encoder, v interface{}) error {
	if len(contentType) > 0 {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	return en.Encode(w, v)
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	return WriteBody(w, status, "application/json", JSON, v)
}

func WriteXML(w http.ResponseWriter, status int, v interface{}) error {
	return WriteBody(w, status, "application/xml", XML, v)
}

func ReadBody(r *http.Request, decoder Decoder, v interface{}) error {
	defer closeBody(r.Body)
	err := decoder.Decode(r.Body, v)
	if err == io.EOF {
		return nil
	}
	return err
}

func ReadJSON(r *http.Request, v interface{}) error {
	return ReadBody(r, JSON, v)
}

func ReadXML(r *http.Request, v interface{}) error {
	return ReadBody(r, XML, v)
}
