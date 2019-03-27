# Dragonfly ![](https://media.giphy.com/media/3o7qDOQLYrStOriGC4/giphy.gif)

![CircleCI](https://circleci.com/gh/NBCFB/Dragonfly/tree/develop.svg?style=svg&circle-token=b846cc3cd91a7f556d8db84c7210ee9fbb38944c)

Codes for accessing Redis DB

![](https://thumbs.dreamstime.com/t/angry-boss-warning-sign-red-evil-head-hazard-attention-symbol-danger-road-triangle-terrible-director-81153486.jpg) 警告：为方便业务代码前移至主项目代码，`case_*`文件暂时保留，他们将在下一次release中被删除。请勿试用他们。


## Install 
`go get github.com/NBCFB/Dragonfly`

## How-To-Use

### Configuration
Put the configuration for Redis connection into **config.json**, the resulted config.json should look like:
```
{
  "mode": "local_test",
  ...
  "local_test": {
    "DB": {
      ...
    },
    
    "redisDB": {
      "host":             "172.18.1.187",
      "pass":             "ilovedragonfly1515",
      "maxIdle":          5,
      "maxActive":        12000,
      "maxConnLifetime":  3,
      "idleTimeout":      200
    }
    ...
  }
}
```
You can add multiple configuration blocks for different **mode** such as 'prod', 'dev'.

### Creat a client (a connection)
```
...
cfg, err := LoadServerConfig()
...
caller = NewCaller(cfg)
...
```

Dont worry about closing the client. The conn pool will automatically manage the connection. But you can always manually close one:
```
caller.client.Close()
```


### Set
Add/update a record in redis DB.
```
v, err := caller.Set("key:1", "val:1", 0)
if err != nil {
    fmt.Printf("error:%v", err)
}
```

### Set In Batch
Add/update multiple records (using one connection) in the redis DB. In this case, you need create a slice of RedisObj first.
```
objs := make([]RedisObj, 3)
objs[0] = RedisObj{K: "key:1", V: "val:1"}
objs[1] = RedisObj{K: "key:2", V: "val:2"}
objs[2] = RedisObj{K: "key:3", V: "val:3"}
...
err := caller.SetInBatch(objs)
if err != nil {
    fmt.Printf("error:%v", err)
}
```

### Get
Get a record from the redis DB.
```
v, err := caller.Get("key:1")
if err != nil {
    fmt.Printf("error:%v", err)
}
```

### Search
Search record(s) using patten string. 

You can **search with without keywords**. In this case, records that match the search pattern return.
```
objs, err := caller.Search("key*", nil)
if err != nil {
    fmt.Printf("error:%v", err)
}
```

You can **also search with keywords**. In this case, records that match the search pattern and (the values) match one of the keywords return.
```
objs, err := caller.Search("key:*", []string{"val:2"})
if err != nil {
    fmt.Printf("error:%v", err)
}
```

The returned result is a slice of RedisObj which has the structure:
```
type RedisObj struct {
    K string
    V string
}
```

### Delete
You can delete records by keys.
```
err = caller.Del("key:1", "key:2")
if err != nil {
    fmt.Printf("error:%v", err)
}
```

## Write you test
We use Behavioral Driven Test Framework [Ginkgo](https://github.com/onsi/ginkgo) to write our test. You have to install Ginkgo and its preferred matcher libs.
```
go get github.com/onsi/ginkgo
go get github.com/onsi/gomega/...
```

Make sure you also install the ginkgo CLI so that you can generate your own test suite.
```
go get github.com/onsi/ginkgo/ginkgo
```

Every time we write a behavioral driven test, we need create a new client. Thus, we make it happen in `BeforeEach` and `AfterEach`:
```
var _ = Describe("Dragonfly", func() {
    var caller *RedisCallers

    BeforeEach(func() {
	caller = NewCaller(nil)
	Expect(caller.Client.FlushDB().Err()).NotTo(HaveOccurred())
    })

    AfterEach(func() {
	Expect(caller.Client.Close()).NotTo(HaveOccurred())
    })
    ...
}
```

Then you can write your test case like this:
```
var _ = Describe("Dragonfly", func() {
    var caller *RedisCallers

    BeforeEach(func() {
	caller = NewCaller(nil)
	Expect(caller.Client.FlushDB().Err()).NotTo(HaveOccurred())
    })

    AfterEach(func() {
	Expect(caller.Client.Close()).NotTo(HaveOccurred())
    })
    ...
    
    It("can set in batch", func() {
    	objs := make([]RedisObj, 3)
    	objs[0] = RedisObj{K: "key:1", V: "val:1"}
    	objs[1] = RedisObj{K: "key:2", V: "val:2"}
    	objs[2] = RedisObj{K: "key:3", V: "val:3"}

    	err := caller.SetInBatch(objs)
	Expect(err).NotTo(HaveOccurred())

	expected, _ := caller.Search("key*", nil)
	Expect(len(expected)).To(Equal(3))

    })
...
})
```
