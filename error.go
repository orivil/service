// Copyright 2020 orivil.com. All rights reserved.

package service

import "strings"

type Errors []error

func (errs Errors) IsError(err error) bool {
	for _, e := range errs {
		if e == err {
			return true
		}
	}
	return false
}

func (errs Errors) Error() string {
	var errors []string
	for _, e := range errs {
		errors = append(errors, e.Error())
	}
	return strings.Join(errors, "; ")
}
