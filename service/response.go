package service

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type Response struct {
	Slots    []Slots       `json:"slots"`
	Messages []ResponseMsg `json:"msg_body"`
	ErrorNo  string        `json:"error_no"`
}

type Slots struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResponseMsg struct {
	Data map[string]utf8String `json:"data"`
	Type string                `json:"type"`
}

func NewResponse(code string) *Response {
	return &Response{
		ErrorNo: code,
	}
}

// ToBytes convert to bytes
func (x *Response) ToBytes() []byte {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(x)
	return buffer.Bytes()
}

type utf8String string

// custom marshal method
func (s utf8String) MarshalJSON() ([]byte, error) {
	return []byte(strconv.QuoteToASCII(string(s))), nil
}
