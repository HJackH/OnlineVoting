package main

import (
	"OnlineVoting/voting"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jamesruan/sodium"
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
	pub := sodium.SignPublicKey{
		Bytes: public_key,
	}

	v := voting.RegisteredVoter{
		Name:       name,
		Group:      group,
		Public_key: pub,
		Alive:      true,
	}
	voting.RVoter = append(voting.RVoter, v)
	for _, vo := range voting.RVoter {
		fmt.Println(vo)
	}
	return 0, nil
}

func register_one(in voting.RegisteredVoter) int {
	fmt.Printf("Register %s...\n", in.Name)
	var result int
	name := in.Name
	for _, vo := range voting.RVoter {
		if vo.Name == name {
			fmt.Println(" Voter with the same name already exists")
			result = 1
			return result
		}
	}
	voting.RVoter = append(voting.RVoter, in)
	result = 0
	return result
}

func register_all() {
	fmt.Println("Register all voter in WVoter")
	for _, vo := range voting.WVoter {
		//fmt.Printf("vo.allive:")
		fmt.Println(vo.Alive)
		if vo.Alive == true {
			r := register_one(vo)
			if r == 1 {
				fmt.Println("register error")
			}
		}
		//voting.WVoter = append(voting.WVoter[:i], voting.WVoter[i+1:]...)
	}
	voting.WVoter = nil
}

func see_all() {
	fmt.Println("see all registered voter")
	for _, vo := range voting.RVoter {
		fmt.Println(vo)
	}
}

func unregisterVoter() (int, error) {
	var name string
	fmt.Println("Perform unregister voter...")
	fmt.Println("Please fill in the required information.")
	fmt.Print("voter's name:")
	fmt.Scan(&name)
	find := false
	for i, vo := range voting.RVoter {
		if vo.Name == name {
			voting.RVoter = append(voting.RVoter[:i], voting.RVoter[i+1:]...)
			find = true
			break
		}
	}
	if find == false {
		fmt.Println("No voter with the name exists on the server")
		return 1, nil
	}
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
	fmt.Println("s : see all registered voter")
	fmt.Println("w : Register all voter in WVoter")
}

func main() {
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
		case "s":
			see_all()
		case "w":
			register_all()
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
