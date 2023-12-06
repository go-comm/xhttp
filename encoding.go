package xhttp

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
)

var (
	JSON = withEncoding(
		EncoderFunc(func(w io.Writer, v interface{}) error {
			return json.NewEncoder(w).Encode(v)
		}),
		DecoderFunc(func(r io.Reader, v interface{}) error {
			return json.NewDecoder(r).Decode(v)
		}),
	)

	Gob = withEncoding(
		EncoderFunc(func(w io.Writer, v interface{}) error {
			return gob.NewEncoder(w).Encode(v)
		}),
		DecoderFunc(func(r io.Reader, v interface{}) error {
			return gob.NewDecoder(r).Decode(v)
		}),
	)

	XML = withEncoding(
		EncoderFunc(func(w io.Writer, v interface{}) error {
			return xml.NewEncoder(w).Encode(v)
		}),
		DecoderFunc(func(r io.Reader, v interface{}) error {
			return xml.NewDecoder(r).Decode(v)
		}),
	)
)

type encoding struct {
	Encoder
	Decoder
}

func withEncoding(en Encoder, de Decoder) *encoding {
	return &encoding{Encoder: en, Decoder: de}
}

type Decoder interface {
	Decode(r io.Reader, v interface{}) error
}

type DecoderFunc func(r io.Reader, v interface{}) error

func (f DecoderFunc) Decode(r io.Reader, v interface{}) error {
	return f(r, v)
}

type Encoder interface {
	Encode(w io.Writer, v interface{}) error
}

type EncoderFunc func(w io.Writer, v interface{}) error

func (f EncoderFunc) Encode(w io.Writer, v interface{}) error {
	return f(w, v)
}
