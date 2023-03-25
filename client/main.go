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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

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
	fmt.Println("c : Create_Electio")
	fmt.Println("e : exit")
	fmt.Println("f : finish task and break")
	fmt.Println("v : Cast_Vote")
	fmt.Println("g : Get_result")
}

func main() {
	flag.Parse()
	fmt.Printf("name:%s\n", *name)

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	c := pb.NewVotingClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
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
			Create_Election(ctx, c)
		case "v":
			Cast_Vote(ctx, c)
		case "g":
			Get_result(ctx, c)
		default:
			fmt.Println("unknown task!")
		}
		time.Sleep(1 * time.Second)
		if b {
			break
		}
	}

}

func Cast_Vote(ctx context.Context, client pb.VotingClient) {
	fmt.Println("(CastVote)please fill the following data...")
	var election_name string
	fmt.Println("Election's name:")
	fmt.Scan(&election_name)
	var choice_name string
	fmt.Println("Choice's name:")
	fmt.Scan(&choice_name)
	r, err := client.CastVote(ctx, &pb.Vote{
		ElectionName: &election_name,
		ChoiceName:   &choice_name,
		Token:        &voting.AuthToken{Value: tokenByte},
	})
	fmt.Println("result status:")
	fmt.Println(r)
	if err != nil {
		fmt.Println("CastVote error")
		fmt.Println(err)
	}
}

func Get_result(ctx context.Context, client pb.VotingClient) {
	fmt.Println("(GetResult)please fill the following data...")
	var election_name string
	fmt.Println("Election's name:")
	fmt.Scan(&election_name)
	result, err := client.GetResult(ctx, &pb.ElectionName{Name: &election_name})
	if err != nil {
		fmt.Println("Get_result error")
		fmt.Println(err)
	} else {
		fmt.Println("election result:")
		fmt.Println(result)
	}
}

func Create_Election(ctx context.Context, client pb.VotingClient) {
	fmt.Println("(CreateElection)please fill the following data...")

	var election_name string
	fmt.Println("Election's name:")
	fmt.Scan(&election_name)

	var time_m int
	fmt.Println("How many minutes will the election be held?")
	fmt.Scan(&time_m)

	t := time.Now().Add(time.Minute * time.Duration(time_m))
	end_date := timestamppb.New(t)
	t1 := time.Unix(end_date.GetSeconds(), 0)
	fmt.Println("end time:")
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

	_, err := client.CreateElection(ctx, &pb.Election{
		Name:    &election_name,
		EndDate: end_date,
		Token:   &voting.AuthToken{Value: tokenByte},
		Groups:  groups,
		Choices: choices,
	})
	if err != nil {
		fmt.Println("create election error")
		fmt.Println(err)
	}

}
