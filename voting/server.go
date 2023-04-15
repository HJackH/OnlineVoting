package voting

import (
	context "context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jamesruan/sodium"
)

var RVoter []RegisteredVoter
var WVoter []RegisteredVoter
var RElection []election
var tokenByte = []byte("GoodLuck")
var challenge_byte = []byte("NiceDay")

// We define a server struct that implements the server interface.
type Server struct {
	UnimplementedVotingServer
}

func GetIntPointerS(value string) *string {
	return &value
}

func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// We implement the SayHello method of the server interface.
func (s *Server) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	return &HelloReply{Message: GetIntPointerS("Hello, " + in.GetName())}, nil
}

func (s *Server) PreAuth(ctx context.Context, in *VoterName) (*Challenge, error) {
	fmt.Println()
	name := *in.Name
	find := false
	for i, vo := range RVoter {
		if name == vo.Name {
			find = true
			RVoter[i].Challenge = challenge_byte
			break
		}
	}
	if find == false {
		fmt.Println("Can't find voter name")
		return nil, nil
	}
	ch := Challenge{
		Value: challenge_byte,
	}
	return &ch, nil
}

func (s *Server) Auth(ctx context.Context, in *AuthRequest) (*AuthToken, error) {
	fmt.Println()
	name := *in.Name.Name
	sig := sodium.Signature{
		Bytes: in.Response.Value,
	}
	// fmt.Print("len:")
	// fmt.Println(len(in.Response.Value))
	if len(in.Response.Value) != 64 {
		fmt.Println("sig size error")
		return nil, nil
	}
	find := false
	var index int
	for i, vo := range RVoter {
		if name == vo.Name {
			find = true
			RVoter[i].Challenge = challenge_byte
			index = i
			break
		}
	}
	if find == false {
		fmt.Println("Can't find voter name")
		return nil, nil
	}
	ch := sodium.Bytes(challenge_byte)
	err := ch.SignVerifyDetached(sig, RVoter[index].Public_key)
	if err != nil {
		fmt.Println("SignVerifyDetached error")
		fmt.Println(err)
	}

	fmt.Println("Create random string...")
	to := RandomString(25)
	fmt.Println(to)
	token := []byte(to)
	au := AuthToken{
		Value: token,
	}
	if err == nil {
		fmt.Println("The challenge is properly signed")
		RVoter[index].V_token = token
		now := time.Now()
		end_time := now.Add(time.Minute * 5)
		fmt.Print("token valid time: ")
		fmt.Println(end_time)
		RVoter[index].token_End_time = end_time
	}
	return &au, err
}

func (s *Server) GetResult(ctx context.Context, in *ElectionName) (*ElectionResult, error) {
	election_name := *in.Name
	var counts []*VoteCount
	success := false
	var co int32
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
				co = 2
				counts = nil
				return &ElectionResult{Status: &co, Counts: counts}, nil
			}
		}
	}

	co = 0
	if !success {
		fmt.Println("Non-existent election")
		co = 1
		return &ElectionResult{Status: &co, Counts: counts}, nil
	}
	return &ElectionResult{Status: &co, Counts: counts}, nil

}

func (s *Server) CastVote(ctx context.Context, in *Vote) (*Status, error) {
	var co int32
	var group string
	var name string
	var index int
	token := in.Token.Value
	find := false
	for _, vo := range RVoter {
		if Equal(token, vo.V_token) == true && token != nil {
			now := time.Now()
			if vo.token_End_time.Before(now) == true {
				fmt.Println("token expire")
				break
			} //token expire
			find = true //token has not expired
			group = vo.Group
			name = vo.Name
			break
		}
	}
	if find == false { //token error
		fmt.Println("token error")
		co = 1
		fmt.Println(co)
		return &Status{Code: &co}, nil
	}
	var election_name string
	election_name = *in.ElectionName
	findEle := false
	var right_ele election
	for i, ele := range RElection {
		if election_name == ele.Name {
			findEle = true
			right_ele = ele
			index = i
			break
		}
	}
	if findEle == false {
		fmt.Println("can't find election")
		co = 2
		return &Status{Code: &co}, nil
	}

	now := time.Now()
	if right_ele.End_time.Before(now) == true {
		fmt.Println("Election deadline has passed")
		RElection = append(RElection[:index], RElection[index+1:]...)
		fmt.Println("RElection:")
		fmt.Println(RElection)
		co = 5
		return &Status{Code: &co}, nil
	}

	findGroup := false
	for _, g := range right_ele.Groups {
		if g == group {
			findGroup = true
			break
		}
	}
	if findGroup == false {
		fmt.Println("group not find")
		co = 3
		return &Status{Code: &co}, nil
	}

	for _, na := range right_ele.Already_voted {
		if name == na {
			co = 4
			return &Status{Code: &co}, nil
		}
	}

	choice := *in.ChoiceName

	for _, ch := range right_ele.Choices {
		if choice == ch {
			RElection[index].Result[choice]++
			RElection[index].Already_voted = append(RElection[index].Already_voted, name)
			break
		}
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
	fmt.Printf("Election end time:")
	fmt.Println(t1)
	var co int32
	now := time.Now()
	if t1.Before(now) {
		fmt.Println("time is expired!")
		co = -1
		return &Status{Code: &co}, nil
	}

	token := in.Token.Value
	find := false
	for _, vo := range RVoter {
		if Equal(token, vo.V_token) == true && token != nil {
			now := time.Now()
			if vo.token_End_time.Before(now) == true {
				fmt.Println("token expire")
				break
			} //token expire
			find = true //token has not expired
		}
	}
	if find == false { //token error
		fmt.Println("token error")
		co = 1
		fmt.Println(co)
		return &Status{Code: &co}, nil
	}

	if in.Groups == nil || in.Choices == nil {
		fmt.Println("something missing")
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
	fmt.Println()
	fmt.Println("Voter from client")
	fmt.Printf("Voter Name: %s\n", *in.Name)
	fmt.Printf("Voter Group: %s\n", *in.Group)
	fmt.Printf("Voter Public key: ")
	fmt.Println(in.PublicKey)
	pub := sodium.SignPublicKey{
		Bytes: in.PublicKey,
	}
	voter := RegisteredVoter{
		Name:       *in.Name,
		Group:      *in.Group,
		Public_key: pub,
		Alive:      true,
	}
	fmt.Println(voter)
	WVoter = append(WVoter, voter)
	var x int32
	x = 0
	return &Status{Code: &x}, nil
}

type RegisteredVoter struct {
	Name           string
	Group          string
	Public_key     sodium.SignPublicKey
	Auth_response  []byte
	V_token        []byte
	Challenge      []byte
	Alive          bool
	token_End_time time.Time
}

type election struct {
	Name          string
	Groups        []string
	Choices       []string
	Already_voted []string
	End_time      time.Time
	Alive         bool
	Result        map[string]int32
}
