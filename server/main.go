package main

import (
	"OnlineVoting/voting"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jamesruan/sodium"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := voting.RVoter_coll.InsertOne(ctx, v)
	fmt.Println(res)
	if err != nil {
		fmt.Println("register failed")
	}

	return 0, nil
}

func register_one(in voting.RegisteredVoter) int {
	fmt.Printf("Register %s...\n", in.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result int

	var vo voting.RegisteredVoter
	err := voting.RVoter_coll.FindOne(ctx, bson.D{{"name", in.Name}}).Decode(&vo)

	// found same name in database
	if err == nil {
		fmt.Println(" Voter with the same name already exists")
		result = 1
		return result
	}

	voting.RVoter_coll.InsertOne(ctx, in)
	result = 0
	return result
}

func register_all() {
	fmt.Println("Register all voter in WVoter")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := voting.WVoter_coll.Find(ctx, bson.M{})

	if err != nil {
		fmt.Println("list all voter error")
		fmt.Println(err)
	}

	for cursor.Next(ctx) {
		var vo voting.RegisteredVoter
		if err := cursor.Decode(&vo); err != nil {
			log.Fatal(err)
		}

		fmt.Println(vo)
		if vo.Alive == true {
			r := register_one(vo)
			if r == 1 {
				fmt.Println("register error")
			}
		}
	}

	// clear all voter in WVoter
	res, err := voting.WVoter_coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted %v documents\n", res.DeletedCount)
}

func see_all() {
	fmt.Println("see all registered voter")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := voting.RVoter_coll.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var vo voting.RegisteredVoter
		if err := cursor.Decode(&vo); err != nil {
			log.Fatal(err)
		}

		fmt.Println(vo)
	}
}

func unregisterVoter() (int, error) {
	var name string
	fmt.Println("Perform unregister voter...")
	fmt.Println("Please fill in the required information.")
	fmt.Print("voter's name:")
	fmt.Scan(&name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var vo voting.RegisteredVoter
	err := voting.RVoter_coll.FindOneAndDelete(ctx, bson.D{{"name", name}}).Decode(&vo)
	if err != nil {
		fmt.Println("No voter with the name exists on the server")
		return 1, nil
	}

	fmt.Println("delete the voter: ", vo)
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
	uri := "mongodb://mongo1:27017,mongo2:27017"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db_client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db_client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	err = db_client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB")

	voting.RVoter_coll = db_client.Database("voting").Collection("RVoter")
	voting.WVoter_coll = db_client.Database("voting").Collection("WVoter")
	voting.RElection_coll = db_client.Database("voting").Collection("RElection")

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
