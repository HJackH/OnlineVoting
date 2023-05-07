package voting

import (
	context "context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jamesruan/sodium"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var tokenByte = []byte("GoodLuck")
var challenge_byte = []byte("NiceDay")
var RVoter_coll *mongo.Collection
var WVoter_coll *mongo.Collection
var RElection_coll *mongo.Collection

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var vo RegisteredVoter
	filter := bson.D{{"name", name}}
	update := bson.D{{"$set", bson.D{{"challenge", challenge_byte}}}}
	err := RVoter_coll.FindOneAndUpdate(ctx, filter, update).Decode(&vo)

	fmt.Println("err:", err)
	if err != nil {
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

	if len(in.Response.Value) != 64 {
		fmt.Println("sig size error")
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var vo RegisteredVoter
	err := RVoter_coll.FindOne(ctx, bson.D{{"name", name}}).Decode(&vo)

	if err != nil {
		fmt.Println("Can't find voter name")
		return nil, nil
	}

	ch := sodium.Bytes(challenge_byte)
	err = ch.SignVerifyDetached(sig, vo.Public_key)
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
		now := time.Now()
		end_time := now.Add(time.Hour * 1)
		fmt.Print("token valid time: ")
		fmt.Println(end_time)

		filter := bson.D{{"name", name}}
		update := bson.D{{"$set", bson.D{{"v_token", token}, {"token_end_time", end_time.Format("2006-01-02 15:04:05")}}}}
		err = RVoter_coll.FindOneAndUpdate(ctx, filter, update).Decode(&vo)

		if err != nil {
			fmt.Println("update RVoter failed")
			fmt.Println(err)
		}
	}
	return &au, err
}

func (s *Server) GetResult(ctx context.Context, in *ElectionName) (*ElectionResult, error) {
	election_name := *in.Name
	var counts []*VoteCount
	var co int32

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ele election
	err := RElection_coll.FindOne(ctx, bson.D{{"name", election_name}}).Decode(&ele)

	if err != nil {
		fmt.Println("Non-existent election")
		co = 1
		return &ElectionResult{Status: &co, Counts: counts}, nil
	}

	now := time.Now()
	ed_time, err := time.Parse("2006-01-02 15:04:05", ele.End_time)
	if ed_time.After(now) {
		fmt.Println("time is not up yet")
		co = 2
		counts = nil
		return &ElectionResult{Status: &co, Counts: counts}, nil
	}

	for _, ch := range ele.Choices {
		fmt.Print("choice: ")
		fmt.Println(ch)
		t := ele.Result[ch]
		c_name := ch
		temp := VoteCount{ChoiceName: &c_name, Count: &t}
		counts = append(counts, &temp)
	}
	fmt.Println(counts)

	co = 0
	return &ElectionResult{Status: &co, Counts: counts}, nil
}

func (s *Server) CastVote(ctx context.Context, in *Vote) (*Status, error) {
	var co int32
	var group string
	var name string
	// var index int
	token := in.Token.Value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var vo RegisteredVoter
	err := RVoter_coll.FindOne(ctx, bson.D{{"v_token", token}}).Decode(&vo)

	if err != nil {
		fmt.Println("token not found")
		co = 1
		return &Status{Code: &co}, nil
	}

	now := time.Now()
	tk_ed_time, err := time.Parse("2006-01-02 15:04:05", vo.Token_End_time)
	if tk_ed_time.Before(now) == true {
		fmt.Println("token expire")
		co = 1
		return &Status{Code: &co}, nil
	}

	group = vo.Group
	name = vo.Name

	var election_name string
	election_name = *in.ElectionName
	var right_ele election
	err = RElection_coll.FindOne(ctx, bson.D{{"name", election_name}}).Decode(&right_ele)

	if err != nil {
		fmt.Println("Non-existent election")
		co = 2
		return &Status{Code: &co}, nil
	}

	now = time.Now()
	ele_ed_time, err := time.Parse("2006-01-02 15:04:05", right_ele.End_time)
	if ele_ed_time.Before(now) == true {
		fmt.Println("Election deadline has passed")
		res, err2 := RElection_coll.DeleteOne(ctx, bson.D{{"name", election_name}})
		fmt.Println(res)
		if err2 != nil {
			fmt.Println("delete failed")
		}
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

	fmt.Println("err:", err)
	if err != nil {
		fmt.Println("Can't find voter name")
		return nil, nil
	}

	for _, ch := range right_ele.Choices {
		if choice == ch {
			right_ele.Result[choice]++
			right_ele.Already_voted = append(right_ele.Already_voted, name)
			break
		}
	}

	var ele election
	filter := bson.D{{"name", election_name}}
	update := bson.D{{"$set", bson.D{{"result", right_ele.Result}, {"already_voted", right_ele.Already_voted}}}}
	err = RElection_coll.FindOneAndUpdate(ctx, filter, update).Decode(&ele)

	if err != nil {
		fmt.Println("Non-existent election")
		co = 2
		return &Status{Code: &co}, nil
	}
	co = 0
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var vo RegisteredVoter
	err := RVoter_coll.FindOne(ctx, bson.D{{"v_token", token}}).Decode(&vo)

	if err != nil {
		fmt.Println("token not found")
		co = 1
		return &Status{Code: &co}, nil
	}

	fmt.Println(vo)

	fmt.Println(now)
	fmt.Println(vo.Token_End_time)
	tk_ed_time, err := time.Parse("2006-01-02 15:04:05", vo.Token_End_time)
	if tk_ed_time.Before(now) == true {
		fmt.Println("token expire")
		co = 1
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
		End_time: t1.Format("2006-01-02 15:04:05"),
		Alive:    true,
		Result:   result,
	}

	res, err := RElection_coll.InsertOne(ctx, e)

	fmt.Println(res)
	if err != nil {
		fmt.Println("insert to RElection failed")
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := WVoter_coll.InsertOne(ctx, voter)

	fmt.Println(res)
	if err != nil {
		fmt.Println("insert to WVoter failed")
	}
	// WVoter = append(WVoter, voter)
	var x int32
	x = 0
	return &Status{Code: &x}, nil
}

type RegisteredVoter struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty"`
	Name           string               `bson:"name"`
	Group          string               `bson:"group"`
	Public_key     sodium.SignPublicKey `bson:"public_key"`
	Auth_response  []byte               `bson:"auth_response"`
	V_token        []byte               `bson:"v_token"`
	Challenge      []byte               `bson:"challenge"`
	Alive          bool                 `bson:"alive"`
	Token_End_time string               `bson:"token_end_time"`
}

type election struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Name          string             `bson:"name"`
	Groups        []string           `bson:"groups"`
	Choices       []string           `bson:"choices"`
	Already_voted []string           `bson:"already_voted"`
	End_time      string             `bson:"end_time"`
	Alive         bool               `bson:"alive"`
	Result        map[string]int32   `bson:"result"`
}
