// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package extensions_test

import (
	"github.com/heynemann/level/extensions"
	"gopkg.in/vmihailenco/msgpack.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Serialization and Deserialization", func() {

	Describe("Serialization", func() {
		It("should serialize using msgpack", func() {
			expected := map[string]interface{}{"a": 1}
			result, err := extensions.Serialize(expected)

			Expect(err).NotTo(HaveOccurred())

			serialized, err := msgpack.Marshal(expected)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEquivalentTo(serialized))
		})

		It("should fail to serialize invalid object", func() {
			_, err := extensions.Serialize(func() {})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("msgpack: Encode(unsupported func())"))
		})
	})

	Describe("Deserialization", func() {
		It("should deserialize using msgpack", func() {
			expected := map[interface{}]interface{}{"a": 1}
			serialized, err := msgpack.Marshal(expected)
			Expect(err).NotTo(HaveOccurred())

			result, err := extensions.Deserialize(string(serialized))
			actual := result.(map[interface{}]interface{})

			Expect(err).NotTo(HaveOccurred())
			Expect(actual["a"]).To(BeEquivalentTo(1))
		})

		It("should fail to deserialize invalid payload", func() {
			_, err := extensions.Deserialize("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("EOF"))
		})
	})
})
