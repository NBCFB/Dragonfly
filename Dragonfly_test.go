package Dragonfly_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	. "github.com/NBCFB/Dragonfly"
)

var _ = Describe("Dragonfly", func() {
	var caller *RedisCallers

	BeforeEach(func() {
		// Set server config
		serverConfig := viper.New()
		serverConfig.SetConfigName("test_config")
		serverConfig.SetConfigType("json")

		// Read config file
		err := serverConfig.ReadInConfig()
		Expect(err).NotTo(HaveOccurred())
		Expect(serverConfig.Get("test.redisDB.host")).To(Equal("127.0.0.1"))

		caller = NewCaller(serverConfig)
		Expect(caller.Client.FlushDB().Err()).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(caller.Client.Close()).NotTo(HaveOccurred())
	})

	It("can set", func() {
		v, err := caller.Set("key:1", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal("val:1"))
	})

	It("fails to set [due-to: empty args]", func() {
		_, err := caller.Set("", "val:1", 0)
		Expect(err).Should(HaveOccurred())
	})

	It("can set in batch", func() {
		objs := make([]RedisObj, 3)
		objs[0] = RedisObj{K: "key:1", V: "val:1"}
		objs[1] = RedisObj{K: "key:2", V: "val:2"}
		objs[2] = RedisObj{K: "key:3", V: "val:3"}

		err := caller.SetInBatch(objs)
		Expect(err).NotTo(HaveOccurred())

		expected, _ := caller.SearchByKeys("key*", nil)
		Expect(len(expected)).To(Equal(3))

	})

	It("fails to set in batch", func() {
		err := caller.SetInBatch([]RedisObj{})
		Expect(err).Should(HaveOccurred())
	})

	It("gets", func() {
		_, err := caller.Set("key:1", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		v, err := caller.Get("key:1")
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal("val:1"))
	})

	It("fails to set [due-to: empty key]", func() {
		_, err := caller.Get("")
		Expect(err).Should(HaveOccurred())
	})

	It("fails to get [due-to: wrong keys]", func() {
		_, err := caller.Set("key:1", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = caller.Get("ghost_key:1")
		Expect(err).Should(HaveOccurred())
	})

	It("can search", func() {
		_, err := caller.Set("key:1", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = caller.Set("key:2", "val:2", 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = caller.Set("wired_key:2", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		objs, err := caller.SearchByKeys("key*", nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(objs)).To(Equal(2))
		Expect(objs).To(Equal([]RedisObj{{K: "key:1", V: "val:1"}, {K: "key:2", V: "val:2"}}))

		objs, err = caller.SearchByKeys("ghost_key*", nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(objs)).To(Equal(0))
	})

	It("can search with keywords", func() {
		_, err := caller.Set("key:1", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = caller.Set("key:2", "val:2", 0)
		Expect(err).NotTo(HaveOccurred())

		objs, err := caller.SearchByKeys("key:*", []string{"val:2"})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(objs)).To(Equal(1))
		Expect(objs).To(Equal([]RedisObj{{K: "key:2", V: "val:2"}}))
	})

	It("fails to search [due-to: empty patten]", func() {
		_, err := caller.Set("key:1", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = caller.Set("key:2", "val:2", 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = caller.SearchByKeys("", []string{"val:2"})
		Expect(err).Should(HaveOccurred())
	})

	It("can delete", func() {
		_, err := caller.Set("key:1", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		_, err = caller.Set("key:2", "val:1", 0)
		Expect(err).NotTo(HaveOccurred())

		err = caller.Del("key:1")
		Expect(err).NotTo(HaveOccurred())

		objs, _ := caller.SearchByKeys("key*", nil)
		Expect(len(objs)).To(Equal(1))

		err = caller.Del("ghost_key:1")
		Expect(err).NotTo(HaveOccurred())
	})

	It("fails to delete [due-to: empty key(s)]", func() {
		err := caller.Del("", "")
		Expect(err).Should(HaveOccurred())
	})
})
