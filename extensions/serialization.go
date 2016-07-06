// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package extensions

import "gopkg.in/vmihailenco/msgpack.v2"

//Serialize dehidrates an item to the session storage
func Serialize(payload interface{}) (string, error) {
	item, err := msgpack.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(item), nil
}

//Deserialize rehidrates an item in the session storage
func Deserialize(payload string) (interface{}, error) {
	var item interface{}
	err := msgpack.Unmarshal([]byte(payload), &item)
	if err != nil {
		return nil, err
	}

	return item, nil
}
