// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

type Product interface {
	Start() error
	Stop() error
}

var products = make(map[string]func(*Server) (Product, error))

func RegisterProduct(name string, f func(*Server) (Product, error)) {
	products[name] = f
}
