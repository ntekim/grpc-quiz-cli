package server

import (
	"context"
	"log"
	"net"
	"sync"

	pb "github.com/ntekim/grpc-cli-quiz/proto"
	"google.golang.org/grpc"
)

var questions = []*pb.Question{
	{
		Id:           1,
		QuestionDesc: "What is the capital of Canada?",
		Options:      []string{"Ottawa", "Toronto", "Vancouver", "Calgary"},
	},
	{
		Id:           2,
		QuestionDesc: "What is the capital of Australia?",
		Options:      []string{"Sydney", "Melbourne", "Brisbane", "Perth", "Canberra"},
	},
	{
		Id:           3,
		QuestionDesc: "Which of these is not a programming language?",
		Options:      []string{"Python", "Go", "C#", "Java", "HTML"},
	},
	{
		Id:           4,
		QuestionDesc: "The product of 10 and 20 is?",
		Options:      []string{"200", "2000", "20000", "200000"},
	},
	{
		Id:           5,
		QuestionDesc: "What is the capital of India?",
		Options:      []string{"New Delhi", "Mumbai", "Chennai", "Kolkata"},
	},
	{
		Id:           6,
		QuestionDesc: "What is the capital of Germany?",
		Options:      []string{"Berlin", "Hamburg", "Munich", "Frankfurt"},
	},
	{
		Id:           7,
		QuestionDesc: "Who is the president of the United States of America?",
		Options:      []string{"Joe Biden", "Donald Trump", "Barack Obama", "George W. Bush"},
	},
	{
		Id:           8,
		QuestionDesc: "Who is the CEO of Google?",
		Options:      []string{"Sundar Pichai", "Mark Zuckerberg", "Larry Page", "Sergey Brin"},
	},
	{
		Id:           9,
		QuestionDesc: "Who is the Test Quiz for?",
		Options:      []string{"Facebook", "Google", "Microsoft", "Apple", "Fast Track"},
	},
	{
		Id:           10,
		QuestionDesc: "GRPC is used for?",
		Options:      []string{"Communication", "Storage", "Network", "None of the above"},
	},
}

var correctAnswers = map[int32]string{
	1:  "Ottawa",
	2:  "Canberra",
	3:  "HTML",
	4:  "200",
	5:  "New Delhi",
	6:  "Berlin",
	7:  "Joe Biden",
	8:  "Sundar Pichai",
	9:  "Fast Track",
	10: "Communication",
}

type cliQuizService struct {
	pb.UnimplementedCLIQuizServiceServer
}

func (s *cliQuizService) GetQuestions(_ context.Context, _ *pb.NoRequestParam) (*pb.QuestionsResponsePayload, error) {
	return &pb.QuestionsResponsePayload{Questions: questions}, nil
}

func (s *cliQuizService) SubmitAnswers(_ context.Context, req *pb.AnswersRequestPayload) (*pb.ResultResponsePayload, error) {
	correctAnswersCount := int32(0)
	for _, answer := range req.GetAnswers() {
		if correctAnswers[answer.GetQuestionId()] == answer.GetAnswer() {
			correctAnswersCount++
		}
	}
	percentageMultiplier := 100
	resultPercentage := float32(correctAnswersCount) / float32(len(questions)) * float32(percentageMultiplier)
	return &pb.ResultResponsePayload{CorrectAnswerCount: correctAnswersCount, ResultPercentage: resultPercentage, UserId: req.GetUserId()}, nil
}

func NewServer(ctx context.Context, wg *sync.WaitGroup, port string) {
	lis, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Printf("Failed to listen: %v", err)
		return
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCLIQuizServiceServer(grpcServer, &cliQuizService{})
	go func() {
		<-ctx.Done()
		log.Println("Received shutdown signal, stopping gRPC server...")
		grpcServer.GracefulStop()
		log.Println("gRPC server stopped gracefully")
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Printf("Failed to serve: %v", err)
		return
	}
}
