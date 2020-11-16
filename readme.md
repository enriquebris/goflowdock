### Gett all flows
```go
package main

import (
	"fmt"
	fd "github.com/enriquebris/goflowdock"
)

const (
	flowdockToken        = "your flowdock token here"
)

func main() {
	manager := fd.NewFlowManager(flowdockToken)
	flows, err := manager.GetFlows()
	if err != nil {
		fmt.Println(err)
		return
	}

	for i:=0; i<len(flows); i++ {
		fmt.Printf("%v :: %v\n", flows[i].Name, flows[i].ID)
	}
}
```

### Post a message
```go
package main

import (
	"fmt"
	fd "github.com/enriquebris/goflowdock"
)

const (
	flowdockToken        = "your flowdock token here"
	flowdockOrganization = "your flowdock organization here"
)

func main() {
	mmanager := fd.NewMessageManager(flowdockToken, flowdockOrganization)
	fmt.Println(mmanager.SendMessage(fd.MessageData{
		Flow: "flow-id",
		Event: "message",
		Content: "write here the text you want to post",
		ExternalUserName: "username",
		Tags: []string{"#tag1", "#tag2"},
	}))
}
```

### Post a message as an embedded answer
```go
package main

import (
	"fmt"
	fd "github.com/enriquebris/goflowdock"
)

const (
	flowdockToken        = "your flowdock token here"
	flowdockOrganization = "your flowdock organization here"
)

func main() {
	mmanager := fd.NewMessageManager(flowdockToken, flowdockOrganization)
	fmt.Println(mmanager.SendMessage(fd.MessageData{
		Flow: "flow-id",
		Event: "message",
		Content: "write here the text you want to post",
		ExternalUserName: "username",
		Tags: []string{"#tag1", "#tag2"},
		// threadID from the original message
		ThreadID: "threadID",
	}))
}
```
