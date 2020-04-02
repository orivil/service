// Copyright 2020 orivil.com. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found at https://mit-license.org.

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
