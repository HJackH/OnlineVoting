package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"OnlineVoting/voting"
	pb "OnlineVoting/voting"

	"github.com/jamesruan/sodium"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type CRegisteredVoter struct {
	Name      string
	Group     string
	key_pair  sodium.SignKP
	sig       sodium.Signature
	V_token   []byte
	Challenge []byte
	Alive     bool
}

var CVoter []CRegisteredVoter

const (
	defaultName = "world"
)

var (
	addr      = flag.String("addr", "localhost:9000", "the address to connect to")
	name      = flag.String("name", defaultName, "Name to greet")
	group     = flag.String("group", "group0", "the group of voting")
	tokenByte = []byte("GoodLuck")
)

func GetIntPointerS(value string) *string {
	return &value
}

func help() {
	fmt.Println("c : Create_Election")
	fmt.Println("e : exit")
	fmt.Println("f : finish task and break")
	fmt.Println("v : Cast_Vote")
	fmt.Println("g : Get_result")
	fmt.Println("r : Registration")
	fmt.Println("p : PreAuth")
	fmt.Println("a : Auth")
}

func main() {
	flag.Parse()
	var ip string
	fmt.Println("Which IP do you want to connect to?(primary)")
	fmt.Scan(&ip)
	var port string
	fmt.Println("Which Port do you want to connect to?(primary)")
	fmt.Scan(&port)
	ip += ":"
	ip += port

	// Set up a connection to the server.
	conn, err := grpc.Dial(ip, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	c := pb.NewVotingClient(conn)

	fmt.Println("Which IP do you want to connect to?(secondary)")
	fmt.Scan(&ip)

	fmt.Println("Which Port do you want to connect to?(secondary)")
	fmt.Scan(&port)
	ip += ":"
	ip += port

	// Set up a connection to the server.
	conn2, err2 := grpc.Dial(ip, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err2 != nil {
		log.Fatalf("did not connect: %v", err2)
	}

	defer conn2.Close()
	c2 := pb.NewVotingClient(conn2)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*10)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	var task string
	for {
		fmt.Println("What task do you want to perform?(h for help)")
		fmt.Print("task:")
		fmt.Scan(&task)
		//fmt.Printf("the task is %s. \n", task)
		b := false
		switch task {
		case "h":
			help()
		case "e":
			fmt.Println("exit")
			os.Exit(0)
		case "f":
			fmt.Println("finish task and break")
			b = true
		case "c":
			Create_Election(ctx, c, c2)
		case "v":
			Cast_Vote(ctx, c, c2)
		case "g":
			Get_result(ctx, c, c2)
		case "r":
			Register(ctx, c, c2)
		case "p":
			pre_auth(ctx, c, c2)
		case "a":
			auth(ctx, c, c2)
		default:
			fmt.Println("unknown task!")
		}
		time.Sleep(1 * time.Second)
		if b {
			break
		}
	}

}

func pre_auth(ctx context.Context, client pb.VotingClient, client2 pb.VotingClient) {
	fmt.Println("Perform PreAuth...")
	var name string
	fmt.Println("Please fill in the required information.")
	fmt.Print("voter's name:")
	fmt.Scan(&name)
	find := false
	var index int
	for i, vo := range CVoter {
		if vo.Name == name {
			find = true
			index = i
			break
		}
	}
	if find == false {
		fmt.Println("can't find name")
		return
	}
	r, err := client.PreAuth(ctx, &pb.VoterName{
		Name: &name,
	})
	if err != nil {
		fmt.Println("preAuth error")
		fmt.Println("switch to secondary server")

		r, err = client2.PreAuth(ctx, &pb.VoterName{
			Name: &name,
		})

		if err != nil {
			fmt.Println("preAuth error with secondary server")
		}
	}
	CVoter[index].Challenge = r.Value
	//fmt.Println(string(r.Value))
	ch := sodium.Bytes(r.Value)
	sig := ch.SignDetached(CVoter[index].key_pair.SecretKey)
	CVoter[index].sig = sig
	//check

	// verify := sodium.Bytes(CVoter[index].Auth_response)
	// mess, _ := verify.SignOpen(CVoter[index].key_pair.PublicKey)
	// fmt.Printf("check... :%s\n", string(mess[:]))
}

func auth(ctx context.Context, client pb.VotingClient, client2 pb.VotingClient) {
	fmt.Println("Perform Auth...")
	var name string
	fmt.Println("Please fill in the required information.")
	fmt.Print("voter's name:")
	fmt.Scan(&name)
	find := false
	var index int
	for i, vo := range CVoter {
		if vo.Name == name {
			find = true
			index = i
			break
		}
	}
	if find == false {
		fmt.Println("can't find name")
		return
	}

	r, err := client.Auth(ctx, &pb.AuthRequest{
		Name: &pb.VoterName{
			Name: &name,
		},
		Response: &pb.Response{
			Value: CVoter[index].sig.Bytes,
		},
	})
	if err != nil {
		fmt.Println("auth error")
		fmt.Println("switch to secondary server")

		r, err = client2.Auth(ctx, &pb.AuthRequest{
			Name: &pb.VoterName{
				Name: &name,
			},
			Response: &pb.Response{
				Value: CVoter[index].sig.Bytes,
			},
		})

		if err != nil {
			fmt.Println("Auth error with secondary server")
		}
	}
	fmt.Print("return token:")
	fmt.Println(r.Value)
	CVoter[index].V_token = r.Value
}

func Register(ctx context.Context, client pb.VotingClient, client2 pb.VotingClient) {
	var name, group string
	fmt.Println("Perform register...")
	fmt.Println("Please fill in the required information.")
	fmt.Print("voter's name:")
	fmt.Scan(&name)
	fmt.Print("voter's group:")
	fmt.Scan(&group)
	myKP := sodium.MakeSignKP()
	fmt.Println(myKP.PublicKey)
	r, err := client.RegisterVoter(ctx, &pb.Voter{
		Name:      &name,
		Group:     &group,
		PublicKey: myKP.PublicKey.Bytes,
	})

	if err != nil {
		fmt.Println("Register error")
		fmt.Println(err)
		fmt.Println("switch to secondary server")

		r, err = client2.RegisterVoter(ctx, &pb.Voter{
			Name:      &name,
			Group:     &group,
			PublicKey: myKP.PublicKey.Bytes,
		})

		if err != nil {
			fmt.Println("Register error")
			fmt.Println(err)
		}
	}

	vo := CRegisteredVoter{
		Name:     name,
		Group:    group,
		key_pair: myKP,
		Alive:    true,
	}
	CVoter = append(CVoter, vo)
	fmt.Print("result status:")
	fmt.Println(r)
}

func Cast_Vote(ctx context.Context, client pb.VotingClient, client2 pb.VotingClient) {
	fmt.Println("(CastVote)please fill the following data...")
	var name string
	fmt.Print("your name:(need your token)")
	fmt.Scan(&name)
	find := false
	var token []byte
	//var index int
	for _, vo := range CVoter {
		if vo.Name == name {
			find = true
			token = vo.V_token
			break
		}
	}
	if find == false {
		fmt.Println("can't find name")
		return
	}

	if token == nil {
		fmt.Println("not have token yet")
		return
	}

	var election_name string
	fmt.Print("Election's name: ")
	fmt.Scan(&election_name)
	var choice_name string
	fmt.Print("Choice's name: ")
	fmt.Scan(&choice_name)
	r, err := client.CastVote(ctx, &pb.Vote{
		ElectionName: &election_name,
		ChoiceName:   &choice_name,
		Token: &voting.AuthToken{
			Value: token,
		},
	})

	if err != nil {
		fmt.Println("CastVote error")
		fmt.Println(err)
		fmt.Println("switch to secondary server")

		r, err = client2.CastVote(ctx, &pb.Vote{
			ElectionName: &election_name,
			ChoiceName:   &choice_name,
			Token: &voting.AuthToken{
				Value: token,
			},
		})

		if err != nil {
			fmt.Println("CastVote error")
			fmt.Println(err)
		}
	}

	fmt.Println("result status:")
	fmt.Println(r)
}

func Get_result(ctx context.Context, client pb.VotingClient, client2 pb.VotingClient) {
	fmt.Println("(GetResult)please fill the following data...")
	var name string
	fmt.Print("your name:(need your token)")
	fmt.Scan(&name)
	find := false
	var token []byte
	//var index int
	for _, vo := range CVoter {
		if vo.Name == name {
			find = true
			token = vo.V_token
			break
		}
	}
	if find == false {
		fmt.Println("can't find name")
		return
	}

	if token == nil {
		fmt.Println("not have token yet")
		return
	}
	var election_name string
	fmt.Println("Election's name:")
	fmt.Scan(&election_name)
	result, err := client.GetResult(ctx, &pb.ElectionName{Name: &election_name})
	if err != nil {
		fmt.Println("Get_result error")
		fmt.Println(err)
		fmt.Println("switch to secondary server")

		result, err = client2.GetResult(ctx, &pb.ElectionName{Name: &election_name})
		if err != nil {
			fmt.Println("Get_result error")
			fmt.Println(err)
		}
	}
	fmt.Println("election result:")
	fmt.Println(result)
}

func Create_Election(ctx context.Context, client pb.VotingClient, client2 pb.VotingClient) {
	fmt.Println("(CreateElection)please fill the following data...")
	var name string
	fmt.Print("your name:(need your token)")
	fmt.Scan(&name)
	find := false
	var token []byte
	//var index int
	for _, vo := range CVoter {
		if vo.Name == name {
			find = true
			token = vo.V_token
			break
		}
	}
	if find == false {
		fmt.Println("can't find name")
		return
	}

	if token == nil {
		fmt.Println("not have token yet")
		return
	}

	var election_name string
	fmt.Print("Election's name: ")
	fmt.Scan(&election_name)

	var time_m int
	fmt.Print("How many minutes will the election be held? ")
	fmt.Scan(&time_m)

	t := time.Now().Add(time.Minute * time.Duration(time_m))
	end_date := timestamppb.New(t)
	t1 := time.Unix(end_date.GetSeconds(), 0)
	fmt.Print("end time: ")
	fmt.Println(t1)
	var groups []string
	fmt.Println("Which groups can vote?(e for end)")
	for {
		var temp string
		fmt.Scan(&temp)
		if temp == "e" {
			break
		}
		groups = append(groups, temp)
	}
	var choices []string
	fmt.Println("what voting options are there?(e for end)")
	for {
		var temp string
		fmt.Scan(&temp)
		if temp == "e" {
			break
		}
		choices = append(choices, temp)
	}

	r, err := client.CreateElection(ctx, &pb.Election{
		Name:    &election_name,
		EndDate: end_date,
		Token:   &voting.AuthToken{Value: token},
		Groups:  groups,
		Choices: choices,
	})

	if err != nil {
		fmt.Println("Create_Election error")
		fmt.Println(err)
		fmt.Println("switch to secondary server")

		r, err = client2.CreateElection(ctx, &pb.Election{
			Name:    &election_name,
			EndDate: end_date,
			Token:   &voting.AuthToken{Value: token},
			Groups:  groups,
			Choices: choices,
		})

		if err != nil {
			fmt.Println("Create_Election error")
			fmt.Println(err)
		}
	}
	fmt.Print("Create Election result:")
	te := *r.Code
	fmt.Println(te)

}
