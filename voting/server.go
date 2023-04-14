package voting

import (
	context "context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jamesruan/sodium"
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
	var co int32
	co = 1
	for _, ele := range RElection {
		if election_name == ele.Name {
			now := time.Now()
			if ele.End_time.Before(now) {
				co = 0
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
				co = 2
			}
		}
	}

	if co != 0 {
		fmt.Println("GetResult error")
	}
	return &ElectionResult{Status: &co, Counts: counts}, nil

}

func (s *Server) CastVote(ctx context.Context, in *Vote) (*Status, error) {
	var election_name string
	election_name = *in.ElectionName
	choice := *in.ChoiceName
	findChoice := false
	var co int32
	co = -1

	now := time.Now()
	var voter_group string
	for _, vo := range RVoter {
		if Equal(in.Token.Value, vo.V_token) {
			// check token
			if vo.Expired_time.Before(now) {
				fmt.Println("token expired")
				co = 1
				return &Status{Code: &co}, nil
			}
			co = 0
			voter_group = vo.Group
		}
	}

	if co == -1 {
		fmt.Println("token error")
		return &Status{Code: &co}, nil
	}

	for _, ele := range RElection {
		if election_name == ele.Name {
			findGroup := false
			for _, grp := range ele.Groups {
				if grp == voter_group {
					findGroup = true
				}
			}
			if !findGroup {
				co = 3
				return &Status{Code: &co}, nil
			}

			for _, ch := range ele.Choices {
				if choice == ch {
					fmt.Println("find a match")
					findChoice = true
					ele.Result[choice]++
				}
			}
		}
	}

	if !findChoice {
		fmt.Println("cant find choice")
		co = 2
		return &Status{Code: &co}, nil
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
		co = 3
		return &Status{Code: &co}, nil
	}

	for _, vo := range RVoter {
		if vo.Name == in.GetName() {
			if !Equal(in.Token.Value, vo.V_token) {
				fmt.Println("token error")
				co = 1
				return &Status{Code: &co}, nil
			}
			if vo.Expired_time.Before(now) {
				fmt.Println("token expired")
				co = 1
				return &Status{Code: &co}, nil
			}
		}
	}

	if len(in.Groups) == 0 || len(in.Choices) == 0 {
		fmt.Println("Missing groups or choices")
		co = 2
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

func (s *Server) PreAuth(ctx context.Context, in *VoterName) (*Challenge, error) {
	return &Challenge{Value: []byte("test challenge")}, nil
}

func (s *Server) Auth(ctx context.Context, in *AuthRequest) (*AuthToken, error) {
	fmt.Println(in)
	for i, vo := range RVoter {
		if vo.Name == in.GetName().GetName() {
			sig := in.GetResponse().GetValue()

			if len(sig) != 64 {
				fmt.Println("sig size error")
				return nil, nil
			}

			err := sodium.Bytes("test challenge").SignVerifyDetached(
				sodium.Signature{sodium.Bytes(sig)},
				sodium.SignPublicKey{sodium.Bytes(vo.Public_key)},
			)
			if err == nil {
				return &AuthToken{Value: []byte("")}, err
			}
			id := uuid.NewString()
			RVoter[i].V_token = []byte(id)
			RVoter[i].Expired_time = time.Now().Add(time.Hour)
			fmt.Println(RVoter[i])
			return &AuthToken{Value: []byte(id)}, err
		}
	}

	return &AuthToken{Value: nil}, nil
}

type RegisteredVoter struct {
	Name          string
	Group         string
	Public_key    []byte
	Auth_response []byte
	V_token       []byte
	Expired_time  time.Time
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
