package model

type Response struct {
	Error    interface{} `json:"error,omitempty"`
	Response interface{} `json:"response,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}
