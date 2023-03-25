package main

import (
	"OnlineVoting/voting"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

func registeredVoter() (int, error) {
	var name, group string
	var public_key []byte
	fmt.Println("Perform register voter...")
	fmt.Println("Please fill in the required information.")
	fmt.Print("voter's name:")
	fmt.Scan(&name)
	fmt.Print("voter's group:")
	fmt.Scan(&group)
	fmt.Printf("voter's public key:")
	fmt.Scan(&public_key)

	v := voting.RegisteredVoter{
		Name:       name,
		Group:      group,
		Public_key: public_key,
		Alive:      true,
	}
	voting.RVoter = append(voting.RVoter, v)
	for _, vo := range voting.RVoter {
		fmt.Println(vo)
	}
	return 0, nil
}

func unregisterVoter() (int, error) {
	var name string
	fmt.Println("Perform unregister voter...")
	fmt.Println("Please fill in the required information.")
	fmt.Print("voter's name:")
	fmt.Scan(&name)
	for i, vo := range voting.RVoter {
		if vo.Name == name {
			voting.RVoter = append(voting.RVoter[:i], voting.RVoter[i+1:]...)
		}
	}
	fmt.Printf("len: %d\n", len(voting.RVoter))
	for _, vo := range voting.RVoter {
		fmt.Println(vo)
	}
	return 0, nil
}

func help() {
	fmt.Println("r : register voter")
	fmt.Println("d : unregister voter")
	fmt.Println("e : exit")
	fmt.Println("f : finish register")
}

func main() {
	println("gRPC server tutorial in Go")
	//fmt.Printf("len of rVote: %d\n", len(voting.RVoter))
	var ip string
	fmt.Println("Which IP do you want to listen on?")
	fmt.Scan(&ip)
	var port string
	fmt.Println("Which Port do you want to listen on?")
	fmt.Scan(&port)
	ip += ":"
	ip += port
	listener, err := net.Listen("tcp", ip)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	done := make(chan string)
	voting.RegisterVotingServer(s, &voting.Server{})
	go func() {
		//fmt.Println("in the goroutine")
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		done <- "done"
	}()
	var task string
	for {
		fmt.Println("What task do you want to perform?(h for help)")
		fmt.Print("task:")
		fmt.Scan(&task)
		//fmt.Printf("the task is %s. \n", task)
		b := false
		switch task {
		case "r":
			registeredVoter()
		case "h":
			help()
		case "d":
			unregisterVoter()
		case "e":
			fmt.Println("exit")
			os.Exit(0)
		case "f":
			fmt.Println("complete registration")
			b = true
		default:
			fmt.Println("unknown task!")
		}
		time.Sleep(1 * time.Second)
		if b {
			break
		}
	}
	fmt.Println("break")
	<-done //wait the goroutine

}
