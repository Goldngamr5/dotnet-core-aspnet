package fakes

import "sync"
import "fmt"

type VersionParser struct {
	ParseVersionCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Version string
			Err     error
		}
		Stub func(string) (string, error)
	}
}

func (f *VersionParser) ParseVersion(param1 string) (string, error) {
	fmt.Println("Parsing...")
	f.ParseVersionCall.mutex.Lock()
	fmt.Println("lock")
	defer f.ParseVersionCall.mutex.Unlock()
	fmt.Println("unlock")
	f.ParseVersionCall.CallCount++
	fmt.Println("inc callcount")
	f.ParseVersionCall.Receives.Path = param1
	fmt.Println("recieves.path set")
	if f.ParseVersionCall.Stub != nil {
		fmt.Println("returning stub")
		return f.ParseVersionCall.Stub(param1)
	}
	fmt.Println("returning version")
	return f.ParseVersionCall.Returns.Version, f.ParseVersionCall.Returns.Err
}
