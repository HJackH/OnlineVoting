package voting

import (
	context "context"
	"fmt"
	"time"
)

var RVoter []RegisteredVoter
var RElection []election
var tokenByte = []byte("GoodLuck")

// We define a server struct that implements the server interface.
type Server struct {
	UnimplementedVotingServer
}

func GetIntPointerS(value string) *string {
	return &value
}

// We implement the SayHello method of the server interface.
func (s *Server) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	return &HelloReply{Message: GetIntPointerS("Hello, " + in.GetName())}, nil
}

func (s *Server) GetResult(ctx context.Context, in *ElectionName) (*ElectionResult, error) {
	election_name := *in.Name
	var counts []*VoteCount
	success := false
	for _, ele := range RElection {
		if election_name == ele.Name {
			now := time.Now()
			if ele.End_time.Before(now) {
				success = true
				for _, ch := range ele.Choices {
					fmt.Print("choice: ")
					fmt.Println(ch)
					t := ele.Result[ch]
					c_name := ch
					temp := VoteCount{ChoiceName: &c_name, Count: &t}
					counts = append(counts, &temp)
				}
				fmt.Println(counts)
			} else {
				fmt.Println("time is not up yet")
			}
		}
	}
	var co int32
	co = 0
	if !success {
		fmt.Println("GetResult error")
		co = -1
	}
	return &ElectionResult{Status: &co, Counts: counts}, nil

}

func (s *Server) CastVote(ctx context.Context, in *Vote) (*Status, error) {
	var election_name string
	election_name = *in.ElectionName
	choice := *in.ChoiceName
	findChoice := false
	var co int32
	if Equal(in.Token.Value, tokenByte) == false {
		fmt.Println("token error")
		co = -1
		return &Status{Code: &co}, nil
	}
	for _, ele := range RElection {
		if election_name == ele.Name {
			for _, ch := range ele.Choices {
				if choice == ch {
					fmt.Println("find a match")
					findChoice = true
					ele.Result[choice]++
				}
			}
		}
	}

	if findChoice == false {
		fmt.Println("cant find choice")
		co = -1
	} else {
		co = 0
	}

	fmt.Println(RElection)
	return &Status{Code: &co}, nil

}

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (s *Server) CreateElection(ctx context.Context, in *Election) (*Status, error) {
	endtime := in.EndDate
	t1 := time.Unix(endtime.GetSeconds(), 0)
	fmt.Println(t1)
	var co int32
	now := time.Now()
	if t1.Before(now) {
		fmt.Println("time is expired!")
		co = -1
		return &Status{Code: &co}, nil
	}

	if Equal(in.Token.Value, tokenByte) == false {
		fmt.Println("token error")
		co = -1
		return &Status{Code: &co}, nil
	}

	result := map[string]int32{}
	for _, ch := range in.Choices {
		result[ch] = int32(0)
	}

	e := election{
		Name:     *in.Name,
		Groups:   in.Groups,
		Choices:  in.Choices,
		End_time: t1,
		Alive:    true,
		Result:   result,
	}
	fmt.Println(e)
	RElection = append(RElection, e)

	co = 0
	return &Status{Code: &co}, nil
}

func GetIntPointer(value int32) *int32 {
	return &value
}

func (s *Server) RegisterVoter(ctx context.Context, in *Voter) (*Status, error) {
	fmt.Println("RegisterVoter")
	var x int32
	x = 1

	return &Status{Code: &x}, nil
}

type RegisteredVoter struct {
	Name          string
	Group         string
	Public_key    []byte
	Auth_response []byte
	V_token       []byte
	Alive         bool
}

type election struct {
	Name     string
	Groups   []string
	Choices  []string
	End_time time.Time
	Alive    bool
	Result   map[string]int32
}
