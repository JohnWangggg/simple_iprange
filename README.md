# simple_iprange

A simple Go library for parsing and handling IPv4&IPv6 ranges. 

## Usage 

```go
package main
import (
    "fmt"
    "github.com/JohnWangggg/simple_iprange"
)

func main() {
	var inplist string = "10.0.0.1,10.0.0.5-10.0.0.10,192.168.1.*,192.168.10.0/24,192.168.1.1-192.168.1.20"
	list, err := simple_iprange.ParseList(inplist)
	//list, err := simple_iprange.Parse("192.168.1.1/24")
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	fmt.Printf("list: %s\n", list)

	rng, err := list.Expand()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	fmt.Printf("rng: %s\n", rng)

}
```


## iprange supports the following formats:

- 192.168.1.1
- 192.168.1.0/24
- 192.168.1.1-192.168.1.10
- 192.168.1.1-10
- 192.168.1.*
- 192.168.10.0/24
- 2001:0db8:85a3:08d3:1319:8a2e:0370:7344
- 2001:0db8:85a3:08d3:1319:8a2e:0370:7344,2002:0db8:0::0:1428:57ab
- 2001:db8::/48