# Dragonfly ![](https://media.giphy.com/media/3o7qDOQLYrStOriGC4/giphy.gif)

![CircleCI](https://circleci.com/gh/NBCFB/Dragonfly/tree/develop.svg?style=svg&circle-token=b846cc3cd91a7f556d8db84c7210ee9fbb38944c)

Case Status Service Package


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
You can add multiple configuration block for different **mode** such as 'prod', 'dev'.

### Initialise Pool
Initilise pool creates a new pool for Redis connections. The pool is a variable defined in **redis_pooler.go**
```
var (
	Pool *redis.Pool
)
```
When server starts, or whenever you are about to start interaction with Redis, call init from **redis_pooler.go**
```
...
init()
...
```
Then, you can get connection from the created Redis pool
```
c := Pool.Get()
defer c.Close()
```  

### Model - CaseStatus
Here is the only model we use it the package
```
type CaseStatus struct {
	UserId		string	`json:"userId"`
	CorpId		string	`json:"corpId"`
	CaseId		string	`json:"caseId"`
	Status		int	`json:"status"`
}
```

### Change Status
Here is an example of updating the status of the case(id:1) under corp(id:1) and user(id:1), the new status is 'read'(status:1):
```
userId := "1"
corpId := "1"
caseId := "1"
status := 1  // 1 - read, 0 - unread
err := SetStatus(userId, corpId, caseId string, status int)
if err != nil {
  // Set status failed, do something
} else {
  // Set status is successful, do something else
}
```

### Change multiple case status
You need pass an array of CaseStatus, so that we can update them all at once.
```
err := BatchSetStatus(css []CaseStatus)
```

### Delete a case status record
Here is an example of deletion the case(id:1) under corp(id:1) and user(id:1):
```
userId := "1"
corpId := "1"
caseId := "1" 
err := DeleteStatus(userId, corpId, caseId string)
```

### Get status use pattern
We provide a function that can search keys in redis using pattern. User does not have to specify the search pattern. Instead, we will setup the pattern based on the values of function parameters. **Make sure userId is not empty, otherwise, it raises an error**. The search pattern is setup according to the parameter values:
- No userId --> Error
- Got userId but no corpId --> Search 'all the corps under a same user'
- Got userId, corpId but no caseId --> Search 'all the cases under a same user and a specified corp'
- Got userId, corpId, caseId --> Search 'a status record of given user, corp and case'

Here is an example of obtaining the status records of corp(id:1) under user(id:1):
```
userId := "1"
corpId := "1"
caseId := "" // no case Id is specified
css, err := GetStatusByMatch(userId, corpId, caseId string)
if err != nil {
    // Do something to handle error
} else {
    for _, cs := range(css) {
        // Print out the cs
        fmt.Sprintf("UserId:%s, CorpId:%s, CaseId:%s, Status:%d", cs.UserId, cs.CorpId, cs.CaseId, cs.Status)
    }
    ...
}
```

### Get status of a single case
Here is an example of obtaining the status of the case(id:1) under corp(id:1) and user(id:1):
```
userId := "1"
corpId := "1"
caseId := "1"
status, err := GetStatusByKey(userId, corpId, caseId string)
if err != nil {
    // Do something to handle the error
} else {
    // Print status
    fmt.Println("Status:", status)
}
```
