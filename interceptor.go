package xhttp

import (
	"errors"
	"strconv"
)

func WhetherStatusCode(statusCode int) func(r Response) error {
	return func(r Response) error {
		res := r.Response()
		if res.StatusCode == statusCode {
			return nil
		}
		return errors.New("wrong status code: " + strconv.FormatInt(int64(res.StatusCode), 10))
	}
}

func WhetherStatusCodes(statusCodes ...int) func(r Response) error {
	return func(r Response) error {
		res := r.Response()
		for _, code := range statusCodes {
			if res.StatusCode == code {
				return nil
			}
		}
		return errors.New("wrong status code: " + strconv.FormatInt(int64(res.StatusCode), 10))
	}
}
