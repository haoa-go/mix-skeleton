package di

import jsoniter "github.com/json-iterator/go"

func Json() jsoniter.API {
	return jsoniter.ConfigCompatibleWithStandardLibrary
}
